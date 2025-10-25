# Multi-stage build for Go binaries
FROM golang:1.25-alpine AS builder

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY main.go ./
COPY cmd/ ./cmd/

# Build the API binary with CGO enabled (required for go-sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o hamqrzdb-api .

# Build the process binary (now includes location processing)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o hamqrzdb-process ./cmd/process/main.go

# Final stage - minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs wget

WORKDIR /app

# Create data directory for the SQLite database (default DB_PATH)
RUN mkdir -p /data

# Copy binaries from builder
COPY --from=builder /build/hamqrzdb-api .
COPY --from=builder /build/hamqrzdb-process .

# Copy the index.html file
COPY html/index.html /app/index.html

# Copy entrypoint script
COPY docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

# Expose port
EXPOSE 8080

# Set default environment variables
ENV DB_PATH=/data/hamqrzdb.sqlite
ENV PORT=8080

# Declare /data as a mount point for optional host persistence
VOLUME ["/data"]

# Use entrypoint script to auto-create database if needed
ENTRYPOINT ["/app/docker-entrypoint.sh"]
