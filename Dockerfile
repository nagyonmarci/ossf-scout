# syntax=docker/dockerfile:1

# Stage 1: frontend build
FROM node:22-alpine@sha256:968df39aedcea65eeb078fb336ed7191baf48f972b4479711397108be0966920 AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY frontend/ ./
RUN npm run build

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
    go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7

# Stage 3: download arch-aware binary tools
FROM --platform=$BUILDPLATFORM debian:12-slim@sha256:0104b334637a5f19aa9c983a91b54c89887c0984081f2068983107a6f6c21eeb AS bintools
ARG TARGETARCH
RUN apt-get update && apt-get install -y --no-install-recommends curl tar ca-certificates \
    && rm -rf /var/lib/apt/lists/*

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

# Stage 4: runtime — all tools bundled, debian for glibc compatibility
FROM debian:12-slim@sha256:0104b334637a5f19aa9c983a91b54c89887c0984081f2068983107a6f6c21eeb
RUN apt-get update && apt-get install -y --no-install-recommends \
    git python3 python3-pip nodejs npm ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN pip3 install --no-cache-dir --break-system-packages checkov==3.2.532
RUN npm install -g pnpm@8

WORKDIR /app
COPY --from=builder /app/ossf-scout .
COPY --from=builder /go/bin/scorecard    /usr/local/bin/scorecard
COPY --from=builder /go/bin/gitleaks     /usr/local/bin/gitleaks
COPY --from=builder /go/bin/actionlint   /usr/local/bin/actionlint
COPY --from=bintools /usr/local/bin/trivy      /usr/local/bin/trivy
COPY --from=bintools /usr/local/bin/helm       /usr/local/bin/helm
COPY --from=bintools /usr/local/bin/zizmor     /usr/local/bin/zizmor
COPY --from=bintools /usr/local/bin/kube-linter  /usr/local/bin/kube-linter
COPY --from=bintools /usr/local/bin/trufflehog   /usr/local/bin/trufflehog

EXPOSE 7878
CMD ["./ossf-scout", "-serve"]
