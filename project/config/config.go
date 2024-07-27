package config

import (
	"os"
)

type Config struct {
	GatewayAddress string
	RedisAddress   string
}

func NewConfig() *Config {
	return &Config{
		GatewayAddress: os.Getenv("GATEWAY_ADDR"),
		RedisAddress:   os.Getenv("REDIS_ADDR"),
	}
}
