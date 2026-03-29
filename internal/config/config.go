package config


var App Config
type Config struct {
	Port string
	DB DbConfig
	JWT JWTConfig
}
type JWTConfig struct {
	Secret string
}

type DbConfig struct {
	Addr string
}