package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	SMS         SMSConfig
	Payment     PaymentConfig
	Bridgetune  BridgetuneConfig
	Telecom     TelecomConfig
	Firebase    FirebaseConfig
	Server      ServerConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// SMSConfig holds SMS provider configuration
type SMSConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
}

// PaymentConfig holds payment gateway configuration
type PaymentConfig struct {
	PaystackSecretKey    string
	PaystackPublicKey    string
	FlutterwaveSecretKey string
	FlutterwavePublicKey string
	WebhookSecret        string
}

// BridgetuneConfig holds Bridgetune API configuration
type BridgetuneConfig struct {
	BaseURL       string
	APIKey        string
	WebhookSecret string
}

// TelecomConfig holds telecom API configuration
type TelecomConfig struct {
	MTN     MTNConfig
	Airtel  AirtelConfig
	Glo     GloConfig
	NineMobile NineMobileConfig
}

// FirebaseConfig holds Firebase configuration
type FirebaseConfig struct {
	CredentialsPath string
	ProjectID       string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	BaseURL      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "rechargemax"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "rechargemax-secret"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},
		SMS: SMSConfig{
			Provider: getEnv("SMS_PROVIDER", "termii"), // Default to Termii, supports: termii, twilio, mock
			APIKey:   getEnv("SMS_API_KEY", ""),
			BaseURL:  getEnv("SMS_BASE_URL", ""),
		},
		Payment: PaymentConfig{
			PaystackSecretKey:    getEnv("PAYSTACK_SECRET_KEY", ""),
			PaystackPublicKey:    getEnv("PAYSTACK_PUBLIC_KEY", ""),
			FlutterwaveSecretKey: getEnv("FLUTTERWAVE_SECRET_KEY", ""),
			FlutterwavePublicKey: getEnv("FLUTTERWAVE_PUBLIC_KEY", ""),
			WebhookSecret:        getEnv("PAYMENT_WEBHOOK_SECRET", "webhook-secret"),
		},
		Bridgetune: BridgetuneConfig{
			BaseURL:       getEnv("BRIDGETUNE_BASE_URL", "https://api.bridgetune.com"),
			APIKey:        getEnv("BRIDGETUNE_API_KEY", ""),
			WebhookSecret: getEnv("BRIDGETUNE_WEBHOOK_SECRET", "bridgetune-secret"),
		},
		Telecom: LoadTelecomConfig(),
		Firebase: FirebaseConfig{
			CredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
			ProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
		},
		Server: ServerConfig{
			BaseURL:      getEnv("SERVER_BASE_URL", "http://localhost:8080"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
