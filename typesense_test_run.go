package main

import (
	"context" // Add this to your imports
	"fmt"
	"time"

	"github.com/typesense/typesense-go/typesense/api"
	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse"
)

func main() {
	fmt.Println("--- Starting Typesense Advanced Test ---")

	// 1. Initialize Connection
	Databse.ConnectTypesense()
	if Databse.TypesenseClient == nil {
		fmt.Println("Exiting: Typesense is not connected.")
		return
	}

	// NEW: Delete the old conflicting collection to start fresh
	fmt.Println("Wiping old collection...")
	Databse.TypesenseClient.Collection("products").Delete(context.Background())

	// 2. Create Schema
	err := Databse.CreateAdvancedProductSchema()
	if err != nil {
		fmt.Printf("Schema Error: %v\n", err)
		return
	}

	// 3. Index Sample Products
	fmt.Println("\n--- Indexing Products ---")
	products := []map[string]interface{}{
		{
			"id":          "1",
			"title":       "Apple iPhone 15 Pro",
			"description": "The latest titanium smartphone with advanced camera.",
			"brand":       "Apple",
			"category":    "Smartphones",
			"price":       999.00,
		},
		{
			"id":          "2",
			"title":       "Samsung Galaxy S24 Ultra",
			"description": "AI-powered Android smartphone with stylus.",
			"brand":       "Samsung",
			"category":    "Smartphones",
			"price":       1199.00,
		},
		{
			"id":          "3",
			"title":       "Sony WH-1000XM5 Headphones",
			"description": "Industry-leading noise canceling over-ear headphones.",
			"brand":       "Sony",
			"category":    "Audio",
			"price":       398.00,
		},
	}

	for _, p := range products {
		err := Databse.IndexProduct(p)
		if err != nil {
			fmt.Printf("Failed to index %s: %v\n", p["title"], err)
		} else {
			fmt.Printf("Indexed: %s\n", p["title"])
		}
	}

	// Wait a brief moment for Typesense to flush data to memory
	time.Sleep(1 * time.Second)

	// 4. Test Typo-Tolerant Search
	fmt.Println("\n--- Test 1: Typo-Tolerant Search ('ipohne') ---")
	res1, err := Databse.SearchProducts("ipohne", "")
	if err != nil {
		fmt.Println("Search Error:", err)
	} else {
		printSearchResults(res1)
	}

	// 5. Test Faceted/Filtered Search
	fmt.Println("\n--- Test 2: Filtered Search (Category: 'Audio') ---")
	res2, err := Databse.SearchProducts("", "Audio")
	if err != nil {
		fmt.Println("Search Error:", err)
	} else {
		printSearchResults(res2)
	}
}

// printSearchResults is a helper to format and print the Typesense response
func printSearchResults(res *api.SearchResult) {
	if res.Hits == nil || len(*res.Hits) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("Found %d results in %d ms:\n", *res.Found, *res.SearchTimeMs)
	for _, hit := range *res.Hits {
		doc := *hit.Document
		
		title := doc["title"].(string)
		price := doc["price"].(float64)
		brand := doc["brand"].(string)
		
		fmt.Printf(" - %s ($%.2f) [Brand: %s]\n", title, price, brand)
	}
}