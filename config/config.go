package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

type Config struct {
	ServerHost    string        `env:"SERVER_HOST"`
	ServerPort    string        `env:"SERVER_PORT"`
	ServerTimeout time.Duration `env:"SERVER_TIMEOUT"`

	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPass     string `env:"POSTGRES_PASSWORD"`
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresPort     string `env:"POSTGRES_PORT"`
	PostgresDb       string `env:"POSTGRES_DB"`
	PostgresSSLMode  string `env:"POSTGRES_SSL_MODE"`
	PostgresMaxConns string `env:"POSTGRES_MAX_POOL_CONNS"`

	RedisHost string `env:"REDIS_HOST"`
	RedisPort string `env:"REDIS_PORT"`
	RedisPass string `env:"REDIS_PASS"`
	RedisDb   int    `env:"REDIS_DB"`

	LogLevel string `env:"LOG_LEVEL" env-default:"DEBUG"`

	SecretKey string `env:"SECRET_KEY"`
}

func New() *Config {
	config := new(Config)
	err := cleanenv.ReadConfig(".env", config)

	if err != nil {
		log.Fatalf("Get config error: %s", err)
	}

	return config
}
