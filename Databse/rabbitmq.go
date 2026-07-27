package Databse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitConn *amqp.Connection
var RabbitChannel *amqp.Channel

// ConnectRabbitMQ establishes connection and provisions advanced queues
func ConnectRabbitMQ() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	var err error
	RabbitConn, err = amqp.Dial(rabbitURL)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return
	}

	RabbitChannel, err = RabbitConn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %v", err)
		return
	}

	setupAdvancedQueues(RabbitChannel)
	fmt.Println("Successfully connected to RabbitMQ and provisioned queues!")
}

// setupAdvancedQueues configures the Main Queue and a Dead Letter Queue (DLQ)
func setupAdvancedQueues(ch *amqp.Channel) {
	// 1. Declare the Dead Letter Exchange (DLX)
	err := ch.ExchangeDeclare("dlx_exchange", "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare DLX: %v", err)
	}

	// 2. Declare the Dead Letter Queue (DLQ)
	_, err = ch.QueueDeclare("invoice_dead_letter_queue", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare DLQ: %v", err)
	}

	// 3. Bind the DLQ to the DLX
	err = ch.QueueBind("invoice_dead_letter_queue", "dlx_routing_key", "dlx_exchange", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind DLQ: %v", err)
	}

	// 4. Declare the Main Queue and configure it to send failed messages to the DLX
	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "dlx_routing_key",
		// Optional: "x-message-ttl": int32(60000), // Messages expire and go to DLQ after 60s
	}

	_, err = ch.QueueDeclare(
		"invoice_queue", // name
		true,            // durable (survives broker restart)
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		args,            // arguments (links to DLX)
	)
	if err != nil {
		log.Fatalf("Failed to declare main queue: %v", err)
	}
}

// PublishInvoiceEvent publishes a message to RabbitMQ safely
func PublishInvoiceEvent(orderID string, userEmail string) error {
	if RabbitChannel == nil {
		return fmt.Errorf("rabbitmq channel is not initialized")
	}

	// Context with timeout prevents the API from hanging if RabbitMQ is slow
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := map[string]string{
		"order_id": orderID,
		"email":    userEmail,
	}
	body, _ := json.Marshal(payload)

	err := RabbitChannel.PublishWithContext(ctx,
		"",              // default exchange
		"invoice_queue", // routing key (queue name)
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Ensure message is saved to disk
			ContentType:  "application/json",
			Body:         body,
		})

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	fmt.Printf(" [x] Sent invoice request for Order: %s\n", orderID)
	return nil
}