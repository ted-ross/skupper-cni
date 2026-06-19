# Multi-stage Dockerfile for Skupper CNI Plugin

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the plugin
RUN make build-linux

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates iptables

# Copy the plugin binary from builder
COPY --from=builder /build/bin/skupper-cni /opt/cni/bin/skupper-cni

# Copy example configurations
COPY --from=builder /build/examples /etc/cni/examples

# Set permissions
RUN chmod +x /opt/cni/bin/skupper-cni

# Add metadata
LABEL maintainer="Skupper Project" \
      description="Skupper CNI Plugin for TUN interface management" \
      version="1.0.0"

# The plugin will be invoked by the container runtime, not run directly
CMD ["/opt/cni/bin/skupper-cni"]