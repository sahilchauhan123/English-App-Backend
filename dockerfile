# Use Go 1.23 base image
FROM golang:1.23-alpine AS builder

# Set Go env vars
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create app directory
WORKDIR /app

# Copy go mod files and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy full source code
COPY . .

# Build your app
RUN go build -o main ./cmd/main.go

# Use lightweight runtime container
FROM alpine:latest

# Set working dir in final container
WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the app
CMD ["./main"]
