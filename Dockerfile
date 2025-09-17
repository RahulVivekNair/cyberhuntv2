# ---- STAGE 1: Build ----
# Use the official Go image as a builder image.
# Using alpine for a smaller build image. Match the Go version from your go.mod if possible.
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies first
# This leverages Docker's layer caching, so dependencies are only re-downloaded if they change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application, creating a statically linked binary.
# -ldflags="-w -s" strips debug information, reducing the binary size.
# CGO_ENABLED=0 is important for creating a static binary that runs on minimal images like Alpine.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /cyberhunt ./cmd/main.go

# ---- STAGE 2: Final ----
# Use a minimal, secure base image for the final container
FROM alpine:latest

# It's good practice to install ca-certificates for any potential HTTPS outbound requests
# and tzdata for correct time handling.
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /cyberhunt .

# Copy the templates directory, which is required by the application at runtime
COPY --from=builder /app/templates ./templates

# The application creates a database in a 'data' directory.
# We need to create this directory and ensure the non-root user has permissions to write to it.
RUN mkdir data && chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the port the application will run on
EXPOSE 8080

# The command to run the application
# Using ENTRYPOINT makes the container behave like an executable
ENTRYPOINT ["/cyberhunt"]

# Default command-line arguments can be provided with CMD
CMD ["-addr", ":8080"]