package Databse

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

// ------------------------------------------------------------------------
// 1. Connection Setup
// ------------------------------------------------------------------------

func ConnectRedis() {
	// Read REDIS_ADDR from environment variable. 
	// Fall back to "localhost:6379" when running locally outside Docker.
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // No password set by default
		DB:       0,  // Use default DB
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		fmt.Printf("Failed to connect to Redis at %s: %v\n", redisAddr, err)
		return
	}
	fmt.Printf("Successfully connected to Redis at %s!\n", redisAddr)
}

// ------------------------------------------------------------------------
// 2. Cache-Aside Pattern (Product Caching)
// ------------------------------------------------------------------------

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// GetProduct implements the Cache-Aside pattern
func GetProduct(db map[string]Product, productID string) (*Product, error) {
	if Rdb == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}

	cacheKey := fmt.Sprintf("product:%s", productID)

	// 1. Try fetching from Redis first
	val, err := Rdb.Get(Ctx, cacheKey).Result()
	if err == nil {
		// Cache Hit
		var p Product
		json.Unmarshal([]byte(val), &p)
		return &p, nil
	} else if err != redis.Nil {
		// Real Redis connection error
		return nil, err
	}

	// 2. Cache Miss: Fetch from primary database (simulated here with a map)
	p, exists := db[productID]
	if !exists {
		return nil, fmt.Errorf("product not found in db")
	}

	// 3. Populate Cache with a TTL (1 hour)
	pBytes, _ := json.Marshal(p)
	err = Rdb.Set(Ctx, cacheKey, pBytes, time.Hour).Err()
	if err != nil {
		fmt.Printf("Failed to set cache for %s: %v\n", productID, err)
	}

	return &p, nil
}

// UpdateProduct updates the DB and invalidates the cache
func UpdateProduct(db map[string]Product, p Product) error {
	db[p.ID] = p

	if Rdb != nil {
		cacheKey := fmt.Sprintf("product:%s", p.ID)
		return Rdb.Del(Ctx, cacheKey).Err()
	}
	return nil
}

// ------------------------------------------------------------------------
// 3. Distributed Locking (Order Processing)
// ------------------------------------------------------------------------

func ProcessOrder(productID string, userID string) error {
	if Rdb == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	lockKey := fmt.Sprintf("lock:product:%s", productID)

	// 1. Acquire Lock using SetNX (Set if Not eXists) with a 5-second TTL
	acquired, err := Rdb.SetNX(Ctx, lockKey, userID, 5*time.Second).Result()
	if err != nil {
		return err
	}

	if !acquired {
		return fmt.Errorf("could not acquire lock, another transaction is processing product %s", productID)
	}

	// 2. Ensure lock is released even if a panic occurs
	defer Rdb.Del(Ctx, lockKey)

	// 3. Critical Section
	fmt.Println("Lock acquired. Checking inventory and deducting for product:", productID)
	time.Sleep(1 * time.Second) // Simulating DB transaction

	return nil
}

// ------------------------------------------------------------------------
// 4. Sliding-Window Rate Limiter
// ------------------------------------------------------------------------

func AllowRequest(userID string, limit int64, window time.Duration) (bool, error) {
	if Rdb == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}

	key := fmt.Sprintf("ratelimit:%s", userID)
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	pipe := Rdb.TxPipeline()

	// 1. Remove outdated request timestamps outside the time window
	pipe.ZRemRangeByScore(Ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 2. Add current request timestamp
	pipe.ZAdd(Ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// 3. Count total requests in the current window
	countCmd := pipe.ZCard(Ctx, key)

	// 4. Reset TTL on the bucket
	pipe.Expire(Ctx, key, window)

	_, err := pipe.Exec(Ctx)
	if err != nil {
		return false, err
	}

	return countCmd.Val() <= limit, nil
}