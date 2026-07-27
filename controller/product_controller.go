package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse" // Your specific DB package
	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/typesense/typesense-go/typesense/api" // Added for Typesense Search
)

// GET /products
func GetProducts(c *gin.Context) {
	db := Databse.DB

	// 1. Parse Query Parameters FIRST
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	sort := c.DefaultQuery("sort", "created_desc")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// 2. Generate a DYNAMIC Cache Key
	cacheKey := fmt.Sprintf("products:page=%d:limit=%d:sort=%s:search=%s", page, limit, sort, search)

	// 3. Check Redis First
	val, err := Databse.Rdb.Get(Databse.Ctx, cacheKey).Result()
	if err == nil {
		fmt.Println("⚡ CACHE HIT:", cacheKey)
		// We stored the entire JSON response string, so we can send it directly!
		c.Data(http.StatusOK, "application/json", []byte(val))
		return
	}

	// 4. CACHE MISS: Build the PostgreSQL Query
	fmt.Println("🐌 CACHE MISS:", cacheKey)
	query := db.Model(&models.Product{})

	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}

	switch sort {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	case "name_asc":
		query = query.Order("name ASC")
	case "name_desc":
		query = query.Order("name DESC")
	default:
		query = query.Order("created_at DESC")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count products"})
		return
	}

	var products []models.Product
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	// 5. Construct the final Payload (Data + Pagination Meta)
	response := gin.H{
		"ok":   true,
		"data": products,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	}

	// 6. Save the ENTIRE payload to Redis (Cache for 5 minutes)
	responseJSON, _ := json.Marshal(response)
	Databse.Rdb.Set(Databse.Ctx, cacheKey, responseJSON, 5*time.Minute)

	// 7. Return to user
	c.JSON(http.StatusOK, response)
}

// GET /products/:id
func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := Databse.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

// POST /products (Admin Only)
func CreateProduct(c *gin.Context) {
	var newProduct models.Product

	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := Databse.DB.Create(&newProduct)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// ==========================================
	// 🔎 TYPESENSE SYNC: Add to Search Engine
	// ==========================================
	go Databse.SyncProductToIndex(newProduct.ID, newProduct.Name, newProduct.Category, newProduct.Price)

	c.JSON(http.StatusCreated, newProduct)
}

// PUT /products/:id (Admin Only)
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	// Check if product exists
	if err := Databse.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Bind new data
	var updatedData models.Product
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update
	Databse.DB.Model(&product).Updates(updatedData)

	// ==========================================
	// 🔎 TYPESENSE SYNC: Update Search Engine
	// ==========================================
	go Databse.SyncProductToIndex(product.ID, product.Name, product.Category, product.Price)

	c.JSON(http.StatusOK, product)
}

// DELETE /products/:id (Admin Only)
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := Databse.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	Databse.DB.Delete(&product)

	// ==========================================
	// 🔎 TYPESENSE SYNC: Remove from Search Engine
	// ==========================================
	go Databse.RemoveProductFromIndex(product.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

// GET /search?q=... (Public Search)
func SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide a search term using ?q="})
		return
	}

	searchParameters := &api.SearchCollectionParams{
		Q:       query,
		QueryBy: "name,description", // Typesense will look for typos in both fields
	}

	result, err := Databse.TsClient.Collection("products").Documents().Search(Databse.Ctx, searchParameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search engine failed"})
		return
	}

	// Format the raw Typesense response to just return the documents
	var hits []interface{}
	if result.Hits != nil {
		for _, hit := range *result.Hits {
			hits = append(hits, *hit.Document)
		}
	}

	fmt.Printf("🔎 SEARCH: Found %d results for '%s'\n", len(hits), query)
	c.JSON(http.StatusOK, gin.H{"results": hits})
}