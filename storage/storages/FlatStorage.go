package storages

import (
	"bootcamp_task/storage/entities"
	"context"
	"database/sql"
	"errors"
	"time"
)

type FlatStorage struct {
}

func (f FlatStorage) CreateFlat(
	conn *sql.Conn,
	ctx context.Context,
	flatId int,
	homeId int,
	price int,
	rooms int) (*entities.Flat, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	queryCheck := "SELECT COUNT(*) FROM flats WHERE home_id=$1 AND number=$2"
	var amount int
	err = txn.QueryRow(queryCheck, homeId, flatId).Scan(&amount)
	if err != nil {
		return nil, err
	}
	if amount != 0 {
		return nil, errors.New("flat already exists")
	}

	queryFlat := "INSERT INTO flats (number, price, rooms, home_id, status) VALUES ($1, $2, $3, $4, 'created')"
	queryHome := "UPDATE homes SET updated_at=$1 WHERE id=$2"
	_, err = txn.Exec(queryFlat, flatId, price, rooms, homeId)
	if err != nil {
		return nil, err
	}
	_, err = txn.Exec(queryHome, time.Now().UTC(), homeId)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return &entities.Flat{
		Number: flatId,
		Price:  price,
		HomeId: homeId,
		Rooms:  rooms,
		Status: "created",
	}, nil
}

func (f FlatStorage) getStatus(s entities.ModerationStatus) string {
	return map[entities.ModerationStatus]string{
		entities.CREATED:       "created",
		entities.APPROVED:      "approved",
		entities.DECLINED:      "declined",
		entities.ON_MODERATION: "on_moderation",
	}[s]
}

func (f FlatStorage) UpdateFlat(
	conn *sql.Conn,
	ctx context.Context,
	flatId int,
	homeId int,
	price int,
	rooms int,
	status entities.ModerationStatus) (*entities.Flat, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	queryFlat := "UPDATE flats SET price=$1, rooms=$2, status=$3 WHERE number=$4 AND home_id=$5"
	queryHome := "UPDATE homes SET updated_at=$1 WHERE id=$2"
	_, err = txn.Exec(queryFlat, price, rooms, status, flatId, homeId)
	if err != nil {
		return nil, err
	}
	_, err = txn.Exec(queryHome, time.Now().UTC(), homeId)
	if err != nil {
		return nil, err
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return &entities.Flat{
		Number: flatId,
		HomeId: homeId,
		Price:  price,
		Rooms:  rooms,
		Status: f.getStatus(status),
	}, nil
}

func (f FlatStorage) FilterFlats(
	conn *sql.Conn,
	ctx context.Context,
	homeId int,
	admin bool) ([]entities.Flat, error) {
	defer conn.Close()

	txn, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	if admin {
		query := "UPDATE flats SET status='on_moderation' WHERE home_id=$1 AND status='created'"
		_, err = txn.Exec(query, homeId)
		if err != nil {
			return nil, err
		}
	}

	query := "SELECT number, price, rooms, home_id, status FROM flats WHERE home_id=$1"
	if !admin {
		query += " AND status='approved'"
	}
	rows, err := txn.Query(query, homeId)
	if err != nil {
		return nil, err
	}
	result := make([]entities.Flat, 0)
	for rows.Next() {
		var flat entities.Flat
		errscan := rows.Scan(&flat.Number, &flat.Price, &flat.Rooms, &flat.HomeId, &flat.Status)
		if errscan != nil {
			return nil, errscan
		}
		result = append(result, flat)
	}

	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return result, nil
}
