# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy necessary files
COPY --from=builder /app/configs ./configs

# Expose port 8080 (Gin's default port)
EXPOSE 8080

# Run the application
CMD ["./main"]
