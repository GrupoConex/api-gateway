package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	KeycloakURL   string
	KeycloakRealm string
	Routes        map[string]string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{
		Port:          getEnv("PORT", "3000"),
		KeycloakURL:   getEnv("KEYCLOAK_URL", ""),
		KeycloakRealm: getEnv("KEYCLOAK_REALM", "fibex"),
		Routes:        make(map[string]string),
	}

	if cfg.KeycloakURL == "" {
		log.Fatal("Critical: KEYCLOAK_URL must be defined in .env")
	}

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		if strings.HasPrefix(key, "PROXY_") {
			serviceName := strings.ToLower(strings.TrimPrefix(key, "PROXY_"))
			cfg.Routes[serviceName] = pair[1]
		}
	}

	if len(cfg.Routes) == 0 {
		log.Println("Warning: No proxy routes defined. Gateway will only serve health check.")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
