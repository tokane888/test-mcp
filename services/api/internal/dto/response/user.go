package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/tokane888/go-repository-template/services/api/internal/domain"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserFromDomain(user *domain.User) User {
	return User{
		ID:        user.ID(),
		Email:     user.Email(),
		Username:  user.Username(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}
}

type UserList struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

func NewUserListFromDomain(users []*domain.User, total int) UserList {
	responses := make([]User, len(users))
	for i, user := range users {
		responses[i] = NewUserFromDomain(user)
	}
	return UserList{
		Users: responses,
		Total: total,
	}
}
