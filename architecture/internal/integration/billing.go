package integration

import "errors"

var ErrInsufficientFunds = errors.New("insufficient funds")

type MockBillingClient struct{}

func (c *MockBillingClient) Charge(userID string, amount float64) error {
	// Эмулируем ответ внешнего API
	if userID == "user-2" {
		return ErrInsufficientFunds
	}
	return nil
}
