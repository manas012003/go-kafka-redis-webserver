package main

import (
    "context" // Standard library for managing context with deadlines, cancellations, etc.
    "encoding/json" // Standard library for encoding and decoding JSON
    "log" // Standard library for logging
    "net/http" // Standard library for HTTP client and server implementation
    "github.com/IBM/sarama" // Kafka client library
    "github.com/go-redis/redis/v8" // Redis client library
    "github.com/gorilla/mux" // HTTP request router
    "time" // Standard library for time manipulation
)

// Global variables for Kafka producer, consumer, Redis client, and context
var (
    kafkaProducer sarama.SyncProducer // Synchronous Kafka producer instance
    kafkaConsumer sarama.Consumer // Kafka consumer instance
    redisClient   *redis.Client // Redis client instance
    ctx           = context.Background() // Context used for Redis operations
)

// Function to initialize Kafka producer and consumer
func initKafka() {
    // Create a new Kafka configuration
    config := sarama.NewConfig()
    config.Producer.Return.Successes = true // Ensure the producer returns successes

    var err error
    // Create a synchronous Kafka producer
    kafkaProducer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
    if err != nil {
        log.Fatal("Error creating Kafka producer:", err)
    }

    // Create a Kafka consumer
    kafkaConsumer, err = sarama.NewConsumer([]string{"localhost:9092"}, nil)
    if err != nil {
        log.Fatal("Error creating Kafka consumer:", err)
    }
}

// Function to initialize Redis client
func initRedis() {
    // Create a new Redis client
    redisClient = redis.NewClient(&redis.Options{
        Addr: "localhost:6379", // Address of the Redis server
    })
    // Ping the Redis server to check connectivity
    if _, err := redisClient.Ping(ctx).Result(); err != nil {
        log.Fatal("Error connecting to Redis:", err)
    }
}

// Handler for POST requests
// Parameters:
// - w http.ResponseWriter: Used to construct an HTTP response
// - r *http.Request: Represents the HTTP request received
func postDataHandler(w http.ResponseWriter, r *http.Request) {
    // Declare a map to hold the incoming JSON data
    var data map[string]interface{}
    // Decode the JSON request body into the data map
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest) // Respond with a 400 Bad Request error if decoding fails
        return
    }

    // Start a new goroutine to handle sending data to Kafka
    go func(data map[string]interface{}) {
        // Marshal the data map into a JSON byte slice
        dataBytes, _ := json.Marshal(data)
        // Create a new Kafka message with the topic "data-topic" and the JSON data as the value
        msg := &sarama.ProducerMessage{
            Topic: "data-topic",
            Value: sarama.ByteEncoder(dataBytes),
        }
        // Send the message to Kafka
        _, _, err := kafkaProducer.SendMessage(msg)
        if err != nil {
            log.Println("Failed to send message to Kafka:", err) // Log an error if the message fails to send
        }
    }(data) // Pass the data map as an argument to the goroutine

    // Respond with a 200 OK status
    w.WriteHeader(http.StatusOK)
}

// Handler for GET requests
// Parameters:
// - w http.ResponseWriter: Used to construct an HTTP response
// - r *http.Request: Represents the HTTP request received
func getDataHandler(w http.ResponseWriter, r *http.Request) {
    // Start a new goroutine to handle consuming data from Kafka and storing it in Redis
    go func() {
        // Consume messages from the "data-topic" partition 0, starting from the oldest offset
        partitionConsumer, err := kafkaConsumer.ConsumePartition("data-topic", 0, sarama.OffsetOldest)
        if err != nil {
            log.Println("Failed to start consumer:", err) // Log an error if the consumer fails to start
            return
        }
        defer partitionConsumer.Close() // Ensure the consumer is closed when the function exits

        // Loop to consume messages
        for msg := range partitionConsumer.Messages() {
            // Store the latest message value in Redis with the key "latest-data"
            err := redisClient.Set(ctx, "latest-data", msg.Value, 0).Err()
            if err != nil {
                log.Println("Failed to write to Redis:", err) // Log an error if writing to Redis fails
                return
            }
            break // Break the loop after storing the first message
        }
    }() // No arguments passed to this goroutine

    // Sleep for 2 seconds to give the goroutine time to complete
    time.Sleep(2 * time.Second)
    
    // Retrieve the latest data from Redis
    data, err := redisClient.Get(ctx, "latest-data").Result()
    if err != nil {
        http.Error(w, "Data not found in Redis", http.StatusNotFound) // Respond with a 404 Not Found error if the data is not found in Redis
        return
    }

    // Set the Content-Type header to application/json
    w.Header().Set("Content-Type", "application/json")
    
    // Write the data retrieved from Redis as the response
    w.Write([]byte(data))
}

// Main function - entry point of the application
func main() {
    initKafka() // Initialize Kafka producer and consumer
    initRedis() // Initialize Redis client

    // Create a new router
    r := mux.NewRouter()
    // Register the POST handler for the /post endpoint
    r.HandleFunc("/post", postDataHandler).Methods("POST")
    // Register the GET handler for the /get endpoint
    r.HandleFunc("/get", getDataHandler).Methods("GET")

    log.Println("Server started at :8080") // Log the server start message
    http.ListenAndServe(":8080", r) // Start the HTTP server on port 8080
}
