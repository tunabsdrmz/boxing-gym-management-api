package config

import "time"

var App Config

type Config struct {
	Port     string
	DB       DbConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Security SecurityConfig
	Dev      DevConfig
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type DbConfig struct {
	Addr string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type SecurityConfig struct {
	RateLimitRPM int
}

type DevConfig struct {
	ReturnPasswordResetToken bool
}
