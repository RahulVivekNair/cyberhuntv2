# Build stage
FROM golang:1.25-alpine AS builder

# Install necessary tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source files
COPY . .

# Build the Go binary
RUN go build -o cyberhunt ./cmd/main.go

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/cyberhunt .

# Copy templates
COPY --from=builder /app/templates ./templates

# Create data directory for SQLite
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Command to run the server
CMD ["./cyberhunt", "-addr", ":8080"]