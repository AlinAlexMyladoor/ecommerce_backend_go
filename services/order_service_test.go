package services

import (
	"testing"

	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOrderTestDB() *gorm.DB {

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	db.AutoMigrate(
		&models.Product{},
		&models.Order{},
		&models.Order_item{},
	)

	return db
}

func TestCreateOrderSuccess(t *testing.T) {

	db := setupOrderTestDB()

	// Seed product
	product := models.Product{
		Name:  "Laptop",
		Price: 1000,
		Stock: 10,
	}
	db.Create(&product)

	service := OrderService{DB: db}

	items := []OrderItemInput{
		{ProductID: product.ID, Quantity: 2},
	}

	err := service.CreateOrder(1, items)

	if err != nil {
		t.Fatalf("expected success but got error: %v", err)
	}

	// Verify stock deduction
	var updated models.Product
	db.First(&updated, product.ID)

	if updated.Stock != 8 {
		t.Errorf("expected stock 8, got %d", updated.Stock)
	}

	// Verify order created
	var orderCount int64
	db.Model(&models.Order{}).Count(&orderCount)

	if orderCount != 1 {
		t.Errorf("expected 1 order, got %d", orderCount)
	}
}

func TestCreateOrderRollbackOnStockFailure(t *testing.T) {

	db := setupOrderTestDB()

	product := models.Product{
		Name:  "Phone",
		Price: 500,
		Stock: 1,
	}
	db.Create(&product)

	service := OrderService{DB: db}

	items := []OrderItemInput{
		{ProductID: product.ID, Quantity: 5},
	}

	err := service.CreateOrder(1, items)

	if err == nil {
		t.Fatalf("expected stock error but got nil")
	}

	// Verify no order created
	var count int64
	db.Model(&models.Order{}).Count(&count)

	if count != 0 {
		t.Errorf("expected rollback but order exists")
	}

	// Verify stock unchanged
	var updated models.Product
	db.First(&updated, product.ID)

	if updated.Stock != 1 {
		t.Errorf("stock should remain 1 after rollback")
	}
}

func TestOrderTotalCalculation(t *testing.T) {

	db := setupOrderTestDB()

	p1 := models.Product{Name: "A", Price: 100, Stock: 10}
	p2 := models.Product{Name: "B", Price: 50, Stock: 10}

	db.Create(&p1)
	db.Create(&p2)

	service := OrderService{DB: db}

	items := []OrderItemInput{
		{ProductID: p1.ID, Quantity: 2}, // 200
		{ProductID: p2.ID, Quantity: 1}, // 50
	}

	service.CreateOrder(1, items)

	var order models.Order
	db.First(&order)

	if order.TotalAmount != 250 {
		t.Errorf("expected total 250, got %.2f", order.TotalAmount)
	}
}