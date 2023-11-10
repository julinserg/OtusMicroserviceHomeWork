package auth_sqlstorage

import (
	"context"
	"fmt"

	// Register pgx driver for postgresql.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	auth_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw05_auth/internal/auth/app"
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
	_, err = s.db.Query(`CREATE TABLE IF NOT EXISTS users_auth (id serial primary key, login text unique, 
		password text, first_name text, last_name text, email text unique);`)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) GetUser(login string, password string) (auth_app.UserAuth, error) {
	user := auth_app.UserAuth{}
	rows, err := s.db.NamedQuery(`SELECT * FROM users_auth WHERE login=:login AND password=:password`,
		map[string]interface{}{"login": login, "password": password})
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
	user.Password = ""
	return user, nil
}

func (s *Storage) RegisterUser(user auth_app.UserAuth) (int, error) {
	lastInsertId := 0
	err := s.db.QueryRowx(`INSERT INTO users_auth (login, password, first_name, last_name, email)
	 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		user.Login, user.Password, user.FirstName, user.LastName, user.Email).Scan(&lastInsertId)

	return lastInsertId, err
}
