# Use the latest Go version (1.24)
FROM golang:1.24 AS builder

# Set the working directory
WORKDIR /app

# Copy go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Cross-compile for Linux (from Mac)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

# Use Ubuntu 22.04 LTS as the base image for production
FROM ubuntu:22.04 

# Set working directory
WORKDIR /root/

# Install necessary dependencies (if needed)
RUN apt update && apt install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port (Docker metadata)
EXPOSE 8080

# Set default PORT environment variable
ENV PORT=8080

# Run the binary
CMD ["./main"]
