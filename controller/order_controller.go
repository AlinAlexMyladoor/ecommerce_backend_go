// controllers/order_controller.go
// transaction-oneki full ayitu nadakum alenki onum nadakila
package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse"

	"github.com/AlinAlexMyladoor/ecommerce-backend/services"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// OrderItemInput represents one item in the posted order
// oronu ayitu store akum
type OrderItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// OrderCreateInput is the body for creating an order
// dive means oro prouct check cheyan,lisintinte ulil uladine check cehyum
type OrderCreateInput struct { //list
	Items []OrderItemInput `json:"items" binding:"required,dive"`
	// optionally you can accept shipping info, etc.
}

// POST /orders
func CreateOrder(c *gin.Context) {
	// get user_id from the middleware (assuming middleware sets it)
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		JSONError(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	userID := uint(userIDRaw.(float64))

	var input OrderCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		JSONError(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	service := services.OrderService{
		DB: Databse.DB,
	}

	var items []services.OrderItemInput

	for _, it := range input.Items {
		items = append(items, services.OrderItemInput{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	err := service.CreateOrder(userID, items)
	if err != nil {
		// if your error text includes "out of stock", convert to 400
		if strings.Contains(err.Error(), "out of stock") {
			JSONError(c, http.StatusBadRequest, err.Error())
			return
		}
		JSONError(c, http.StatusInternalServerError, "failed to create order: "+err.Error())
		return
	}

	// ==========================================
	// 🐰 RABBITMQ PRODUCER LOGIC STARTS HERE
	// ==========================================

	// 1. Create the Ticket (The Message)
	// We use formatting to write the user's ID into a JSON string
	messageBody := fmt.Sprintf(`{"user_id": %d, "status": "pending_invoice"}`, userID)

	// 2. Set up the Stopwatch (Context)
	// We give RabbitMQ exactly 5 seconds to accept the ticket.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 3. Stick the Ticket on the Wheel (Publish)
	if Databse.RabbitChannel == nil {
		fmt.Println("Warning: RabbitMQ channel is not ready; skipping invoice task publish")
	} else {
		rabbitErr := Databse.RabbitChannel.PublishWithContext(ctx,
			"",              // exchange (We leave this blank to use the default router)
			"order_created", // routing key (This MUST match the queue name we made earlier)
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(messageBody), // The actual ticket data
			})

		// 4. Check if the Ticket Wheel is broken
		if rabbitErr != nil {
			// CRITICAL: We do NOT crash the app or send an error to the user here!
			// The order is already saved safely in PostgreSQL. We just log the error.
			fmt.Println("Warning: Failed to publish invoice task to RabbitMQ:", rabbitErr)
		} else {
			fmt.Println("PUBLISHED MESSAGE: Ticket sent to RabbitMQ for User", userID)
		}
	}

	// ==========================================
	// 🐰 RABBITMQ PRODUCER LOGIC ENDS HERE
	// ==========================================

	// 5. Instantly tell the user they are good to go
	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully! Generating invoice in the background.",
	})
}

// Helper function to send consistent JSON error responses
func JSONError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": message,
	})
}
