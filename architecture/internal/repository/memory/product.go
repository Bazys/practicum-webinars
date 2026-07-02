package memory

import (
	"webinars/architecture/internal/service"
)

type ProductRepo struct {
	DB map[string]float64
}

func (r *ProductRepo) GetPrice(productID string) (float64, error) {
	price, ok := r.DB[productID]
	if !ok {
		return 0, service.ErrProductNotFound
	}
	return price, nil
}
