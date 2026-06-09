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
FROM golang:1.25-alpine@sha256:c05ba4b73604069d376c4f41346b05374335b5ca0c46fb6dfede5a59f5196931 AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o ossf-scout .
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
RUN apk add --no-cache curl ca-certificates && mkdir -p /usr/local/bin

# trivy
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "ARM64" || echo "64bit") && \
    curl -sfL "https://github.com/aquasecurity/trivy/releases/download/v0.71.0/trivy_0.71.0_Linux-${ARCH}.tar.gz" \
    | tar xz -C /usr/local/bin trivy

# helm
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    curl -sfL "https://get.helm.sh/helm-v4.2.0-linux-${ARCH}.tar.gz" \
    | tar xz --strip-components=1 -C /usr/local/bin "linux-${ARCH}/helm"

# zizmor
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "aarch64" || echo "x86_64") && \
    curl -sfL "https://github.com/zizmorcore/zizmor/releases/download/v1.25.2/zizmor-${ARCH}-unknown-linux-gnu.tar.gz" \
    | tar xz -C /usr/local/bin zizmor

# kube-linter
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "_arm64" || echo "") && \
    curl -sfL -o /usr/local/bin/kube-linter \
    "https://github.com/stackrox/kube-linter/releases/download/v0.8.3/kube-linter-linux${ARCH}" && \
    chmod +x /usr/local/bin/kube-linter

# trufflehog
RUN ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    curl -sfL "https://github.com/trufflesecurity/trufflehog/releases/download/v3.95.5/trufflehog_3.95.5_linux_${ARCH}.tar.gz" \
    | tar xz -C /usr/local/bin trufflehog

# Stage 4: runtime — Wolfi base for near-zero OS CVE surface
FROM cgr.dev/chainguard/wolfi-base@sha256:b78bb982194828b6c9c214230bf34d51944e2102ea8468f01ac21e5f99328efd
RUN apk add --no-cache git nodejs npm ca-certificates python3 py3-pip \
    && pip3 install --no-cache-dir semgrep==1.165.0 \
    && apk del py3-pip
RUN npm install -g pnpm@11

RUN addgroup -S -g 10001 scout && adduser -S -u 10001 -G scout -H -D scout \
    && mkdir -p /home/scout && chown scout:scout /home/scout

WORKDIR /app
COPY --from=builder /app/ossf-scout .
COPY --from=builder /go/bin/scorecard      /usr/local/bin/scorecard
COPY --from=builder /go/bin/gitleaks       /usr/local/bin/gitleaks
COPY --from=builder /go/bin/actionlint     /usr/local/bin/actionlint
COPY --from=builder /go/bin/osv-scanner    /usr/local/bin/osv-scanner
COPY --from=bintools /usr/local/bin/trivy      /usr/local/bin/trivy
COPY --from=bintools /usr/local/bin/helm       /usr/local/bin/helm
COPY --from=bintools /usr/local/bin/zizmor     /usr/local/bin/zizmor
COPY --from=bintools /usr/local/bin/kube-linter  /usr/local/bin/kube-linter
COPY --from=bintools /usr/local/bin/trufflehog   /usr/local/bin/trufflehog

RUN chown scout:scout /app

USER 10001:10001

EXPOSE 7878
CMD ["./ossf-scout", "-serve"]
