package config

import "os"

const defaultPort = "50051"

type Config struct {
	Port string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	return Config{
		Port: port,
	}
}

func (c Config) Address() string {
	return ":" + c.Port
}
