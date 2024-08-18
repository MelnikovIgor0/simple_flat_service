package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	ServerPort int `yaml:"server_port"`
	Postgres   struct {
		User               string `yaml:"user"`
		Database           string `yaml:"database"`
		SSLMode            bool   `yaml:"ssl_mode"`
		Password           string `yaml:"password"`
		Host               string `yaml:"host"`
		Port               int    `yaml:"port"`
		DataBaseTimeout    int    `yaml:"database_timeout"`
		MaxConnections     int    `yaml:"max_connections"`
		MaxIdleConnections int    `yaml:"max_idle_connections"`
	} `yaml:"postgres"`
	Redis struct {
		Host             string `yaml:"host"`
		Password         string `yaml:"password"`
		PoolSize         int    `yaml:"pool_size"`
		Timeout          int    `yaml:"timeout"`
		IdleTimeOut      int    `yaml:"idle_timeout"`
		SessionTimeout   int    `yaml:"session_timeout"`
		FlatCacheTimeout int    `yaml:"flat_cache_timeout"`
	} `yaml:"redis"`
}

func ParseConfig() *Config {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filename, err := filepath.Abs(filepath.Join(currentDir, "config.yaml"))
	if err != nil {
		panic(err)
	}
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var cfg Config
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}

func (c Config) BuildPGConnectionString() string {
	result := "user=" + c.Postgres.User + " dbname=" + c.Postgres.Database + " password=" + c.Postgres.Password + " host=" + c.Postgres.Host + " port=" + strconv.Itoa(c.Postgres.Port)
	if c.Postgres.SSLMode == false {
		result += " sslmode=disable"
	}
	return result
}
