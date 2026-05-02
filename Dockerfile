# Stage 1: Build the binary
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
# -ldflags="-s -w" reduces binary size by removing debug info
# CGO_ENABLED=0 ensures the binary is statically linked
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o text2sql .

# Stage 2: Final minimal image
FROM alpine:latest

# Install CA certificates for secure communication with LLM APIs
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/text2sql .

# Expose the application port
EXPOSE 3000

# Run the binary
CMD ["./text2sql"]
