package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"webinars/architecture/internal/domain"
	"webinars/architecture/internal/service"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

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

	ordersDB = make(map[string]domain.Order)
	mu       sync.Mutex
)

// MockUserRepo реализует service.UserChecker
type MockUserRepo struct{}

func (m *MockUserRepo) GetUserName(userID string) (string, error) {
	name, ok := usersDB[userID]
	if !ok {
		return "", service.ErrUserNotFound
	}
	return name, nil
}

// MockProductRepo реализует service.ProductInfoProvider
type MockProductRepo struct{}

func (m *MockProductRepo) GetPrice(productID string) (float64, error) {
	price, ok := productsDB[productID]
	if !ok {
		return 0, errors.New("product not found")
	}
	return price, nil
}

// MockOrderRepo реализует service.OrderSaver
type MockOrderRepo struct{}

func (m *MockOrderRepo) Save(order *domain.Order) error {
	mu.Lock()
	defer mu.Unlock()
	ordersDB[order.ID] = *order
	return nil
}

// MockBilling реализует service.PaymentProcessor
type MockBilling struct{}

func (m *MockBilling) Charge(userID string, amount float64) error {
	// Эмулируем ошибку биллинга для Bob
	if userID == "user-2" {
		return ErrInsufficientFunds
	}
	return nil
}

// --- HTTP ТРАНСПОРТ ---

type OrderRequest struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// orderService инжектится в хендлер
type OrderHandler struct {
	svc *service.OrderService
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Вызываем Use Case
	order, err := h.svc.CreateOrder(req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		// Маппинг доменных/инфраструктурных ошибок в HTTP статусы (Error Architecture)
		if service.IsUserNotFoundError(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// Проверяем, есть ли внутри ошибка биллинга
		if errors.Is(err, ErrInsufficientFunds) { // В реальном проекте лучше использовать свой тип ошибки BillingError
			http.Error(w, "billing failed: insufficient funds", http.StatusPaymentRequired)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("Error creating order: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func main() {
	// 1. Инициализация адаптеров (Инфраструктура)
	userRepo := &MockUserRepo{}
	productRepo := &MockProductRepo{}
	orderRepo := &MockOrderRepo{}
	billing := &MockBilling{}

	// 2. Сборка графа зависимостей (DI)
	orderService := service.NewOrderService(userRepo, productRepo, orderRepo, billing)

	// 3. Инициализация HTTP хендлеров с зависимостями
	handler := &OrderHandler{svc: orderService}

	// 4. Запуск сервера
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", handler.CreateOrder)

	log.Println("Clean Architecture (Step 1) Server starting on :8080...")
	if err := http.ListenAndServe(":8888", mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}
