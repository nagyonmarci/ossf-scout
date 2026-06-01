# syntax=docker/dockerfile:1

# Stage 1: frontend build
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Go build
FROM golang:1.25-alpine AS builder
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
    go install github.com/ossf/scorecard/v5@latest

# Stage 3: minimal runtime
FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/ossf-scout .
COPY --from=builder /go/bin/scorecard /usr/local/bin/scorecard
EXPOSE 7878
CMD ["./ossf-scout", "-serve"]
