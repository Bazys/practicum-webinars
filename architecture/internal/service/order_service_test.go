/**
* Почему в Go пишут Table-Driven Tests
* DRY (Don't Repeat Yourself): Мы не дублируем логику создания моков и вызова
* svc.CreateOrder 4 раза. Мы вынесли это в цикл.
* Читаемость как документация: Структура tests := []struct{...}
* читается как таблица требований.
* Продакт-менеджер может посмотреть на этот код и понять все бизнес-сценарии
* (Happy path, юзер не найден, невалидное количество, ошибка биллинга),
* даже не умея программировать.
* Масштабируемость: Завтра нам приходит требование: 'а что если товар стоит 0 рублей?'.
* Мы просто добавляем еще один элемент в слайс tests — и всё.
* Никаких новых функций писать не нужно.
* Изоляция через t.Run: Каждый кейс выполняется в отдельной подпрограмме.
* Если упадет кейс 'Billing failed', другие кейсы отработают нормально,
* и в терминале мы сразу увидим, где именно произошел провал.
* go test -v ./architecture/internal/service/...
**/
package service

import (
	"errors"
	"strings"
	"testing"

	"webinars/architecture/internal/domain"
)

// --- СОЗДАЕМ МОКИ НАШИХ ИНТЕРФЕЙСОВ ---
type mockUserRepo struct{ err error }

func (m *mockUserRepo) GetUserName(userID string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "Test User", nil
}

type mockProductRepo struct {
	price float64
	err   error
}

func (m *mockProductRepo) GetPrice(productID string) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.price, nil
}

type mockOrderRepo struct{ savedOrder *domain.Order }

func (m *mockOrderRepo) Save(order *domain.Order) error {
	m.savedOrder = order
	return nil
}

type mockBilling struct{ err error }

func (m *mockBilling) Charge(userID string, amount float64) error {
	return m.err
}

// --- ТАБЛИЧНЫЙ ТЕСТ ---

func TestCreateOrder(t *testing.T) {
	// 1. Описываем структуру для наших тест-кейсов
	tests := []struct {
		name      string
		userID    string
		productID string
		quantity  int
		// Функция для гибкой настройки моков под конкретный кейс
		setupMocks  func() (UserChecker, ProductInfoProvider, OrderSaver, PaymentProcessor)
		expectedErr string // Проверяем текст ошибки
		checkResult func(t *testing.T, order *domain.Order, orderSaver *mockOrderRepo)
	}{
		{
			name:      "Happy path",
			userID:    "user-1",
			productID: "prod-1",
			quantity:  2,
			setupMocks: func() (UserChecker, ProductInfoProvider, OrderSaver, PaymentProcessor) {
				return &mockUserRepo{}, &mockProductRepo{price: 100.0}, &mockOrderRepo{}, &mockBilling{}
			},
			expectedErr: "",
			checkResult: func(t *testing.T, order *domain.Order, saver *mockOrderRepo) {
				if order.Status != "paid" {
					t.Errorf("expected status 'paid', got '%s'", order.Status)
				}
				if order.Total != 200.0 {
					t.Errorf("expected total 200.0, got %.2f", order.Total)
				}
				if saver.savedOrder == nil {
					t.Error("expected order to be saved")
				}
			},
		},
		{
			name:      "User not found",
			userID:    "bad-user",
			productID: "prod-1",
			quantity:  1,
			setupMocks: func() (UserChecker, ProductInfoProvider, OrderSaver, PaymentProcessor) {
				return &mockUserRepo{err: ErrUserNotFound}, &mockProductRepo{}, &mockOrderRepo{}, &mockBilling{}
			},
			expectedErr: "user not found",
			checkResult: func(t *testing.T, order *domain.Order, saver *mockOrderRepo) {
				if saver.savedOrder != nil {
					t.Error("order should NOT be saved when user not found")
				}
			},
		},
		{
			name:      "Invalid quantity (Domain validation)",
			userID:    "user-1",
			productID: "prod-1",
			quantity:  -5, // Нарушаем бизнес-правило
			setupMocks: func() (UserChecker, ProductInfoProvider, OrderSaver, PaymentProcessor) {
				return &mockUserRepo{}, &mockProductRepo{price: 100.0}, &mockOrderRepo{}, &mockBilling{}
			},
			expectedErr: "domain: quantity must be positive",
			checkResult: func(t *testing.T, order *domain.Order, saver *mockOrderRepo) {
				if order != nil {
					t.Error("expected nil order when validation fails")
				}
				if saver.savedOrder != nil {
					t.Error("order should NOT be saved when domain validation fails")
				}
			},
		},
		{
			name:      "Billing failed",
			userID:    "user-2",
			productID: "prod-2",
			quantity:  1,
			setupMocks: func() (UserChecker, ProductInfoProvider, OrderSaver, PaymentProcessor) {
				return &mockUserRepo{}, &mockProductRepo{price: 50.0}, &mockOrderRepo{}, &mockBilling{err: errors.New("insufficient funds")}
			},
			expectedErr: "billing failed",
			checkResult: func(t *testing.T, order *domain.Order, saver *mockOrderRepo) {
				if saver.savedOrder != nil {
					t.Error("order should NOT be saved if billing fails")
				}
			},
		},
	}

	// 2. Запускаем кейсы в цикле
	for _, tt := range tests {
		// t.Run создает под-тест. В консоли будет красиво написано:
		// --- PASS: TestCreateOrder/Happy_path (0.00s)
		// --- PASS: TestCreateOrder/Billing_failed (0.00s)
		t.Run(tt.name, func(t *testing.T) {
			// Вызываем настройку моков для конкретного кейса
			userRepo, productRepo, orderRepo, billing := tt.setupMocks()
			mockSaver := orderRepo.(*mockOrderRepo) // Приводим к типу, чтобы проверить сохранение

			// Собираем сервис
			svc := NewOrderService(userRepo, productRepo, orderRepo, billing)

			// Выполняем
			order, err := svc.CreateOrder(tt.userID, tt.productID, tt.quantity)

			// Проверяем ошибки
			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing '%s', got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("expected error to contain '%s', got '%v'", tt.expectedErr, err)
				}
				return // Если ждали ошибку, дальше проверки нет смысла делать
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Вызываем кастомные проверки результата, если они есть
			if tt.checkResult != nil {
				tt.checkResult(t, order, mockSaver)
			}
		})
	}
}
