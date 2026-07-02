package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"webinars/architecture/internal/integration"
	"webinars/architecture/internal/service"
)

type OrderRequest struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.svc.CreateOrder(req.UserID, req.ProductID, req.Quantity)
	if err != nil {
		// Error Mapping на границе транспорта
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrProductNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, integration.ErrInsufficientFunds) {
			http.Error(w, "billing failed: insufficient funds", http.StatusPaymentRequired)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
