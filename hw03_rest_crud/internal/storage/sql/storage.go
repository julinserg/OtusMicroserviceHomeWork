package sqlstorage

import (
	"context"
	"fmt"
	"strconv"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw03_rest_crud/internal/app"
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
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS users (id integer primary key, username text, 
		firstname text, lastname text, email text, phone text);`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) FindUserById(id string) (app.User, error) {
	user := app.User{}
	rows, err := s.db.NamedQuery(`SELECT * FROM users WHERE id=:id`, map[string]interface{}{"id": id})
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

func (s *Storage) CreateUser(user app.User) error {
	if user.Id == 0 {
		return app.ErrUserIDNotSet
	}
	_, err := s.db.NamedExec(`INSERT INTO users (id,username,firstname,lastname,email,phone)
	 VALUES (:id,:username,:firstname,:lastname,:email,:phone)`,
		map[string]interface{}{
			"id":        user.Id,
			"username":  user.Username,
			"firstname": user.FirstName,
			"lastname":  user.LastName,
			"email":     user.Email,
			"phone":     user.Phone,
		})
	return err
}

func (s *Storage) UpdateUser(user app.User) error {
	if user.Id == 0 {
		return app.ErrUserIDNotSet
	}
	strId := strconv.FormatInt(user.Id, 10)
	result, err := s.db.NamedExec(`UPDATE users SET username=:username, firstname=:firstname,
	lastname=:lastname,email=:email, 
	phone =:phone 
	WHERE id = `+`'`+strId+`'`,
		map[string]interface{}{
			"username":  user.Username,
			"firstname": user.FirstName,
			"lastname":  user.LastName,
			"email":     user.Email,
			"phone":     user.Phone,
		})
	if result != nil {
		rowAffected, errResult := result.RowsAffected()
		if err == nil && rowAffected == 0 && errResult == nil {
			return app.ErrUserIDNotExist
		}
	}
	return err
}

func (s *Storage) DeleteUser(id string) error {
	result, err := s.db.Exec(`DELETE FROM users	WHERE id = ` + `'` + id + `'`)
	rowAffected, errResult := result.RowsAffected()
	if err == nil && rowAffected == 0 && errResult == nil {
		return app.ErrUserIDNotExist
	}
	return err
}
