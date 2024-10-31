# Gosumer Rabbitmq



## Getting started

### Prerequisites

Environment variables:

```bash
    RABBIT_MQ_CONNECTION_STRING=amqp://user:pass@localhost:5672/
    RABBIT_MQ_QUEUE_NAME=shortener
    MONGODB_COLLECTION_NAME=documents
    MONGODB_CONNECTION_STRING=mongodb://admin:admin@localhost:27017
    MONGODB_DATABASE_NAME=documents
```

### Install the dependencies

```bash
    go mod tidy
```


### Running the application

```bash
    go run cmd/main.go
```

### Building the application

```bash
    go build cmd/main.go
```

