# Use the latest Go version (1.24)
FROM golang:1.24 AS builder

# Set the working directory
WORKDIR /app

# Copy go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN GOOS=linux GOARCH=amd64 go build -o main main.go

# Use a lightweight base image for production
FROM debian:bullseye

# Set working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port (Docker metadata)
EXPOSE 8080

# Set default PORT environment variable
ENV PORT=8080

# Run the binary
CMD ["./main"]


