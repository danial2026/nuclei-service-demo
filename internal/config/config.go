package config

import (
	"os"
	"strconv"
)

// DB represents the database configuration
type DB struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Config represents the application configuration
type Config struct {
	Server struct {
		Port        int    `json:"port"`
		Host        string `json:"host"`
		DemoPort    int    `json:"demo_port"`
		DemoHost    string `json:"demo_host"`
		DemoEnabled bool   `json:"demo_enabled"`
	} `json:"server"`
	DB     DB `json:"db"`
	Nuclei struct {
		TemplatesDir    string `json:"templates_dir"`
		Concurrency     int    `json:"concurrency"`
		RateLimit       int    `json:"rate_limit"`
		Timeout         int    `json:"timeout"`
		Retries         int    `json:"retries"`
		Headless        bool   `json:"headless"`
		FollowRedirects bool   `json:"follow_redirects"`
	} `json:"nuclei"`
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Server configuration
	cfg.Server.Port = getEnvAsInt("SERVER_PORT", 3742)
	cfg.Server.Host = getEnv("SERVER_HOST", "localhost")

	// Database configuration
	cfg.DB.Host = getEnv("DB_HOST", "nuclei-postgres")
	cfg.DB.Port = getEnvAsInt("DB_PORT", 15432)
	cfg.DB.User = getEnv("DB_USER", "postgres")
	cfg.DB.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.DB.Name = getEnv("DB_NAME", "nuclei")

	// Demo configuration
	cfg.Server.DemoPort = getEnvAsInt("DEMO_PORT", 3743)
	cfg.Server.DemoHost = getEnv("DEMO_HOST", "localhost")
	cfg.Server.DemoEnabled = getEnvAsBool("DEMO_ENABLED", true)

	// Nuclei configuration
	cfg.Nuclei.TemplatesDir = getEnv("NUCLEI_TEMPLATES_DIR", "./templates")
	cfg.Nuclei.Concurrency = getEnvAsInt("NUCLEI_CONCURRENCY", 10)
	cfg.Nuclei.RateLimit = getEnvAsInt("NUCLEI_RATE_LIMIT", 100)
	cfg.Nuclei.Timeout = getEnvAsInt("NUCLEI_TIMEOUT", 30)
	cfg.Nuclei.Retries = getEnvAsInt("NUCLEI_RETRIES", 3)
	cfg.Nuclei.Headless = getEnvAsBool("NUCLEI_HEADLESS", false)
	cfg.Nuclei.FollowRedirects = getEnvAsBool("NUCLEI_FOLLOW_REDIRECTS", true)

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
