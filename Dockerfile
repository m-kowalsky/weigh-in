# Start from the Go image
FROM golang:1.24.1 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app

# Final minimal image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates curl && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/app .
COPY templates ./templates
COPY static ./static
COPY ./sql/schema ./sql/schema/

RUN mkdir -p /app/data && chown -R nobody:nogroup /app/data

# Expose correct port
EXPOSE 8080

CMD ["./app"]
