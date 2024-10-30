#!/bin/bash

# Set the Docker image tag
IMAGE_TAG="gosumer-rabbitmq"

# Print the image tag being used
echo "Building Docker image with tag: $IMAGE_TAG"

# Build the Docker image
docker build -t $IMAGE_TAG .

# Print completion message
echo "Docker image $IMAGE_TAG built successfully"
