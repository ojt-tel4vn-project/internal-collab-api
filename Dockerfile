# ─────────────────────────────────────────────
# Stage 1: Builder
# ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install build tools
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Download dependencies first (leverages Docker layer cache)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build a statically-linked Linux binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2: Production image
# ─────────────────────────────────────────────
FROM alpine:3.19

# Security: run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# CA certs for HTTPS calls (Brevo, Supabase, etc.)
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/server .

# Scripts (seed, clear) – optional, copy if needed at runtime
COPY --from=builder /app/scripts ./scripts

# Run as non-root
USER appuser

# Port exposed by the API
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./server"]
