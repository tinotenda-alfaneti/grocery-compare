package config

import "os"

type Config struct {
	HTTPPort string
	DBPath   string
	WebRoot  string
	AppEnv   string
}

func Load() Config {
	return Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		DBPath:   getEnv("DB_PATH", "./data/grocery.db"),
		WebRoot:  getEnv("WEB_ROOT", ""),
		AppEnv:   getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
