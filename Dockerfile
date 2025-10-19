# Multi-stage build for Go API server
FROM golang:1.25-alpine AS builder

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build the binary with CGO enabled (required for go-sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o hamqrzdb-api .

# Final stage - minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/hamqrzdb-api .

# Copy the index.html file
COPY html/index.html /app/index.html

# Expose port
EXPOSE 8080

# Set default environment variables
ENV DB_PATH=/data/hamqrzdb.sqlite
ENV PORT=8080

# Run the binary
CMD ["./hamqrzdb-api"]
