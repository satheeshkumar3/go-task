package configs

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Logger   LoggerConfig
    Auth     AuthConfig
}

type ServerConfig struct {
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
    Environment  string
}

type DatabaseConfig struct {
    Host         string
    Port         string
    User         string
    Password     string
    DBName       string
    SSLMode      string
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  time.Duration
}

type LoggerConfig struct {
    Level  string
    Format string
}

type AuthConfig struct {
    JWTSecret     string
    TokenDuration time.Duration
}

func LoadConfig() (*Config, error) {
    // Load .env file if it exists
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    config := &Config{
        Server: ServerConfig{
            Port:         getEnv("SERVER_PORT", "8080"),
            ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
            WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
            IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
            Environment:  getEnv("ENVIRONMENT", "development"),
        },
        Database: DatabaseConfig{
            Host:         getEnv("DB_HOST", "localhost"),
            Port:         getEnv("DB_PORT", "5432"),
            User:         getEnv("DB_USER", "postgres"),
            Password:     getEnv("DB_PASSWORD", "postgres"),
            DBName:       getEnv("DB_NAME", "taskmanager"),
            SSLMode:      getEnv("DB_SSL_MODE", "disable"),
            MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 10),
            MaxLifetime:  getDurationEnv("DB_MAX_LIFETIME", 5*time.Minute),
        },
        Logger: LoggerConfig{
            Level:  getEnv("LOG_LEVEL", "info"),
            Format: getEnv("LOG_FORMAT", "json"),
        },
        Auth: AuthConfig{
            JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
            TokenDuration: getDurationEnv("JWT_TOKEN_DURATION", 24*time.Hour),
        },
    }

    return config, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
    )
}