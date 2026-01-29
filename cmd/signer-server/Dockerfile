# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
# Copy the source code
COPY . .
# Build the binary
RUN go build -o signer-server ./cmd/signer-server/main.go

# Run stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/signer-server .

# Expose port
EXPOSE 8080

# Run
CMD ["./signer-server"]
