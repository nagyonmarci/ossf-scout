# syntax=docker/dockerfile:1

# Stage 1: frontend build
FROM node:22-alpine@sha256:968df39aedcea65eeb078fb336ed7191baf48f972b4479711397108be0966920 AS frontend
WORKDIR /app/frontend
RUN npm install -g pnpm@11
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

# Stage 2: Go build + Go-based security tools
FROM golang:1.26-alpine@sha256:f23e8b227fb4493eabe03bede4d5a32d04092da71962f1fb79b5f7d1e6c2a17f AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /go/bin/ossf-scout .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/ossf/scorecard/v5@v5.5.0 && \
    go install github.com/zricethezav/gitleaks/v8@v8.21.2 && \
    go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7 && \
    go install github.com/google/osv-scanner/cmd/osv-scanner@v1.9.2

# Stage 3: download arch-aware binary tools
# Wolfi uses glibc (not musl), so all pre-built glibc binaries are compatible.
FROM --platform=$BUILDPLATFORM cgr.dev/chainguard/wolfi-base@sha256:b78bb982194828b6c9c214230bf34d51944e2102ea8468f01ac21e5f99328efd AS bintools
ARG TARGETARCH
ARG TRIVY_VERSION=0.71.0
ARG HELM_VERSION=4.2.0
ARG ZIZMOR_VERSION=1.25.2
ARG KUBELINTER_VERSION=0.8.3
ARG TRUFFLEHOG_VERSION=3.95.5
RUN apk add --no-cache curl ca-certificates && mkdir -p /export/bin

# trivy — SHA256 verified
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "ARM64" || echo "64bit") && \
    HASH=$([ "$TARGETARCH" = "arm64" ] \
      && echo "2561be394a3199c911f82fced606cbc05e1cb23eb6ce1da6935540adb76f4252" \
      || echo "30a3d22b23f88c233f1658f562fb477cae3b3e8b4761109d515b7698daf85814") && \
    curl -sfL -o /tmp/trivy.tar.gz \
      "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-${ARCH}.tar.gz" && \
    echo "${HASH}  /tmp/trivy.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/trivy.tar.gz -C /export/bin trivy && \
    rm /tmp/trivy.tar.gz

# helm — SHA256 verified
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    HASH=$([ "$TARGETARCH" = "arm64" ] \
      && echo "1f8de130dfbd04de64978e7b852a7a547be1404956a366608276d2520b678670" \
      || echo "97dbeb971be4ac4b27e3839976d9564c0fb35c6f3b1da89dd1e292d236af4096") && \
    curl -sfL -o /tmp/helm.tar.gz \
      "https://get.helm.sh/helm-v${HELM_VERSION}-linux-${ARCH}.tar.gz" && \
    echo "${HASH}  /tmp/helm.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/helm.tar.gz -C /export/bin --strip-components=1 "linux-${ARCH}/helm" && \
    rm /tmp/helm.tar.gz

# zizmor — no official checksum published for this release
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "aarch64" || echo "x86_64") && \
    curl -sfL "https://github.com/zizmorcore/zizmor/releases/download/v${ZIZMOR_VERSION}/zizmor-${ARCH}-unknown-linux-gnu.tar.gz" \
    | tar xz -C /export/bin zizmor

# kube-linter — no official checksum; project uses Sigstore for verification
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "_arm64" || echo "") && \
    curl -sfL -o /export/bin/kube-linter \
      "https://github.com/stackrox/kube-linter/releases/download/v${KUBELINTER_VERSION}/kube-linter-linux${ARCH}" && \
    chmod +x /export/bin/kube-linter

# trufflehog — SHA256 verified
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    HASH=$([ "$TARGETARCH" = "arm64" ] \
      && echo "bb876c4e5a84fa4fdbda4fc24143ed2d12eac32cfd3f7e41c79cbd7d33607b4a" \
      || echo "8d151a19465973bec226be5992a2a11b053f4ab92c77861f642089892ae9aa58") && \
    curl -sfL -o /tmp/trufflehog.tar.gz \
      "https://github.com/trufflesecurity/trufflehog/releases/download/v${TRUFFLEHOG_VERSION}/trufflehog_${TRUFFLEHOG_VERSION}_linux_${ARCH}.tar.gz" && \
    echo "${HASH}  /tmp/trufflehog.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/trufflehog.tar.gz -C /export/bin trufflehog && \
    rm /tmp/trufflehog.tar.gz

# Stage 4: runtime — Wolfi base for near-zero OS CVE surface
FROM cgr.dev/chainguard/wolfi-base@sha256:b78bb982194828b6c9c214230bf34d51944e2102ea8468f01ac21e5f99328efd

LABEL org.opencontainers.image.title="ossf-scout" \
      org.opencontainers.image.description="Discover GitHub repositories where security practices are weakest" \
      org.opencontainers.image.source="https://github.com/nagyonmarci/ossf-scout" \
      org.opencontainers.image.licenses="Unlicense"

RUN apk add --no-cache git nodejs npm ca-certificates python3 py3-pip dumb-init curl \
    && pip3 install --no-cache-dir semgrep==1.165.0 \
    && apk del py3-pip
RUN npm install -g pnpm@11

RUN addgroup -S -g 10001 scout && adduser -S -u 10001 -G scout -H -D scout \
    && mkdir -p /home/scout && chown scout:scout /home/scout

ENV TRIVY_CACHE_DIR=/tmp/trivy \
    TMPDIR=/tmp
VOLUME ["/tmp"]

WORKDIR /app
COPY --from=builder  /go/bin/      /usr/local/bin/
COPY --from=bintools /export/bin/  /usr/local/bin/
RUN chown scout:scout /app

USER 10001:10001

EXPOSE 7878
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:7878/ || exit 1
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["ossf-scout", "-serve"]
