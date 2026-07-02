package domain

import "errors"

// Order - доменная сущность. Ничего не знает про JSON, БД или HTTP.
type Order struct {
	ID        string
	UserID    string
	ProductID string
	Quantity  int
	Total     float64
	Status    string
}

// NewOrder - фабричный метод, инкапсулирующий бизнес-валидацию.
func NewOrder(id, userID, productID string, quantity int, price float64) (*Order, error) {
	if quantity <= 0 {
		return nil, errors.New("domain: quantity must be positive")
	}
	if price < 0 {
		return nil, errors.New("domain: price cannot be negative")
	}

	return &Order{
		ID:        id,
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Total:     price * float64(quantity),
		Status:    "pending_payment",
	}, nil
}

// MarkAsPaid - доменное поведение.
func (o *Order) MarkAsPaid() {
	o.Status = "paid"
}
