package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// --- Модели (смешанные в одном файле) ---

type OrderRequest struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Order struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
	Status    string  `json:"status"`
}

// --- Эмуляция инфраструктуры (БД и Внешние сервисы) ---

var (
	usersDB = map[string]string{
		"user-1": "Alice",
		"user-2": "Bob",
	}

	productsDB = map[string]float64{
		"prod-1": 100.0,
		"prod-2": 50.0,
	}

	ordersDB = make(map[string]Order)
	mu       sync.Mutex
)

// --- HTTP Хендлер (Тот самый спагетти-код) ---

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Транспортный слой: парсинг HTTP
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 2. Бизнес-логика: валидация
	if req.Quantity <= 0 {
		http.Error(w, "quantity must be positive", http.StatusBadRequest)
		return
	}

	// 3. Инфраструктура: проверка пользователя в БД
	userName, ok := usersDB[req.UserID]
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// 4. Инфраструктура: проверка товара и получение цены
	price, ok := productsDB[req.ProductID]
	if !ok {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	// 5. Бизнес-логика: расчет стоимости
	total := price * float64(req.Quantity)

	// 6. Доменная логика: создание сущности
	order := Order{
		ID:        fmt.Sprintf("order-%d", len(ordersDB)+1),
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Total:     total,
		Status:    "pending_payment",
	}

	// 7. Инфраструктура: вызов внешнего сервиса Billing
	// Эмулируем вызов. Представим, что здесь http.Post к микросервису биллинга.
	// По легенде, у пользователя "Bob" (user-2) нет денег.
	if userName == "Bob" {
		// Проблема: мы возвращаем HTTP статус из бизнес-логики!
		http.Error(w, "billing failed: insufficient funds", http.StatusPaymentRequired)
		return
	}
	order.Status = "paid"

	// 8. Инфраструктура: сохранение в БД
	mu.Lock()
	ordersDB[order.ID] = order
	mu.Unlock()

	// 9. Транспортный слой: формирование ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func main() {
	http.HandleFunc("/orders", createOrderHandler)

	log.Println("Spaghetti Server starting on :8080...")
	if err := http.ListenAndServe(":8888", nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}
