# Stage 1: frontend build
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Go build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN go build -o ossf-scout .

# Stage 3: minimal runtime
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/ossf-scout .
EXPOSE 7878
ENTRYPOINT ["./ossf-scout", "-serve"]
