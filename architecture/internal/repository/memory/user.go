package memory

import (
	"webinars/architecture/internal/service"
)

type UserRepo struct {
	DB map[string]string
}

func (r *UserRepo) GetUserName(userID string) (string, error) {
	name, ok := r.DB[userID]
	if !ok {
		return "", service.ErrUserNotFound
	}
	return name, nil
}
