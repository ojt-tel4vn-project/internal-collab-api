# Build Stage
FROM golang:alpine AS builder

# Install build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# We build a Linux binary named 'main' (not main.exe) because Docker containers run Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Production Stage
FROM alpine:latest

WORKDIR /app

# Install certificates for external APIs (like Brevo)
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy environment file example (optional, better to set via docker-compose)
# COPY .env.example .env

# Expose the port (matches default in .env)
EXPOSE 8080

# Run the binary
CMD ["./main"]
