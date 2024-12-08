package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	JWT      JWTConfig      `yaml:"jwt"`
	Admin    AdminConfig    `yaml:"admin"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr" env-default:"0.0.0.0:11864"`
}

type PostgresConfig struct {
	URL            string `yaml:"url" env-required:"true"`
	MigrationsPath string `yaml:"migrations_path" env-required:"true"`
}

type JWTConfig struct {
	JWTAccessExpirationTime  time.Duration `yaml:"jwt_access_expiration_time" env-required:"true"`
	JWTRefreshExpirationTime time.Duration `yaml:"jwt_refresh_expiration_time" env-required:"true"`
	JWTAccessSecretKey       string        `yaml:"jwt_access_secret_key" env-required:"true"`
	JWTRefreshSecretKey      string        `yaml:"jwt_refresh_secret_key" env-required:"true"`
}

type AdminConfig struct {
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

func New() *Config {
	var cfg Config
	err := cleanenv.ReadConfig("./config/prod.yml", &cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}
