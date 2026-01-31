# Build stage
FROM golang:1.25.3-alpine AS builder

# Install dependencies for CGO (required by some packages)
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s" \
    -o main cmd/main.go

# Build migration binary
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s" \
    -o migration migrations/cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migration .

# Copy migrations folder
COPY --from=builder /app/migrations ./migrations

# Note: .env file will be mounted as volume or copied at runtime
# Do not copy .env here to keep secrets secure

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
