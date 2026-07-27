package Databse

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
)

var TypesenseClient *typesense.Client

// 1. Connect and Initialize
func ConnectTypesense() {
	// Load the .env file so os.Getenv can read your actual API key
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file. Ensure it exists in your root directory.")
	}

	typesenseURL := os.Getenv("TYPESENSE_URL")
	if typesenseURL == "" {
		typesenseURL = "http://localhost:8108"
	}

	apiKey := os.Getenv("TYPESENSE_API_KEY")
	if apiKey == "" {
		// If it's STILL empty, ensure this fallback matches the key in your docker-compose.yml exactly.
		apiKey = "super-secret-key" 
	}

	TypesenseClient = typesense.NewClient(
		typesense.WithServer(typesenseURL),
		typesense.WithAPIKey(apiKey),
	)

	// Health Check
	_, err = TypesenseClient.Health(context.Background(), 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to connect to Typesense: %v\n", err)
		return
	}

	fmt.Println("Successfully connected to Typesense!")
}

// 2. Create an Advanced Schema
func CreateAdvancedProductSchema() error {
	ctx := context.Background()
	collectionName := "products"

	// Check if collection already exists
	_, err := TypesenseClient.Collection(collectionName).Retrieve(ctx)
	if err == nil {
		fmt.Println("Typesense collection already exists.")
		return nil
	}

	schema := &api.CollectionSchema{
		Name: collectionName,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "title", Type: "string"},
			{Name: "description", Type: "string"},
			{Name: "brand", Type: "string", Facet: pointer.True()},
			{Name: "category", Type: "string", Facet: pointer.True()},
			{Name: "price", Type: "float", Facet: pointer.True()},
			{Name: "embedding", Type: "float[]", NumDim: pointer.Int(384), Optional: pointer.True()},
		},
		DefaultSortingField: pointer.String("price"),
	}

	_, err = TypesenseClient.Collections().Create(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	fmt.Println("Advanced product schema created successfully!")
	return nil
}

// 3. Index Product
func IndexProduct(product map[string]interface{}) error {
	ctx := context.Background()

	_, err := TypesenseClient.Collection("products").Documents().Upsert(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to index product: %w", err)
	}

	return nil
}

// 4. Perform Search
func SearchProducts(query string, categoryFilter string) (*api.SearchResult, error) {
	ctx := context.Background()

	filterBy := ""
	if categoryFilter != "" {
		filterBy = fmt.Sprintf("category:=%s", categoryFilter)
	}

	// Your SDK expects NumTypos as *string
	numTypos := "2"

	searchParameters := &api.SearchCollectionParams{
		Q:        query,
		QueryBy:  "title,description",
		FilterBy: pointer.String(filterBy),
		FacetBy:  pointer.String("brand,category,price"),
		NumTypos: &numTypos,
		SortBy:   pointer.String("_text_match:desc,price:asc"),
	}

	result, err := TypesenseClient.Collection("products").Documents().Search(ctx, searchParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return result, nil
}