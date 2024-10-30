# Use the official Golang image as the base image
FROM golang:1.22.3 AS builder

# Enable CGO and set target OS and architecture
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Install necessary dependencies for CGO
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    libc6-dev

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main cmd/main.go

# Use a minimal base image for the final container
FROM scratch

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose the port the application runs on
EXPOSE 8080
ENV RABBIT_MQ_CONNECTION_STRING replace
ENV RABBIT_MQ_QUEUE_NAME replace
ENV MONGODB_COLLECTION_NAME replace
ENV MONGODB_CONNECTION_STRING replace
ENV MONGODB_DATABASE_NAME replace


# Command to run the application
CMD ["./main"]
