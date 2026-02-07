# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./
COPY static/ ./static/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o taskmate .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/taskmate .
COPY --from=builder /app/static ./static

# Create volume for persistent data
VOLUME ["/app/data"]

# Expose port
EXPOSE 8080

# Run the application
CMD ["./taskmate"]
