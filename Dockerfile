# syntax=docker/dockerfile:1

# Stage 1: frontend build
FROM node:24-alpine@sha256:2bdb65ed1dab192432bc31c95f94155ca5ad7fc1392fb7eb7526ab682fa5bf14 AS frontend
WORKDIR /app/frontend
RUN npm install -g pnpm@11
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

# Stage 2: Go build + Go-based security tools
FROM golang:1.27-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder
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
    go install github.com/ossf/scorecard/v5@v5.6.0 && \
    go install github.com/zricethezav/gitleaks/v8@v8.30.1 && \
    go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 && \
    go install github.com/google/osv-scanner/v2/cmd/osv-scanner@v2.7.1

# Stage 3: download arch-aware binary tools
# Wolfi uses glibc (not musl), so all pre-built glibc binaries are compatible.
FROM --platform=$BUILDPLATFORM cgr.dev/chainguard/wolfi-base@sha256:28c39be5a5e1e9b85081c10740e8495f6518a865b23bc6a486637c4b21e74b43 AS bintools
ARG TARGETARCH
ARG TRIVY_VERSION=0.76.2
ARG HELM_VERSION=4.3.2
ARG ZIZMOR_VERSION=1.25.2
ARG KUBELINTER_VERSION=0.8.3
ARG TRUFFLEHOG_VERSION=3.99.9
RUN apk add --no-cache curl ca-certificates && mkdir -p /export/bin

# trivy — SHA256 verified
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "ARM64" || echo "64bit") && \
    HASH=$([ "$TARGETARCH" = "arm64" ] \
      && echo "ab9d51f86e6a5fb936f605faff6b593ac51bc378fff4d3e8b9d58358b9c367c3" \
      || echo "bc0fc726e017b8927a62b2d3531d7a84ee01769daf5c99ed866dced54089fdbf") && \
    curl -sfL -o /tmp/trivy.tar.gz \
      "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-${ARCH}.tar.gz" && \
    echo "${HASH}  /tmp/trivy.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/trivy.tar.gz -C /export/bin trivy && \
    rm /tmp/trivy.tar.gz

# helm — SHA256 verified
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    HASH=$([ "$TARGETARCH" = "arm64" ] \
      && echo "df31cb61680ef728f82c770e5280dd565617bee6b039d37e85c677288be4bbb0" \
      || echo "d1c0278d46beab79582d7e50f5b3dc206944ba8e577cde618513fd07daa53983") && \
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
      && echo "778b1913196dc9b9234e8fd366f64387efeb08c01930a1d2110dcb2ffd07e68b" \
      || echo "12050095c8063bf4d34f9f6cbce6ba9fa8d0deea9d503d0570364feba093658e") && \
    curl -sfL -o /tmp/trufflehog.tar.gz \
      "https://github.com/trufflesecurity/trufflehog/releases/download/v${TRUFFLEHOG_VERSION}/trufflehog_${TRUFFLEHOG_VERSION}_linux_${ARCH}.tar.gz" && \
    echo "${HASH}  /tmp/trufflehog.tar.gz" | sha256sum -c - && \
    tar -xzf /tmp/trufflehog.tar.gz -C /export/bin trufflehog && \
    rm /tmp/trufflehog.tar.gz

# Stage 4: runtime — Wolfi base for near-zero OS CVE surface
FROM cgr.dev/chainguard/wolfi-base@sha256:28c39be5a5e1e9b85081c10740e8495f6518a865b23bc6a486637c4b21e74b43

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
