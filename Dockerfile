# Use the official Golang image as the base image
FROM golang:1.22.2 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# Use a minimal base image for the final container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /app/main .

# Expose the port the application runs on
EXPOSE 8080

ENV RABBIT_MQ_CONNECTION_STRING amqp://default_user_IlR2D-NZ6d5U0tWa34m:9ZH8eHWc3bmpCxrbboHArn3qiUP1VieK@localhost:5672/
ENV RABBIT_MQ_QUEUE_NAME shortener
ENV MONGODB_COLLECTION_NAME documents
ENV MONGODB_CONNECTION_STRING mongodb://admin:admin@localhost:27017
ENV MONGODB_DATABASE_NAME documents


# Command to run the application
CMD ["./main"]
