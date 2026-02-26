# syntax=docker/dockerfile:1
# ^ enables BuildKit features (faster, better caching)

# ────────────────────────────────────────────────
# Builder stage
# ────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install git (sometimes needed for private modules)
RUN apk add --no-cache git

WORKDIR /app

# Download dependencies first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# run tests / lint
RUN go vet ./...
RUN go test ./... -v

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /app/main ./cmd/server/main.go

# ────────────────────────────────────────────────
# Final stage — minimal runtime
# ────────────────────────────────────────────────
FROM alpine:3.21

# Install ca-certificates (for HTTPS calls, e.g. to RDS)
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main /app/main

# copy .env 
COPY .env .

# Switch to non-root user
USER appuser

# Expose port (can be overridden with -p or env)
EXPOSE 880

# Healthcheck (ECS/ALB loves this)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:880/health || exit 1

# Run the binary
CMD ["/app/main"]