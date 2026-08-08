# 🛒 E-Commerce Backend — Go

> A production-oriented e-commerce REST API built with **Go, Gin, PostgreSQL, Redis, RabbitMQ, and Typesense**, designed with caching, asynchronous processing, distributed locking, rate limiting, JWT authentication, and transactional order processing.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.11-00ADD8)](https://gin-gonic.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-4169E1?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?logo=redis)](https://redis.io/)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-Messaging-FF6600?logo=rabbitmq)](https://www.rabbitmq.com/)
[![Typesense](https://img.shields.io/badge/Typesense-Search-FF6B35)](https://typesense.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](https://www.docker.com/)

---

## 📌 Overview

This project is a backend REST API for an e-commerce platform, built with **Go and Gin**.

It provides user authentication, product management, shopping cart operations, order processing, inventory management, search, caching, and asynchronous invoice processing.

The architecture combines **PostgreSQL for transactional data, Redis for caching and distributed coordination, RabbitMQ for asynchronous messaging, and Typesense for fast product search**.

---

## ✨ Features

### 🔐 Authentication & Authorization

* User registration and login
* JWT-based authentication
* bcrypt password hashing
* Protected routes
* Role-based access control
* Admin-only product management

### 📦 Product Management

* Create, update and delete products
* Product retrieval
* Pagination
* Search and filtering
* Price/name sorting
* Inventory tracking

### 🛒 Cart & Orders

* User-specific shopping carts
* Add products to cart
* Create multi-item orders
* Stock validation
* Automatic inventory deduction
* Order total calculation
* Transactional order processing

### ⚡ Performance & Scalability

* Redis cache-aside pattern
* Redis distributed locking
* Sliding-window rate limiting
* Typesense product search
* RabbitMQ asynchronous processing
* Dead Letter Queue (DLQ)

---

## 🏗️ Architecture

```text id="5f2x8h"
                         ┌──────────────────┐
                         │      Client      │
                         └────────┬─────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │    Gin Router    │
                         │     REST API     │
                         └────────┬─────────┘
                                  │
                         ┌────────▼────────┐
                         │   Middleware    │
                         │ JWT + RBAC      │
                         └────────┬────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │   Controllers   │
                         └────────┬─────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │    Services     │
                         │ Business Logic  │
                         └────┬────┬────┬──┘
                              │    │    │
                ┌─────────────┘    │    └──────────────┐
                ▼                  ▼                   ▼
         ┌────────────┐     ┌────────────┐     ┌────────────┐
         │ PostgreSQL │     │   Redis    │     │ Typesense  │
         │  Primary DB│     │ Cache/Lock │     │   Search   │
         └────────────┘     └────────────┘     └────────────┘
                                  │
                                  ▼
                           ┌────────────┐
                           │ RabbitMQ   │
                           │ Message    │
                           │ Queue      │
                           └─────┬──────┘
                                 ▼
                         ┌──────────────┐
                         │ Invoice      │
                         │ Worker       │
                         └──────────────┘
```

---

## 🔄 Order Processing

Orders are created inside a database transaction.

```text id="3j8qj4"
Create Order
     ↓
Lock Product Row
     ↓
Validate Stock
     ↓
Create Order Items
     ↓
Deduct Inventory
     ↓
Calculate Total
     ↓
Commit Transaction
     ↓
Publish Invoice Event
     ↓
RabbitMQ
     ↓
Background Worker
```

If any operation fails, the transaction is rolled back.

---

## 🔒 Concurrency & Distributed Locking

Redis is used for distributed coordination during inventory-related operations.

```text id="q6yd3j"
Request
   ↓
Acquire Redis Lock
   ↓
Critical Section
   ↓
Inventory Operation
   ↓
Release Lock
```

The implementation uses Redis `SETNX` with an expiration window to prevent concurrent operations from entering the same critical section.

---

## ⚡ Redis Caching

Product listing responses are cached using a **Cache-Aside pattern**.

```text id="kz6i3v"
GET /products
      ↓
Redis
  │
  ├── HIT ───────► Return Cached Response
  │
  └── MISS
       ↓
   PostgreSQL
       ↓
   Store in Redis
       ↓
   Return Response
```

Cached product queries use a dynamic key based on:

```text id="3s5l8g"
page + limit + sort + search
```

The current cache TTL is **5 minutes**.

---

## 🔎 Typesense Search

Typesense is integrated as a dedicated product search engine.

It supports:

* Full-text product search
* Typo tolerance
* Category filtering
* Faceted search
* Brand filtering
* Price filtering
* Relevance-based sorting

```text id="0r0h15"
Product Created / Updated
          ↓
     PostgreSQL
          ↓
    Typesense Index
          ↓
      Fast Search
```

---

## 📨 RabbitMQ & Asynchronous Processing

Invoice generation is decoupled from the main request flow using RabbitMQ.

```text id="h3qj9s"
Order Created
     ↓
Publish Invoice Event
     ↓
RabbitMQ
     ↓
Invoice Worker
     ↓
Process Event
```

The messaging architecture also includes a **Dead Letter Exchange and Dead Letter Queue** for failed messages.

```text id="w5t9r3"
Invoice Queue
     │
     ▼
Processing
     │
  Failure
     │
     ▼
Dead Letter Exchange
     │
     ▼
Dead Letter Queue
```

---

## 🛡️ Rate Limiting

Redis implements a **sliding-window rate limiter** using sorted sets.

```text id="e8m4a7"
Request
   ↓
Redis Sorted Set
   ↓
Remove Expired Requests
   ↓
Count Current Window
   ↓
Allow / Reject
```

This protects endpoints from excessive request bursts and abuse.

---

## 🛠️ Tech Stack

| Layer                 | Technology              |
| --------------------- | ----------------------- |
| **Language**          | Go 1.25+                |
| **Framework**         | Gin                     |
| **ORM**               | GORM                    |
| **Database**          | PostgreSQL              |
| **Cache**             | Redis                   |
| **Messaging**         | RabbitMQ                |
| **Search**            | Typesense               |
| **Authentication**    | JWT                     |
| **Password Security** | bcrypt                  |
| **Containerization**  | Docker / Docker Compose |
| **Testing**           | Go Testing + SQLite     |

---

## 📁 Project Structure

```text id="7c3wq1"
ecommerce_backend_go/
├── controller/
│   ├── auth_controller.go
│   ├── cart_controller.go
│   ├── order_controller.go
│   └── product_controller.go
│
├── services/
│   ├── auth_service.go
│   ├── order_service.go
│   └── *_test.go
│
├── models/
│   ├── user.go
│   ├── product.go
│   ├── cart.go
│   ├── order.go
│   └── order_item.go
│
├── Middleware/
│   └── auth_middleware.go
│
├── Databse/
│   ├── db.go
│   ├── redis.go
│   ├── rabbitmq.go
│   └── typesense.go
│
├── workers/
├── main.go
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

---

## 🔌 API

### Public

```http id="1h2c5q"
POST /signup
POST /login

GET  /products
GET  /products/:id
GET  /search
```

### Authenticated User

```http id="4h9k6n"
POST /cart
GET  /cart
POST /orders
```

### Admin

```http id="f1u3k0"
POST   /admin/products
PUT    /admin/products/:id
DELETE /admin/products/:id
```

---

## ⚙️ Getting Started

### Prerequisites

* Go 1.25+
* Docker & Docker Compose
* PostgreSQL
* Redis
* RabbitMQ
* Typesense

### Clone

```bash id="6p3n8j"
git clone https://github.com/AlinAlexMyladoor/ecommerce_backend_go.git
cd ecommerce_backend_go
```

### Configure Environment

Create a `.env` file:

```env id="1w5q2p"
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ecommerce
DB_PORT=5432

REDIS_ADDR=localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

TYPESENSE_URL=http://localhost:8108
TYPESENSE_API_KEY=your_api_key
JWT_SECRET=your_secret
```

### Start Infrastructure

```bash id="2j5v7m"
docker compose up -d
```

### Run Application

```bash id="r6k8s1"
go mod download
go run .
```

The API runs on:

```text id="x2k7pv"
http://localhost:8080
```

---

## 🧪 Testing

Run the test suite:

```bash id="g4q1yb"
go test ./...
```

The project includes service-level tests covering authentication and transactional order processing.

---

## 💡 Engineering Concepts

* REST API design
* Layered Go architecture
* JWT authentication & RBAC
* PostgreSQL transactions
* Row-level database locking
* Redis caching
* Distributed locking
* Sliding-window rate limiting
* RabbitMQ asynchronous messaging
* Dead Letter Queues
* Typesense search indexing
* Background workers
* Inventory consistency
* Pagination and filtering
* Unit testing
* Dockerized infrastructure

---

## 🚀 Future Improvements

* Refresh-token authentication
* API documentation with Swagger/OpenAPI
* Structured logging
* Prometheus/Grafana observability
* Distributed tracing
* Cache invalidation strategy
* Retry policies for RabbitMQ consumers
* Comprehensive integration testing
* CI/CD pipeline
* Cloud deployment
* Kubernetes orchestration

---

## 👨‍💻 Author

### Alin Alex

Computer Science & Engineering Student
Christ College of Engineering, Kerala, India

**GitHub:**
https://github.com/AlinAlexMyladoor

**Repository:**
https://github.com/AlinAlexMyladoor/ecommerce_backend_go

---

⭐ If you find this project useful, consider giving the repository a star.
