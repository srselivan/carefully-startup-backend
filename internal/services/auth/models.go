package auth

import "time"

type JWTConfig struct {
	JWTAccessExpirationTime  time.Duration
	JWTRefreshExpirationTime time.Duration
	JWTAccessSecretKey       string
	JWTRefreshSecretKey      string
}

type AdminCredentials struct {
	Username string
	Password string
}

func (c AdminCredentials) String() string {
	return c.Username + ":" + c.Password
}
