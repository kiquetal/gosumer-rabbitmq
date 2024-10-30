package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"sync"
	"time"
)

type ShortCode struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Host      string    `json:"host"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Creator   string    `json:"creator"`
}

var toMongo bool = false
var messageCount int
var messageMux sync.Mutex
var limiter *rate.Limiter
var batchSize = 100
var batchChan = make(chan ShortCode, batchSize)
var mongoClient *mongo.Client
var mongo_databse string = ""
var mongo_collection string = ""

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	fmt.Println("Hello, World!")
	rabbit_mq := os.Getenv("RABBIT_MQ_CONNECTION_STRING")
	queue_name := os.Getenv("RABBIT_MQ_QUEUE_NAME")
	mongo_uri := os.Getenv("MONGO_CONNECTION_STRING")
	mongo_databse = os.Getenv("MONGODB_DATABASE_NAME")
	mongo_collection = os.Getenv("MONGODB_COLLECTION_NAME")
	fmt.Printf("Mongo URI: %s\n", mongo_uri)
	fmt.Printf("Mongo Database: %s\n", mongo_databse)
	fmt.Printf("Mongo Collection: %s\n", mongo_collection)
	fmt.Printf("Queue Name: %s\n", queue_name)
	fmt.Println("RabbitMQ Connection String: ", rabbit_mq)
	conn, err := amqp.Dial(rabbit_mq)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ: ", err)
		return
	}

	limiter = rate.NewLimiter(rate.Limit(30), 20)
	/*
		mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongo_uri))
		if err != nil {

			fmt.Println("Failed to connect to MongoDB: ", err)
			return
		}

	*/

	defer mongoClient.Disconnect(context.Background())

	defer conn.Close()
	go batchInsert()
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to create channel: ", err)
		return
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		queue_name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println("Failed to declare queue: ", err)
		return
	}
	msg, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println("Failed to register consumer: ", err)
		return
	}
	var wg sync.WaitGroup
	var numConsumers = 15
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go consume(msg, &wg)
	}
	http.HandleFunc("/", handleFunction)
	go http.ListenAndServe(":8080", nil)

	for {
		time.Sleep(1 * time.Second)
	}
}

func batchInsert() {
	var batch []interface{}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case shortCode := <-batchChan:
			batch = append(batch, shortCode)
			if len(batch) >= batchSize {
				insertBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				insertBatch(batch)
				batch = nil
			}
		}
	}
}

func insertBatch(batch []interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	collection := mongoClient.Database(mongo_databse).Collection(mongo_collection)
	if _, err := collection.InsertMany(ctx, batch); err != nil {
		fmt.Println("Failed to insert batch: ", err)
		return err
	}
	fmt.Println("Inserted batch of ", len(batch), " documents")
	return nil
}

func handleFunction(w http.ResponseWriter, r *http.Request) {
	messageMux.Lock()
	fmt.Fprintf(w, "Total messages received: %d", messageCount)
	messageMux.Unlock()
}

func consume(msg <-chan amqp.Delivery, wg *sync.WaitGroup) {
	defer wg.Done()
	for d := range msg {
		if err := limiter.Wait(context.Background()); err != nil {
			fmt.Println("!!!!![Rate limit exceeded]!!!!!!!!!!!")
			d.Nack(false, true)
			continue
		}

		messageMux.Lock()
		messageCount++
		messageMux.Unlock()
		fmt.Println("Received message: ", string(d.Body))
		var shortCode ShortCode
		if err := json.Unmarshal(d.Body, &shortCode); err != nil {
			fmt.Printf("failed to unmarshal JSON: %v", err)
		}
		fmt.Println("ShortCode: ", shortCode.Code)

		if toMongo {
			if err := writeToMongo(d); err != nil {
				fmt.Println("Failed to write to MongoDB: ", err)
				d.Nack(false, true)
				continue
			}
		}

		d.Ack(false)
	}
}

func writeToMongo(d amqp.Delivery) error {

	var shortCode ShortCode
	if err := json.Unmarshal(d.Body, &shortCode); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	batchChan <- shortCode

	return nil
}