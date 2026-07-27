package main

import (
	"fmt"
	"time"

	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse"
	"github.com/AlinAlexMyladoor/ecommerce-backend/workers"
)

func main() {
	fmt.Println("--- Starting RabbitMQ Advanced Test ---")

	// 1. Initialize RabbitMQ Connection and provision the Queues/Exchanges
	Databse.ConnectRabbitMQ()
	time.Sleep(1 * time.Second) // Give it a moment to connect

	if Databse.RabbitChannel == nil {
		fmt.Println("Exiting: RabbitMQ is not connected.")
		return
	}

	// 2. Start the background worker (Consumer) in a Goroutine
	// This will listen for messages without blocking the rest of our script
	go workers.StartInvoiceWorker()

	// Give the worker a split second to set up its QoS and begin listening
	time.Sleep(1 * time.Second)

	fmt.Println("\n--- Publishing Events ---")

	// 3. Publish a Valid Order (Should Succeed and Ack)
	err := Databse.PublishInvoiceEvent("ORD-001", "alin@example.com")
	if err != nil {
		fmt.Printf("Failed to publish: %v\n", err)
	}

	// 4. Publish an Invalid Order (Should Fail, Nack, and go to Dead Letter Queue)
	err = Databse.PublishInvoiceEvent("ORD-002", "bad@email.com")
	if err != nil {
		fmt.Printf("Failed to publish: %v\n", err)
	}

	// 5. Publish another Valid Order to ensure the worker didn't crash
	err = Databse.PublishInvoiceEvent("ORD-003", "alex@example.com")
	if err != nil {
		fmt.Printf("Failed to publish: %v\n", err)
	}

	// 6. Keep the main function alive long enough for the worker to process all 3 messages.
	// Since our worker has a simulated 2-second delay per message, we wait 10 seconds.
	fmt.Println("\n[Main] Waiting 10 seconds for worker to process queue...")
	time.Sleep(10 * time.Second)
	fmt.Println("[Main] Test Complete. Exiting.")
}