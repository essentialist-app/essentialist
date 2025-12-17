# Build Stage
FROM golang:1.24 AS builder

# Install Fyne dependencies
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Fyne CLI
RUN go install fyne.io/fyne/v2/cmd/fyne@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build WASM Frontend
WORKDIR /app/cmd/essentialist
RUN fyne package -os web

# Build Go Server
WORKDIR /app
RUN go build -o server ./cmd/server

# Runtime Stage
FROM debian:bookworm-slim

WORKDIR /app

# Copy server binary
COPY --from=builder /app/server /app/server

# Copy static frontend files
COPY --from=builder /app/cmd/essentialist/wasm /app/cmd/essentialist/wasm

# Expose port
EXPOSE 8080

# Create data directory
RUN mkdir -p /data

# Default command
CMD ["./server", "-data", "/data", "-static", "/app/cmd/essentialist/wasm"]
