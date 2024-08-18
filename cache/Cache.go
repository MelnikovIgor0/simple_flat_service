package cache

import (
	"bootcamp_task/config"
	"bootcamp_task/storage/entities"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"time"
)

type Cache struct {
	rCl              *redis.Client
	timeout          time.Duration
	sessionTimeout   time.Duration
	flatCacheTimeout time.Duration
}

func NewCache(cfg *config.Config) *Cache {
	c := Cache{}
	err := c.Init(
		cfg.Redis.Host,
		cfg.Redis.Password,
		cfg.Redis.PoolSize,
		cfg.Redis.Timeout,
		cfg.Redis.IdleTimeOut,
		cfg.Redis.SessionTimeout,
		cfg.Redis.FlatCacheTimeout,
	)
	if err != nil {
		panic(err)
	}
	return &c
}

func (c *Cache) Init(
	host string,
	password string,
	poolSize int,
	timeout int,
	idleTimeout int,
	sessionTimeout int,
	flatCacheTimeout int) error {
	c.rCl = redis.NewClient(&redis.Options{
		Addr:         host,
		Password:     password,
		DB:           0,
		PoolSize:     poolSize,
		IdleTimeout:  time.Duration(idleTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(timeout) * time.Millisecond,
		WriteTimeout: time.Duration(timeout) * time.Millisecond,
	})
	c.sessionTimeout = time.Duration(sessionTimeout) * time.Minute
	c.flatCacheTimeout = time.Duration(flatCacheTimeout) * time.Minute
	_, err := c.rCl.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) getConnection() *redis.Conn {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	return c.rCl.Conn(ctx)
}

type userSession struct {
	Uid   string `json:"uid"`
	Admin bool   `json:"admin"`
}

func (c *Cache) CreateSession(
	userId string,
	admin bool) (string, error) {
	conn := c.getConnection()
	defer conn.Close()
	uid := uuid.New()
	for {
		_, err := conn.Get(context.Background(), uid.String()).Result()
		if errors.Is(err, redis.Nil) {
			break
		}
		uid = uuid.New()
	}
	j, err := json.Marshal(userSession{
		userId,
		admin,
	})
	if err != nil {
		return "", err
	}
	return uid.String(), conn.Set(context.Background(), uid.String(), j, c.sessionTimeout).Err()
}

func (c *Cache) GetSession(id string) (string, bool, error) {
	conn := c.getConnection()
	defer conn.Close()
	value, err := conn.Get(context.Background(), id).Result()
	if err != nil {
		return "", false, err
	}
	var response userSession
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return "", false, err
	}
	return response.Uid, response.Admin, nil
}

func (c *Cache) GetFlatsCache(cacheId string) ([]entities.Flat, error) {
	conn := c.getConnection()
	defer conn.Close()
	value, err := conn.Get(context.Background(), cacheId).Result()
	if err != nil {
		return []entities.Flat{}, err
	}
	var response []entities.Flat
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return []entities.Flat{}, err
	}
	return response, nil
}

func (c *Cache) PutFlatsCache(cacheId string, flats []entities.Flat) error {
	body, err := json.Marshal(flats)
	if err != nil {
		return err
	}
	conn := c.getConnection()
	defer conn.Close()
	return conn.Set(context.Background(), cacheId, body, c.flatCacheTimeout).Err()
}
