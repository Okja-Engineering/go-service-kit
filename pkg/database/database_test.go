package database

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"Host", "localhost", config.Host},
		{"Port", 5432, config.Port},
		{"User", "postgres", config.User},
		{"Database", "postgres", config.Database},
		{"SSLMode", "require", config.SSLMode},
		{"MaxOpenConns", 25, config.MaxOpenConns},
		{"MaxIdleConns", 5, config.MaxIdleConns},
		{"ConnMaxLifetime", 5 * time.Minute, config.ConnMaxLifetime},
		{"ConnMaxIdleTime", 5 * time.Minute, config.ConnMaxIdleTime},
		{"ConnectTimeout", 10 * time.Second, config.ConnectTimeout},
		{"QueryTimeout", 30 * time.Second, config.QueryTimeout},
		{"RLSContextVarName", "app.current_tenant_id", config.RLSContextVarName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected != tt.actual {
				t.Errorf("Expected %s %v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig(
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
		WithConnectTimeout(5*time.Second),
		WithQueryTimeout(60*time.Second),
		WithRLSContextVarName("custom.tenant_id"),
	)

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"Host", "custom-host", config.Host},
		{"Port", 5433, config.Port},
		{"User", "custom-user", config.User},
		{"Password", "custom-password", config.Password},
		{"Database", "custom-db", config.Database},
		{"SSLMode", "disable", config.SSLMode},
		{"MaxOpenConns", 50, config.MaxOpenConns},
		{"MaxIdleConns", 10, config.MaxIdleConns},
		{"ConnMaxLifetime", 10 * time.Minute, config.ConnMaxLifetime},
		{"ConnMaxIdleTime", 10 * time.Minute, config.ConnMaxIdleTime},
		{"ConnectTimeout", 5 * time.Second, config.ConnectTimeout},
		{"QueryTimeout", 60 * time.Second, config.QueryTimeout},
		{"RLSContextVarName", "custom.tenant_id", config.RLSContextVarName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected != tt.actual {
				t.Errorf("Expected %s %v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestNewPostgreSQL(t *testing.T) {
	config := DefaultConfig()
	db := NewPostgreSQL(config)

	if db.config != config {
		t.Error("Expected config to be set")
	}

	if db.db != nil {
		t.Error("Expected db to be nil before connection")
	}

	if db.closed != false {
		t.Error("Expected closed to be false")
	}
}

func TestPostgreSQLBuildDSN(t *testing.T) {
	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "testdb",
		SSLMode:  "require",
	}

	db := &PostgreSQL{config: config}
	dsn := db.buildDSN()

	expected := "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=require"
	if dsn != expected {
		t.Errorf("Expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestPostgreSQLGetDB(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	if db.GetDB() != nil {
		t.Error("Expected nil when db is nil")
	}
}

func TestPostgreSQLHealthCheck(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	if err := db.HealthCheck(); err == nil {
		t.Error("Expected error when db is nil")
	}

	// Test when closed
	db.closed = true
	if err := db.HealthCheck(); err == nil {
		t.Error("Expected error when db is closed")
	}
}

func TestPostgreSQLGetStats(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	stats := db.GetStats()
	if stats.OpenConnections != 0 {
		t.Error("Expected zero stats when db is nil")
	}
}

func TestPostgreSQLSetTenantContext(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	ctx := context.Background()
	if err := db.SetTenantContext(ctx, "tenant123"); err == nil {
		t.Error("Expected error when db is nil")
	}

	// Test empty tenant ID
	if err := db.SetTenantContext(ctx, ""); err == nil {
		t.Error("Expected error when tenant ID is empty")
	}
}

func TestPostgreSQLClearTenantContext(t *testing.T) {
	db := &PostgreSQL{}

	// Test when db is nil
	ctx := context.Background()
	if err := db.ClearTenantContext(ctx); err == nil {
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

func TestConnectionStatsStruct(t *testing.T) {
	stats := ConnectionStats{
		OpenConnections:   10,
		InUse:             5,
		Idle:              5,
		WaitCount:         100,
		WaitDuration:      time.Second,
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

	if stats.WaitDuration != time.Second {
		t.Errorf("Expected WaitDuration %v, got %v", time.Second, stats.WaitDuration)
	}

	if stats.MaxIdleClosed != 50 {
		t.Errorf("Expected MaxIdleClosed 50, got %d", stats.MaxIdleClosed)
	}

	if stats.MaxLifetimeClosed != 25 {
		t.Errorf("Expected MaxLifetimeClosed 25, got %d", stats.MaxLifetimeClosed)
	}
}

func TestTenantContextStruct(t *testing.T) {
	now := time.Now()
	tenant := TenantContext{
		TenantID: "tenant123",
		SetAt:    now,
	}

	if tenant.TenantID != "tenant123" {
		t.Errorf("Expected TenantID 'tenant123', got '%s'", tenant.TenantID)
	}

	if !tenant.SetAt.Equal(now) {
		t.Errorf("Expected SetAt %v, got %v", now, tenant.SetAt)
	}
}

func TestFunctionalOptions(t *testing.T) {
	config := NewConfig(
		WithHost("test-host"),
		WithPort(9999),
		WithUser("test-user"),
		WithPassword("test-password"),
		WithDatabase("test-database"),
		WithSSLMode("disable"),
		WithMaxOpenConns(100),
		WithMaxIdleConns(20),
		WithConnMaxLifetime(15*time.Minute),
		WithConnMaxIdleTime(10*time.Minute),
		WithConnectTimeout(20*time.Second),
		WithQueryTimeout(45*time.Second),
		WithRLSContextVarName("test.tenant_id"),
	)

	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"Host", "test-host", config.Host},
		{"Port", 9999, config.Port},
		{"User", "test-user", config.User},
		{"Password", "test-password", config.Password},
		{"Database", "test-database", config.Database},
		{"SSLMode", "disable", config.SSLMode},
		{"MaxOpenConns", 100, config.MaxOpenConns},
		{"MaxIdleConns", 20, config.MaxIdleConns},
		{"ConnMaxLifetime", 15 * time.Minute, config.ConnMaxLifetime},
		{"ConnMaxIdleTime", 10 * time.Minute, config.ConnMaxIdleTime},
		{"ConnectTimeout", 20 * time.Second, config.ConnectTimeout},
		{"QueryTimeout", 45 * time.Second, config.QueryTimeout},
		{"RLSContextVarName", "test.tenant_id", config.RLSContextVarName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected != tt.actual {
				t.Errorf("Expected %s %v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}
}
