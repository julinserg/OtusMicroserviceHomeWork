package pay_app

import "errors"

var (
	ErrUserIDNotSet   = errors.New("User ID not set")
	ErrUserIDNotExist = errors.New("User ID not exist")
)

type UserAuth struct {
	Id        int64  `json:"id,omitempty" db:"id"`
	Login     string `json:"login,omitempty" db:"login"`
	Password  string `json:"password,omitempty" db:"password"`
	FirstName string `json:"first_name,omitempty" db:"first_name"`
	LastName  string `json:"last_name,omitempty" db:"last_name"`
	Email     string `json:"email,omitempty" db:"email"`
}

type LoginAuth struct {
	Login    string `json:"login,omitempty" db:"login"`
	Password string `json:"password,omitempty" db:"password"`
}

type Storage interface {
	CreateSchema() error
	RegisterUser(user UserAuth) (int, error)
	GetUser(login string, password string) (UserAuth, error)
}
