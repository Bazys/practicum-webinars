package memory

import (
	"sync"

	"webinars/architecture/internal/domain"
)

type OrderRepo struct {
	DB map[string]*domain.Order
	Mu sync.Mutex
}

func (r *OrderRepo) Save(order *domain.Order) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.DB[order.ID] = order
	return nil
}
