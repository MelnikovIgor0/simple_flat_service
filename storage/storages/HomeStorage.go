package storages

import (
	"bootcamp_task/storage/entities"
	"context"
	"database/sql"
	"time"
)

type HomeStorage struct {
}

func (h HomeStorage) CreateHome(
	conn *sql.Conn,
	ctx context.Context,
	address string,
	year int,
	developer string,
	reviewer string) (*entities.Home, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	creationTime := time.Now().UTC()
	var insertedId int
	query := "INSERT INTO homes (address, year, created_at, updated_at, developer, reviewer) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	err = txn.QueryRow(query, address, year, creationTime, creationTime, developer, reviewer).Scan(&insertedId)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return &entities.Home{
		Id:        insertedId,
		Address:   address,
		Year:      year,
		CreatedAt: creationTime,
		UpdatedAt: creationTime,
		Developer: developer,
		Reviewer:  reviewer,
	}, nil
}

func (h HomeStorage) GetLastHomeUpdate(
	conn *sql.Conn,
	ctx context.Context,
	homeId int) (time.Time, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return time.Unix(0, 0), err
	}

	query := "SELECT updated_at FROM homes WHERE id=$1"
	var lastUpdated time.Time
	err = txn.QueryRow(query, homeId).Scan(&lastUpdated)
	if err != nil {
		return time.Unix(0, 0), err
	}

	err = txn.Commit()
	if err != nil {
		return time.Unix(0, 0), err
	}
	return lastUpdated, nil
}

func (h HomeStorage) GetHomeReviewer(
	conn *sql.Conn,
	ctx context.Context,
	homeId int) (string, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}

	query := "SELECT reviewer FROM homes WHERE id=$1"
	var reviewer string
	err = txn.QueryRow(query, homeId).Scan(&reviewer)
	if err != nil {
		return "", err
	}

	err = txn.Commit()
	if err != nil {
		return "", err
	}
	return reviewer, nil
}
