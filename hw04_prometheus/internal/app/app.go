package app

import "errors"

var (
	ErrUserIDNotSet   = errors.New("User ID not set")
	ErrUserIDNotExist = errors.New("User ID not exist")
)

type User struct {
	Id        int64  `json:"id,omitempty" db:"id"`
	Username  string `json:"username,omitempty" db:"username"`
	FirstName string `json:"firstName,omitempty" db:"firstname"`
	LastName  string `json:"lastName,omitempty" db:"lastname"`
	Email     string `json:"email,omitempty" db:"email"`
	Phone     string `json:"phone,omitempty" db:"phone"`
}

type Storage interface {
	CreateSchema() error
	CreateUser(user User) error
	DeleteUser(id string) error
	FindUserById(id string) (User, error)
	UpdateUser(user User) error
}
