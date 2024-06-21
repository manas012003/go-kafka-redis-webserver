# Go Kafka Redis Webserver

A simple web server built with Go that interacts with Kafka and Redis. It demonstrates how to handle POST requests to send data to Kafka and GET requests to retrieve data from Redis.

## Features

- **POST /post**: Sends data to a Kafka topic.
- **GET /get**: Retrieves the latest data from Redis.

## Architecture

- **Go**: The main programming language used.
- **Kafka**: Used for message queuing.
- **Redis**: Used for data storage.
- **Docker**: Used to containerize Kafka, Zookeeper, and Redis.

## Prerequisites

- Docker and Docker Compose
- Go (Golang)

## Setup

1. Clone the repository:
    ```bash
    git clone https://github.com/yourusername/go-kafka-redis-webserver.git
    cd go-kafka-redis-webserver
    ```

2. Set up Docker containers:
    ```bash
    docker-compose up -d
    ```

3. Install Go dependencies:
    ```bash
    go get github.com/IBM/sarama
    go get github.com/go-redis/redis/v8
    go get github.com/gorilla/mux
    ```

4. Run the Go web server:
    ```bash
    go run main.go
    ```

## Endpoints

### POST /post

- Description: Sends data to a Kafka topic.
- Request body:
    ```json
    {
        "key": "value"
    }
    ```
- Example:
    ```bash
    curl -X POST -d '{"key":"value"}' -H "Content-Type: application/json" http://localhost:8080/post
    ```

### GET /get

- Description: Retrieves the latest data from Redis.
- Example:
    ```bash
    curl http://localhost:8080/get
    ```
