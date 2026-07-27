package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse" 
)

func main() {
	// 1. Initialize Redis Connection
	Databse.ConnectRedis()

	time.Sleep(1 * time.Second)
	if Databse.Rdb == nil {
		fmt.Println("Exiting: Redis is not connected.")
		return
	}

	testCacheAside()
	testDistributedLock()
	testRateLimiter()
}

func testCacheAside() {
	fmt.Println("\n--- Testing Cache-Aside Pattern ---")
	
	mockDB := map[string]Databse.Product{
		"101": {ID: "101", Name: "MacBook Pro", Price: 2500.00},
	}

	fmt.Println("Attempt 1 (Should be Cache Miss):")
	prod1, err := Databse.GetProduct(mockDB, "101")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Result: %+v\n", prod1)
	}

	fmt.Println("Attempt 2 (Should be Cache Hit):")
	prod2, err := Databse.GetProduct(mockDB, "101")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Result: %+v\n", prod2)
	}
}

func testDistributedLock() {
	fmt.Println("\n--- Testing Distributed Lock ---")
	
	var wg sync.WaitGroup
	wg.Add(2)

	productID := "limited_edition_sneaker"

	// User 1
	go func() {
		defer wg.Done()
		err := Databse.ProcessOrder(productID, "user_1")
		if err != nil {
			fmt.Printf("User 1 failed: %v\n", err)
		} else {
			fmt.Println("User 1 successfully processed the order!")
		}
	}()

	// User 2
	go func() {
		defer wg.Done()
		err := Databse.ProcessOrder(productID, "user_2")
		if err != nil {
			fmt.Printf("User 2 failed: %v\n", err)
		} else {
			fmt.Println("User 2 successfully processed the order!")
		}
	}()

	wg.Wait()
}

func testRateLimiter() {
	fmt.Println("\n--- Testing Sliding-Window Rate Limiter ---")
	
	userID := "user_xyz"
	limit := int64(3)
	window := 10 * time.Second

	for i := 1; i <= 5; i++ {
		allowed, err := Databse.AllowRequest(userID, limit, window)
		if err != nil {
			fmt.Println("Error checking rate limit:", err)
			continue
		}

		if allowed {
			fmt.Printf("Request %d: ALLOWED (Processed)\n", i)
		} else {
			fmt.Printf("Request %d: DENIED (Rate limit exceeded)\n", i)
		}
		
		time.Sleep(100 * time.Millisecond) 
	}
}