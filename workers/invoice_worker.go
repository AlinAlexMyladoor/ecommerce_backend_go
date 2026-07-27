package workers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse"
)

// StartInvoiceWorker boots up the background listener
func StartInvoiceWorker() {
	if Databse.RabbitChannel == nil {
		log.Fatal("Cannot start worker: RabbitMQ channel is nil")
	}

	// 1. Set Quality of Service (QoS)
	// PrefetchCount = 5 means this worker will only process 5 messages at a time.
	// It won't receive more until it explicitly acknowledges the previous ones.
	err := Databse.RabbitChannel.Qos(5, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// 2. Consume messages from the main queue
	// autoAck is set to FALSE for manual acknowledgments
	msgs, err := Databse.RabbitChannel.Consume(
		"invoice_queue", // queue
		"",              // consumer
		false,           // auto-ack (FALSE = MUST MANUALLY ACK)
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// 3. Listen to the channel forever
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf("\n[Worker] Received a message: %s\n", d.Body)

			var payload map[string]string
			json.Unmarshal(d.Body, &payload)

			// Simulate processing time (e.g., generating PDF, connecting to SMTP)
			time.Sleep(2 * time.Second)

			// Simulate a random failure logic
			if payload["email"] == "bad@email.com" {
				fmt.Printf("[Worker] Failed to process invoice for %s. Sending to DLQ.\n", payload["order_id"])
				
				// Nack (Negative Acknowledgement)
				// multiple=false, requeue=false -> This forces it into the Dead Letter Queue
				d.Nack(false, false)
			} else {
				fmt.Printf("[Worker] Successfully generated and sent invoice to %s\n", payload["email"])
				
				// Ack (Positive Acknowledgement) -> Removes it from RabbitMQ completely
				d.Ack(false)
			}
		}
	}()

	log.Printf(" [*] Invoice Worker is waiting for messages. To exit press CTRL+C")
	<-forever
}