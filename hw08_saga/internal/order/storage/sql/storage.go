package order_sqlstorage

import (
	"context"
	"fmt"
	"strconv"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

type Storage struct {
	db *sqlx.DB
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Connect(ctx context.Context, dsn string) error {
	var err error
	s.db, err = sqlx.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("cannot open pgx driver: %w", err)
	}
	return s.db.PingContext(ctx)
}

func (s *Storage) CreateSchema() error {
	var err error
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS users_profile (id integer primary key, phone text, age text, avatar_uri text);`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) FindUserById(id string) (order_app.User, error) {
	user := order_app.User{}
	rows, err := s.db.NamedQuery(`SELECT * FROM users_profile WHERE id=:id`, map[string]interface{}{"id": id})
	if err != nil {
		return user, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.StructScan(&user)
		if err != nil {
			return user, err
		}
	}
	return user, nil
}

func (s *Storage) CreateUser(user order_app.User) error {
	if user.Id == 0 {
		return order_app.ErrUserIDNotSet
	}
	_, err := s.db.NamedExec(`INSERT INTO users_profile (id, phone, age, avatar_uri)
	 VALUES (:id,:phone,:age,:avatar_uri)`,
		map[string]interface{}{
			"id":         user.Id,
			"phone":      user.Phone,
			"age":        user.Age,
			"avatar_uri": user.AvatarUri,
		})
	return err
}

func (s *Storage) UpdateUser(user order_app.User) error {
	if user.Id == 0 {
		return order_app.ErrUserIDNotSet
	}
	strId := strconv.FormatInt(user.Id, 10)
	result, err := s.db.NamedExec(`UPDATE users_profile SET phone=:phone, age=:age,
	avatar_uri=:avatar_uri 	
	WHERE id = `+`'`+strId+`'`,
		map[string]interface{}{
			"phone":      user.Phone,
			"age":        user.Age,
			"avatar_uri": user.AvatarUri,
		})
	if result != nil {
		rowAffected, errResult := result.RowsAffected()
		if err == nil && rowAffected == 0 && errResult == nil {
			return order_app.ErrUserIDNotExist
		}
	}
	return err
}
