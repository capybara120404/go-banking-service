package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	Security SecurityConfig
}

type DatabaseConfig struct {
	connectionString string
	maxConnections   int
	minConnections   int
	maxConnIdleTime  int
}

type JWTConfig struct {
	Secret string
}

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

type SecurityConfig struct {
	HMACSecret   string
	PGPPublicKey string
	PGPPrivateKey string
}

func LoadConfig() *Config {
	return &Config{
		Database: loadDatabaseConfig(),
		JWT:      loadJWTConfig(),
		SMTP:     loadSMTPConfig(),
		Security: loadSecurityConfig(),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		connectionString: getEnv("DB_CONNECTION_STRING", "postgres://postgres:postgres@localhost:5432/go_banking_service?sslmode=disable"),
		maxConnections:   getIntEnv("DB_MAX_CONNECTIONS", 10),
		minConnections:   getIntEnv("DB_MIN_CONNECTIONS", 2),
		maxConnIdleTime:  getIntEnv("DB_MAX_CONN_IDLE_TIME", 300),
	}
}

func loadJWTConfig() JWTConfig {
	return JWTConfig{
		Secret: getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
	}
}

func loadSMTPConfig() SMTPConfig {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	return SMTPConfig{
		Host: getEnv("SMTP_HOST", "smtp.example.com"),
		Port: port,
		User: getEnv("SMTP_USER", "noreply@example.com"),
		Pass: getEnv("SMTP_PASS", "password"),
	}
}

func loadSecurityConfig() SecurityConfig {
	public_data, err := os.ReadFile("./pgp_public.key")
	if err != nil {
		panic("Failed to read PGP key file")
	}
	private_data, err := os.ReadFile("./pgp_private.key")
	if err != nil {
		panic("Failed to read PGP private key file")
	}
	return SecurityConfig{
		HMACSecret:   getEnv("HMAC_SECRET", "hmac-secret-key-change-in-production"),
		PGPPublicKey: string(public_data),
		PGPPrivateKey: string(private_data),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func (c *DatabaseConfig) GetConnectionString() string {
	return c.connectionString
}

func (c *DatabaseConfig) GetMaxConnections() int {
	return c.maxConnections
}

func (c *DatabaseConfig) GetMinConnections() int {
	return c.minConnections
}

func (c *DatabaseConfig) GetMaxConnIdleTime() int {
	return c.maxConnIdleTime
}
