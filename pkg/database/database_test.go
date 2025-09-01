package database

import (
	"database/sql"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	testCases := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", config.Host, "localhost"},
		{"Port", config.Port, 5432},
		{"User", config.User, "postgres"},
		{"Database", config.Database, "postgres"},
		{"SSLMode", config.SSLMode, "require"},
		{"MaxOpenConns", config.MaxOpenConns, 25},
		{"MaxIdleConns", config.MaxIdleConns, 5},
		{"ConnMaxLifetime", config.ConnMaxLifetime, 5 * time.Minute},
		{"ConnMaxIdleTime", config.ConnMaxIdleTime, 5 * time.Minute},
		{"ConnectTimeout", config.ConnectTimeout, 10 * time.Second},
		{"QueryTimeout", config.QueryTimeout, 30 * time.Second},
		{"RetryAttempts", config.RetryAttempts, 3},
		{"RetryDelay", config.RetryDelay, 1 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("Expected %s '%v', got '%v'", tc.name, tc.expected, tc.got)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	// Test with no options (should use defaults)
	config := NewConfig()
	if config.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", config.Host)
	}

	// Test with custom options
	config = NewConfig(
		WithHost("custom-host"),
		WithPort(5433),
		WithUser("custom-user"),
		WithPassword("custom-password"),
		WithDatabase("custom-db"),
		WithSSLMode("disable"),
		WithMaxOpenConns(50),
		WithMaxIdleConns(10),
		WithConnMaxLifetime(10*time.Minute),
		WithConnMaxIdleTime(10*time.Minute),
		WithConnectTimeout(20*time.Second),
		WithQueryTimeout(60*time.Second),
		WithRetryAttempts(5),
		WithRetryDelay(2*time.Second),
	)

	testCases := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", config.Host, "custom-host"},
		{"Port", config.Port, 5433},
		{"User", config.User, "custom-user"},
		{"Password", config.Password, "custom-password"},
		{"Database", config.Database, "custom-db"},
		{"SSLMode", config.SSLMode, "disable"},
		{"MaxOpenConns", config.MaxOpenConns, 50},
		{"MaxIdleConns", config.MaxIdleConns, 10},
		{"ConnMaxLifetime", config.ConnMaxLifetime, 10 * time.Minute},
		{"ConnMaxIdleTime", config.ConnMaxIdleTime, 10 * time.Minute},
		{"ConnectTimeout", config.ConnectTimeout, 20 * time.Second},
		{"QueryTimeout", config.QueryTimeout, 60 * time.Second},
		{"RetryAttempts", config.RetryAttempts, 5},
		{"RetryDelay", config.RetryDelay, 2 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("Expected %s '%v', got '%v'", tc.name, tc.expected, tc.got)
			}
		})
	}
}

func TestNewPostgreSQL(t *testing.T) {
	config := &Config{
		Host:     "test-host",
		Port:     5432,
		User:     "test-user",
		Password: "test-password",
		Database: "test-db",
	}

	db := NewPostgreSQL(config)

	if db.config != config {
		t.Error("Expected config to be set")
	}

	if db.db != nil {
		t.Error("Expected db to be nil before connection")
	}

	if db.closed {
		t.Error("Expected closed to be false")
	}
}

func TestPostgreSQLBuildDSN(t *testing.T) {
	config := &Config{
		Host:     "test-host",
		Port:     5432,
		User:     "test-user",
		Password: "test-password",
		Database: "test-db",
		SSLMode:  "require",
	}

	db := &PostgreSQL{config: config}
	dsn := db.buildDSN()

	expected := "host=test-host port=5432 user=test-user password=test-password dbname=test-db sslmode=require"
	if dsn != expected {
		t.Errorf("Expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestPostgreSQLGetDB(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	if db.GetDB() != nil {
		t.Error("Expected nil when db is not set")
	}

	// Test when db is set
	mockDB := &sql.DB{}
	db.db = mockDB
	if db.GetDB() != mockDB {
		t.Error("Expected mockDB to be returned")
	}
}

func TestPostgreSQLGetStats(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	stats := db.GetStats()
	if stats.OpenConnections != 0 {
		t.Error("Expected zero stats when db is nil")
	}

	// Test when db is set (we can't easily mock sql.DB.Stats() in tests)
	// This test verifies the function doesn't panic
	db.db = &sql.DB{}
	_ = db.GetStats() // Should not panic
}

func TestPostgreSQLSortMigrations(t *testing.T) {
	migrations := []Migration{
		{Version: 3, Description: "Third"},
		{Version: 1, Description: "First"},
		{Version: 2, Description: "Second"},
	}

	db := &PostgreSQL{}
	sorted := db.sortMigrations(migrations)

	if len(sorted) != 3 {
		t.Errorf("Expected 3 migrations, got %d", len(sorted))
	}

	if sorted[0].Version != 1 {
		t.Errorf("Expected first migration version 1, got %d", sorted[0].Version)
	}

	if sorted[1].Version != 2 {
		t.Errorf("Expected second migration version 2, got %d", sorted[1].Version)
	}

	if sorted[2].Version != 3 {
		t.Errorf("Expected third migration version 3, got %d", sorted[2].Version)
	}
}

func TestPostgreSQLClose(t *testing.T) {
	db := &PostgreSQL{}

	// Test closing when already closed
	if err := db.Close(); err != nil {
		t.Errorf("Expected no error when closing already closed db, got %v", err)
	}

	// Test closing when db is nil
	db.db = nil
	if err := db.Close(); err != nil {
		t.Errorf("Expected no error when closing nil db, got %v", err)
	}
}

func TestPostgreSQLHealthCheck(t *testing.T) {
	db := &PostgreSQL{}

	// Test health check when closed
	db.closed = true
	if err := db.HealthCheck(); err == nil {
		t.Error("Expected error when db is closed")
	}

	// Test health check when db is nil
	db.closed = false
	db.db = nil
	if err := db.HealthCheck(); err == nil {
		t.Error("Expected error when db is nil")
	}
}

func TestPostgreSQLGetMigrationVersion(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	if _, err := db.GetMigrationVersion(); err == nil {
		t.Error("Expected error when db is nil")
	}
}

func TestPostgreSQLMigrate(t *testing.T) {
	db := &PostgreSQL{}

	// Test when closed
	db.closed = true
	if err := db.Migrate([]Migration{}); err == nil {
		t.Error("Expected error when db is closed")
	}

	// Test when db is nil
	db.closed = false
	db.db = nil
	if err := db.Migrate([]Migration{}); err == nil {
		t.Error("Expected error when db is nil")
	}
}

func TestNewPostgreSQLWithOptions(t *testing.T) {
	db := NewPostgreSQLWithOptions(
		WithHost("custom-host"),
		WithPort(5433),
		WithUser("custom-user"),
	)

	if db.config.Host != "custom-host" {
		t.Errorf("Expected host 'custom-host', got '%s'", db.config.Host)
	}

	if db.config.Port != 5433 {
		t.Errorf("Expected port 5433, got %d", db.config.Port)
	}

	if db.config.User != "custom-user" {
		t.Errorf("Expected user 'custom-user', got '%s'", db.config.User)
	}

	// Should use defaults for other values
	if db.config.Database != "postgres" {
		t.Errorf("Expected default database 'postgres', got '%s'", db.config.Database)
	}
}

func TestMigrationStruct(t *testing.T) {
	migration := Migration{
		Version:     1,
		Description: "Create users table",
		UpSQL:       "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT)",
		DownSQL:     "DROP TABLE users",
	}

	if migration.Version != 1 {
		t.Errorf("Expected version 1, got %d", migration.Version)
	}

	if migration.Description != "Create users table" {
		t.Errorf("Expected description 'Create users table', got '%s'", migration.Description)
	}

	if migration.UpSQL != "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT)" {
		t.Errorf("Expected UpSQL to match, got '%s'", migration.UpSQL)
	}

	if migration.DownSQL != "DROP TABLE users" {
		t.Errorf("Expected DownSQL to match, got '%s'", migration.DownSQL)
	}
}

func TestConnectionStatsStruct(t *testing.T) {
	stats := ConnectionStats{
		OpenConnections:   10,
		InUse:             5,
		Idle:              5,
		WaitCount:         100,
		WaitDuration:      1 * time.Second,
		MaxIdleClosed:     50,
		MaxLifetimeClosed: 25,
	}

	if stats.OpenConnections != 10 {
		t.Errorf("Expected OpenConnections 10, got %d", stats.OpenConnections)
	}

	if stats.InUse != 5 {
		t.Errorf("Expected InUse 5, got %d", stats.InUse)
	}

	if stats.Idle != 5 {
		t.Errorf("Expected Idle 5, got %d", stats.Idle)
	}

	if stats.WaitCount != 100 {
		t.Errorf("Expected WaitCount 100, got %d", stats.WaitCount)
	}

	if stats.WaitDuration != 1*time.Second {
		t.Errorf("Expected WaitDuration 1s, got %v", stats.WaitDuration)
	}

	if stats.MaxIdleClosed != 50 {
		t.Errorf("Expected MaxIdleClosed 50, got %d", stats.MaxIdleClosed)
	}

	if stats.MaxLifetimeClosed != 25 {
		t.Errorf("Expected MaxLifetimeClosed 25, got %d", stats.MaxLifetimeClosed)
	}
}

func TestPoolStatsStruct(t *testing.T) {
	stats := PoolStats{
		MaxOpenConnections: 25,
		OpenConnections:    10,
		InUse:              5,
		Idle:               5,
		WaitCount:          100,
		WaitDuration:       1 * time.Second,
		MaxIdleClosed:      50,
		MaxLifetimeClosed:  25,
	}

	if stats.MaxOpenConnections != 25 {
		t.Errorf("Expected MaxOpenConnections 25, got %d", stats.MaxOpenConnections)
	}

	if stats.OpenConnections != 10 {
		t.Errorf("Expected OpenConnections 10, got %d", stats.OpenConnections)
	}

	if stats.InUse != 5 {
		t.Errorf("Expected InUse 5, got %d", stats.InUse)
	}

	if stats.Idle != 5 {
		t.Errorf("Expected Idle 5, got %d", stats.Idle)
	}

	if stats.WaitCount != 100 {
		t.Errorf("Expected WaitCount 100, got %d", stats.WaitCount)
	}

	if stats.WaitDuration != 1*time.Second {
		t.Errorf("Expected WaitDuration 1s, got %v", stats.WaitDuration)
	}

	if stats.MaxIdleClosed != 50 {
		t.Errorf("Expected MaxIdleClosed 50, got %d", stats.MaxIdleClosed)
	}

	if stats.MaxLifetimeClosed != 25 {
		t.Errorf("Expected MaxLifetimeClosed 25, got %d", stats.MaxLifetimeClosed)
	}
}

// Test context timeout handling
func TestContextTimeout(t *testing.T) {
	config := &Config{
		ConnectTimeout: 1 * time.Millisecond,
		QueryTimeout:   1 * time.Millisecond,
	}

	db := NewPostgreSQL(config)

	// Test connection timeout
	if err := db.Connect(); err == nil {
		t.Error("Expected connection timeout error")
	}
}

// Test functional options
func TestFunctionalOptions(t *testing.T) {
	config := NewConfig(
		WithHost("test-host"),
		WithPort(5433),
		WithUser("test-user"),
		WithPassword("test-pass"),
		WithDatabase("test-db"),
		WithSSLMode("disable"),
		WithMaxOpenConns(100),
		WithMaxIdleConns(20),
		WithConnMaxLifetime(10*time.Minute),
		WithConnMaxIdleTime(10*time.Minute),
		WithConnectTimeout(5*time.Second),
		WithQueryTimeout(60*time.Second),
		WithRetryAttempts(10),
		WithRetryDelay(5*time.Second),
	)

	testCases := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", config.Host, "test-host"},
		{"Port", config.Port, 5433},
		{"User", config.User, "test-user"},
		{"Password", config.Password, "test-pass"},
		{"Database", config.Database, "test-db"},
		{"SSLMode", config.SSLMode, "disable"},
		{"MaxOpenConns", config.MaxOpenConns, 100},
		{"MaxIdleConns", config.MaxIdleConns, 20},
		{"ConnMaxLifetime", config.ConnMaxLifetime, 10 * time.Minute},
		{"ConnMaxIdleTime", config.ConnMaxIdleTime, 10 * time.Minute},
		{"ConnectTimeout", config.ConnectTimeout, 5 * time.Second},
		{"QueryTimeout", config.QueryTimeout, 60 * time.Second},
		{"RetryAttempts", config.RetryAttempts, 10},
		{"RetryDelay", config.RetryDelay, 5 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("Expected %s '%v', got '%v'", tc.name, tc.expected, tc.got)
			}
		})
	}
}
