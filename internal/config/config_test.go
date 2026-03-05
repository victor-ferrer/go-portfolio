package config

import (
	"testing"
)

func TestLoad_success(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("expected DBHost=localhost, got %s", cfg.DBHost)
	}
	if cfg.DBSSLMode != "disable" {
		t.Errorf("expected default DBSSLMode=disable, got %s", cfg.DBSSLMode)
	}
	if cfg.MigrationsPath != "file://./internal/infrastructure/migrations" {
		t.Errorf("expected default MigrationsPath, got %s", cfg.MigrationsPath)
	}
}

func TestLoad_customSSLModeAndMigrationsPath(t *testing.T) {
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_NAME", "mydb")
	t.Setenv("DB_USER", "admin")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("MIGRATIONS_PATH", "file:///app/migrations")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.DBSSLMode != "require" {
		t.Errorf("expected DBSSLMode=require, got %s", cfg.DBSSLMode)
	}
	if cfg.MigrationsPath != "file:///app/migrations" {
		t.Errorf("expected custom MigrationsPath, got %s", cfg.MigrationsPath)
	}
}

func TestLoad_missingRequired(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "missing_DB_HOST",
			env:  map[string]string{"DB_PORT": "5432", "DB_NAME": "testdb", "DB_USER": "user", "DB_PASSWORD": "pass"},
		},
		{
			name: "missing_DB_PORT",
			env:  map[string]string{"DB_HOST": "localhost", "DB_NAME": "testdb", "DB_USER": "user", "DB_PASSWORD": "pass"},
		},
		{
			name: "missing_DB_NAME",
			env:  map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432", "DB_USER": "user", "DB_PASSWORD": "pass"},
		},
		{
			name: "missing_DB_USER",
			env:  map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432", "DB_NAME": "testdb", "DB_PASSWORD": "pass"},
		},
		{
			name: "missing_DB_PASSWORD",
			env:  map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432", "DB_NAME": "testdb", "DB_USER": "user"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			_, err := Load()
			if err == nil {
				t.Errorf("expected error when a required variable is missing, got nil")
			}
		})
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "go-portfolio",
		DBUser:     "portfolio",
		DBPassword: "portfolio",
		DBSSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN %q, got %q", expected, dsn)
	}
}

func TestDSN_specialCharsInPassword(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "mydb",
		DBUser:     "user",
		DBPassword: "p@ss:word!",
		DBSSLMode:  "disable",
	}

	dsn := cfg.DSN()
	// DSN should be a valid URL with encoded password
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
	// Ensure special chars don't break the URL
	if dsn == "postgres://user:p@ss:word!@localhost:5432/mydb?sslmode=disable" {
		t.Error("password special characters should be URL-encoded in DSN")
	}
}
