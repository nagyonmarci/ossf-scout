# Stage 1: frontend build
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Go build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN go build -o ossf-scout .
RUN go install github.com/ossf/scorecard/v5@latest

# Stage 3: minimal runtime
FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/ossf-scout .
COPY --from=builder /root/go/bin/scorecard /usr/local/bin/scorecard
EXPOSE 7878
CMD ["./ossf-scout", "-serve"]
