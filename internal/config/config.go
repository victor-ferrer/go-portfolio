package config

import (
	"fmt"
	"net/url"
	"os"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
	DBSSLMode      string
	MigrationsPath string
	// dsnOverride holds the value of DATABASE_DSN when set; callers always
	// access the connection string via DSN(), so this field is intentionally
	// unexported.
	dsnOverride string
}

// Load reads configuration from environment variables and returns a Config.
// If DATABASE_DSN is set, it is used directly and individual DB_* variables are not required.
// Otherwise, the following variables are required: DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD.
// Optional: DB_SSLMODE (default: "disable"), MIGRATIONS_PATH (default: "file://./migrations").
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.MigrationsPath = os.Getenv("MIGRATIONS_PATH")
	if cfg.MigrationsPath == "" {
		cfg.MigrationsPath = "file://./migrations"
	}

	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		cfg.dsnOverride = dsn
		return cfg, nil
	}

	cfg.DBHost = os.Getenv("DB_HOST")
	cfg.DBPort = os.Getenv("DB_PORT")
	cfg.DBName = os.Getenv("DB_NAME")
	cfg.DBUser = os.Getenv("DB_USER")
	cfg.DBPassword = os.Getenv("DB_PASSWORD")
	cfg.DBSSLMode = os.Getenv("DB_SSLMODE")

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST environment variable is not set")
	}
	if cfg.DBPort == "" {
		return nil, fmt.Errorf("DB_PORT environment variable is not set")
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("DB_NAME environment variable is not set")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER environment variable is not set")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is not set")
	}

	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}

	return cfg, nil
}

// DSN builds and returns a PostgreSQL connection string from the configuration fields.
// If DATABASE_DSN was provided, it is returned directly.
func (c *Config) DSN() string {
	if c.dsnOverride != "" {
		return c.dsnOverride
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DBUser, c.DBPassword),
		Host:   fmt.Sprintf("%s:%s", c.DBHost, c.DBPort),
		Path:   "/" + c.DBName,
	}
	q := u.Query()
	q.Set("sslmode", c.DBSSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}
