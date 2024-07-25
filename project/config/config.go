package config

import (
	"os"
)

type Config struct {
	WebPort        int
	GatewayAddress string
	RedisAddress   string
}

func NewConfig() *Config {
	return &Config{
		WebPort:        8080,
		GatewayAddress: os.Getenv("GATEWAY_ADDR"),
		RedisAddress:   os.Getenv("REDIS_ADDR"),
	}
}
