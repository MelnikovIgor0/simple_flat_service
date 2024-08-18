package storages

import (
	"bootcamp_task/storage/entities"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type UserStorage struct {
}

func (u UserStorage) CreateUser(
	conn *sql.Conn,
	ctx context.Context,
	email string,
	password string,
	isAdmin bool) (string, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}

	query := "INSERT INTO users (id, email, password, is_admin) VALUES ($1, $2, $3, $4)"
	supportiveQuery := "SELECT COUNT(*) FROM users WHERE id=$1"
	id := uuid.New()
	var value int
	for {
		err := txn.QueryRow(supportiveQuery, id.String()).Scan(&value)
		if err != nil {
			return "", err
		}
		if value == 0 {
			break
		}
		id = uuid.New()
	}
	_, err = txn.Exec(query, id.String(), email, password, isAdmin)
	if err != nil {
		return "", err
	}

	err = txn.Commit()
	if err != nil {
		return "", err
	}
	return id.String(), err
}

func (u UserStorage) GetUser(
	conn *sql.Conn,
	ctx context.Context,
	email string) (*entities.User, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	query := "SELECT id, email, password, is_admin FROM users WHERE email=$1"
	user := entities.User{}
	err = txn.QueryRow(query, email).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.IsAdmin,
	)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return &user, nil
}
