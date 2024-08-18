package storages

import (
	"bootcamp_task/config"
	"bootcamp_task/storage/entities"
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
)

type Storage struct {
	flats   FlatStorage
	homes   HomeStorage
	users   UserStorage
	db      *sql.DB
	timeout time.Duration
}

func NewStorage(cfg *config.Config) *Storage {
	s := Storage{}
	err := s.Init(
		cfg.BuildPGConnectionString(),
		cfg.Postgres.MaxConnections,
		cfg.Postgres.MaxIdleConnections,
		cfg.Postgres.DataBaseTimeout)
	if err != nil {
		panic(err)
	}
	return &s
}

func (s *Storage) Init(
	connectionString string,
	maxConnections int,
	maxIdleConnections int,
	timeout int) error {
	var err error
	s.db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	if errping := s.db.Ping(); errping != nil {
		return errping
	}

	s.db.SetMaxOpenConns(maxConnections)
	s.db.SetMaxIdleConns(maxIdleConnections)
	s.timeout = time.Duration(timeout) * time.Millisecond
	s.flats = FlatStorage{}
	s.homes = HomeStorage{}
	s.users = UserStorage{}
	return nil
}

func (s *Storage) getConnection() (*sql.Conn, error) {
	conn, errcon := s.db.Conn(context.Background())
	if errcon != nil {
		return nil, errcon
	}
	return conn, nil
}

func (s *Storage) getContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), s.timeout)
}

func (s *Storage) CreateUser(
	email string,
	password string,
	isAdmin bool) (string, error) {
	conn, err := s.getConnection()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.users.CreateUser(conn, ctx, email, password, isAdmin)
}

func (s *Storage) GetUser(email string) (*entities.User, error) {
	conn, err := s.getConnection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.users.GetUser(conn, ctx, email)
}

func (s *Storage) CreateHome(
	address string,
	year int,
	developer string,
	reviewer string) (*entities.Home, error) {
	conn, err := s.getConnection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.homes.CreateHome(conn, ctx, address, year, developer, reviewer)
}

func (s *Storage) CreateFlat(
	flatId int,
	houseId int,
	price int,
	rooms int) (*entities.Flat, error) {
	conn, err := s.getConnection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.flats.CreateFlat(conn, ctx, flatId, houseId, price, rooms)
}

func (s *Storage) UpdateFlat(
	flatId int,
	homeId int,
	price int,
	rooms int,
	status entities.ModerationStatus) (*entities.Flat, error) {
	conn, err := s.getConnection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.flats.UpdateFlat(conn, ctx, flatId, homeId, price, rooms, status)
}

func (s *Storage) FilterFlats(
	homeId int,
	admin bool) ([]entities.Flat, error) {
	conn, err := s.getConnection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.flats.FilterFlats(conn, ctx, homeId, admin)
}

func (s *Storage) GetLastHomeUpdate(homeId int) (time.Time, error) {
	conn, err := s.getConnection()
	if err != nil {
		return time.Unix(0, 0), err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.homes.GetLastHomeUpdate(conn, ctx, homeId)
}

func (s *Storage) GetHomeReviewer(homeId int) (string, error) {
	conn, err := s.getConnection()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.homes.GetHomeReviewer(conn, ctx, homeId)
}
