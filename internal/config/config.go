package config

import "os"

const defaultPort = "50051"

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	databaseURL := os.Getenv("DATABASE_URL")

	if port == "" {
		port = defaultPort
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
	}
}

func (c Config) Address() string {
	return ":" + c.Port
}
