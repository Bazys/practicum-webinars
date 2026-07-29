// Пример 5: LogValuer — кастомное форматирование своих типов.
//
// Запуск:
//
//	go run ./02_slog/05_logvaluer.go
//
// Что демонстрирует:
//   - реализация интерфейса slog.LogValuer для своего типа
//   - скрытие чувствительных полей (пароль не попадает в логи)
//   - маскирование данных (номер карты → ****1234)
//   - как slog.Any автоматически вызывает LogValue()
package main

import (
	"log/slog"
	"os"
	"strings"
)

// User — тип с публичными и приватными полями.
// Реализуем LogValuer, чтобы логировать только безопасные поля.
type User struct {
	ID       int
	Email    string
	Password string // чувствительное поле — НИКОГДА в лог
}

// LogValue реализует slog.LogValuer.
// Возвращаем только id и email, пароль умышленно опускаем.
func (u User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("id", u.ID),
		slog.String("email", maskEmail(u.Email)),
	)
}

func maskEmail(e string) string {
	at := strings.IndexByte(e, '@')
	if at <= 1 {
		return "***"
	}
	return string(e[0]) + "***" + e[at:]
}

// CreditCard — ещё один пример: маскируем номер карты.
type CreditCard struct {
	Number string // 16 цифр
}

func (c CreditCard) LogValue() slog.Value {
	last4 := ""
	if len(c.Number) >= 4 {
		last4 = c.Number[len(c.Number)-4:]
	}
	return slog.GroupValue(
		slog.String("masked", "****"+last4),
	)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	alice := User{ID: 42, Email: "alice@example.com", Password: "super-secret-123"}
	logger.Info("пользователь авторизован", "user", alice)
	// Пароль не появится в выводе! Только id и замаскированный email.

	card := CreditCard{Number: "4111111111111111"}
	logger.Info("оплата прошла", "card", card)
	// → "card":{"masked":"****1111"}

	// Тот же эффект при использовании slog.Any:
	logger.Info("через Any",
		slog.Any("user", alice),
		slog.Any("card", card),
	)
}
