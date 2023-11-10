package users_app

import "errors"

var (
	ErrUserIDNotSet   = errors.New("User ID not set")
	ErrUserIDNotExist = errors.New("User ID not exist")
)

type User struct {
	Id        int64  `json:"id,omitempty" db:"id"`
	Phone     string `json:"phone,omitempty" db:"phone"`
	Age       string `json:"age,omitempty" db:"age"`
	AvatarUri string `json:"avatar_uri,omitempty" db:"avatar_uri"`
}

type UserFull struct {
	Id        int64  `json:"id,omitempty" db:"id"`
	Login     string `json:"login,omitempty" db:"login"`
	Password  string `json:"password,omitempty" db:"password"`
	FirstName string `json:"first_name,omitempty" db:"first_name"`
	LastName  string `json:"last_name,omitempty" db:"last_name"`
	Email     string `json:"email,omitempty" db:"email"`
	Phone     string `json:"phone,omitempty" db:"phone"`
	Age       string `json:"age,omitempty" db:"age"`
	AvatarUri string `json:"avatar_uri,omitempty" db:"avatar_uri"`
}

type Storage interface {
	CreateSchema() error
	CreateUser(user User) error
	FindUserById(id string) (User, error)
	UpdateUser(user User) error
}
