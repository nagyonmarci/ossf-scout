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
    go install sigs.k8s.io/kube-linter/cmd/kube-linter@v0.8.3

# Stage 3: download binary tools (trivy, helm, zizmor)
FROM alpine:3.22@sha256:310c62b5e7ca5b08167e4384c68db0fd2905dd9c7493756d356e893909057601 AS bintools
RUN apk add --no-cache curl tar

RUN curl -sfL https://github.com/aquasecurity/trivy/releases/download/v0.71.0/trivy_0.71.0_Linux-64bit.tar.gz \
    | tar xz -C /usr/local/bin trivy

RUN curl -sfL https://get.helm.sh/helm-v4.2.0-linux-amd64.tar.gz \
    | tar xz --strip-components=1 -C /usr/local/bin linux-amd64/helm

RUN curl -sfL https://github.com/zizmorcore/zizmor/releases/download/v1.25.2/zizmor-x86_64-unknown-linux-gnu.tar.gz \
    | tar xz -C /usr/local/bin zizmor

# Stage 4: runtime — all tools bundled
FROM alpine:3.22@sha256:310c62b5e7ca5b08167e4384c68db0fd2905dd9c7493756d356e893909057601
RUN apk add --no-cache git python3 py3-pip
RUN pip3 install --no-cache-dir checkov==3.2.532

WORKDIR /app
COPY --from=builder /app/ossf-scout .
COPY --from=builder /go/bin/scorecard    /usr/local/bin/scorecard
COPY --from=builder /go/bin/gitleaks     /usr/local/bin/gitleaks
COPY --from=builder /go/bin/actionlint   /usr/local/bin/actionlint
COPY --from=builder /go/bin/kube-linter  /usr/local/bin/kube-linter
COPY --from=bintools /usr/local/bin/trivy  /usr/local/bin/trivy
COPY --from=bintools /usr/local/bin/helm   /usr/local/bin/helm
COPY --from=bintools /usr/local/bin/zizmor /usr/local/bin/zizmor

EXPOSE 7878
CMD ["./ossf-scout", "-serve"]
