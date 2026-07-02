package main

import (
	"errors"
	"log"
	"net/http"

	"webinars/architecture/internal/domain"
	"webinars/architecture/internal/integration"
	"webinars/architecture/internal/repository/memory"
	"webinars/architecture/internal/service"
	httpTransport "webinars/architecture/internal/transport/http"
)

func main() {
	// 1. Инициализация инфраструктуры (In-Memory БД)
	userDB := &memory.UserRepo{DB: map[string]string{"user-1": "Alice", "user-2": "Bob"}}
	productDB := &memory.ProductRepo{DB: map[string]float64{"prod-1": 100.0, "prod-2": 50.0}}
	orderDB := &memory.OrderRepo{DB: make(map[string]*domain.Order)}

	billingClient := &integration.MockBillingClient{}

	// 2. Сборка бизнес-логики (Dependency Injection)
	orderService := service.NewOrderService(userDB, productDB, orderDB, billingClient)

	// 3. Инициализация транспортного слоя
	orderHandler := httpTransport.NewOrderHandler(orderService)

	// 4. Запуск сервера
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", orderHandler.CreateOrder)

	log.Println("Clean Architecture (Step 2) Server starting on :8080...")
	if err := http.ListenAndServe(":8888", mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}
