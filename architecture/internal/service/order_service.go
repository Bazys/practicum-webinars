package service

import (
	"errors"
	"fmt"

	"webinars/architecture/internal/domain"
)

// Вспомогательные ошибки для маппинга на верхних слоях
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrProductNotFound = errors.New("product not found") // <-- Добавили это
)

// --- ИНТЕРФЕЙСЫ (Определяются потребителем - Service) ---

type UserChecker interface {
	GetUserName(userID string) (string, error)
}

type ProductInfoProvider interface {
	GetPrice(productID string) (float64, error)
}

type OrderSaver interface {
	Save(order *domain.Order) error
}

type PaymentProcessor interface {
	Charge(userID string, amount float64) error
}

// --- СЕРВИС ---

type OrderService struct {
	userRepo    UserChecker
	productRepo ProductInfoProvider
	orderRepo   OrderSaver
	billing     PaymentProcessor
}

// NewOrderService - Конструктор с Dependency Injection.
// Принимает интерфейсы, возвращает структуру.
func NewOrderService(u UserChecker, p ProductInfoProvider, o OrderSaver, b PaymentProcessor) *OrderService {
	return &OrderService{
		userRepo:    u,
		productRepo: p,
		orderRepo:   o,
		billing:     b,
	}
}

// CreateOrder - Основной бизнес-процесс (Use Case).
// Возвращает доменную ошибку или готовую сущность.
func (s *OrderService) CreateOrder(userID, productID string, quantity int) (*domain.Order, error) {
	// 1. Проверяем пользователя (через интерфейс)
	_, err := s.userRepo.GetUserName(userID)
	if err != nil {
		return nil, fmt.Errorf("service: check user: %w", err)
	}

	// 2. Получаем цену (через интерфейс)
	price, err := s.productRepo.GetPrice(productID)
	if err != nil {
		return nil, fmt.Errorf("service: get product price: %w", err)
	}

	// 3. Создаем доменный объект (вся валидация внутри)
	// Для упрощения генерируем ID здесь, в реальном проекте ID генератор тоже был бы интерфейсом
	orderID := "order-generated-123"
	order, err := domain.NewOrder(orderID, userID, productID, quantity, price)
	if err != nil {
		return nil, err // Возвращаем ошибку валидации как есть
	}

	// 4. Вызываем биллинг (через интерфейс)
	if err := s.billing.Charge(userID, order.Total); err != nil {
		return nil, fmt.Errorf("service: billing failed: %w", err)
	}

	// 5. Меняем статус на уровне домена
	order.MarkAsPaid()

	// 6. Сохраняем (через интерфейс)
	if err := s.orderRepo.Save(order); err != nil {
		return nil, fmt.Errorf("service: save order: %w", err)
	}

	return order, nil
}

// Вспомогательная функция для анализа ошибок в верхних слоях
func IsUserNotFoundError(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
