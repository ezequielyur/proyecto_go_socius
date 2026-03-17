package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	RawgKey     string
}

func LoadConfig() *Config {
	error := godotenv.Load(".env")

	if error != nil {
		log.Println("No se encontro el archivo .env")
	}

	appConfig := &Config{
		Port:        getEnv("PORT", "8090"),
		DatabaseURL: getEnv("DATABASE_URL", "root:admin@tcp(localhost:3306)/default_db"),
		RawgKey:     getEnv("RAWG_KEY", "12345"),
	}

	return appConfig
}

func getEnv(key, defaultValue string) string {
	valor := os.Getenv(key)

	if valor == "" {
		return defaultValue
	}

	return valor
}
