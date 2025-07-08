# Start from the Go image
FROM golang:1.24.1 as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
RUN go build -o app

# Start a new stage for a smaller final image
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/app .

# Copy templates
COPY templates ./templates

# Expose port
EXPOSE 80

# Run the app
CMD ["./app"]

