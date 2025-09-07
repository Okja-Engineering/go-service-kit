package database

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
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
		{"RLSContextVarName", config.RLSContextVarName, "app.current_tenant_id"},
		{"RLSContextTimeout", config.RLSContextTimeout, time.Hour},
		{"TenantIDPattern", config.TenantIDPattern, `^[a-zA-Z0-9_-]{3,50}$`},
		{"EnableQueryStats", config.EnableQueryStats, true},
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
		WithRLSContextVarName("custom.tenant_id"),
		WithRLSContextTimeout(2*time.Hour),
		WithTenantIDPattern(`^[a-z]{3,10}$`),
		WithQueryStats(false),
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
		{"RLSContextVarName", config.RLSContextVarName, "custom.tenant_id"},
		{"RLSContextTimeout", config.RLSContextTimeout, 2 * time.Hour},
		{"TenantIDPattern", config.TenantIDPattern, `^[a-z]{3,10}$`},
		{"EnableQueryStats", config.EnableQueryStats, false},
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

	if db.queryStats == nil {
		t.Error("Expected queryStats to be initialized")
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

// Test RLS multitenancy functionality
func TestTenantContext(t *testing.T) {
	tenant := TenantContext{
		TenantID: "test-tenant",
		SetAt:    time.Now(),
	}

	if !tenant.IsValid() {
		t.Error("Expected valid tenant context")
	}

	if tenant.String() != "tenant:test-tenant" {
		t.Errorf("Expected string representation, got %s", tenant.String())
	}

	// Test expiration
	if tenant.IsExpired() {
		t.Error("Expected tenant context to not be expired")
	}

	// Test expired context
	expiredTenant := TenantContext{
		TenantID: "expired-tenant",
		SetAt:    time.Now().Add(-2 * time.Hour),
	}
	if !expiredTenant.IsExpired() {
		t.Error("Expected expired tenant context to be expired")
	}
}

func TestWithTenant(t *testing.T) {
	config := &Config{
		Host:                "test-host",
		Port:                5432,
		User:                "test-user",
		Password:            "test-password",
		Database:            "test-db",
		MultitenancyEnabled: true,
	}

	db := NewPostgreSQL(config)
	tenantDB := db.WithTenant("test-tenant")

	if tenantDB == nil {
		t.Error("Expected tenant database instance")
	}

	// Test that the original database is unchanged
	if db.currentTenant != nil {
		t.Error("Expected original database to be unchanged")
	}

	// Test that tenant context is set in new instance
	tenantCtx, err := tenantDB.GetTenantContext(context.Background())
	if err != nil {
		t.Errorf("Expected no error getting tenant context: %v", err)
	}

	if tenantCtx.TenantID != "test-tenant" {
		t.Errorf("Expected tenant ID 'test-tenant', got '%s'", tenantCtx.TenantID)
	}
}

func TestRLSMultitenancyConfiguration(t *testing.T) {
	config := NewConfig(
		WithMultitenancy(true),
		WithRLSContextVarName("custom.tenant_id"),
		WithRLSContextTimeout(2*time.Hour),
		WithTenantIDPattern(`^[a-z]{3,10}$`),
		WithQueryStats(false),
	)

	if !config.MultitenancyEnabled {
		t.Error("Expected multitenancy to be enabled")
	}

	if config.RLSContextVarName != "custom.tenant_id" {
		t.Errorf("Expected RLS context var name 'custom.tenant_id', got '%s'", config.RLSContextVarName)
	}

	if config.RLSContextTimeout != 2*time.Hour {
		t.Errorf("Expected RLS context timeout 2h, got %v", config.RLSContextTimeout)
	}

	if config.TenantIDPattern != `^[a-z]{3,10}$` {
		t.Errorf("Expected tenant ID pattern, got '%s'", config.TenantIDPattern)
	}

	if config.EnableQueryStats {
		t.Error("Expected query stats to be disabled")
	}
}

func TestTenantContextValidation(t *testing.T) {
	// Test valid tenant context
	validTenant := TenantContext{
		TenantID: "valid-tenant",
		SetAt:    time.Now(),
	}
	if !validTenant.IsValid() {
		t.Error("Expected valid tenant context to be validated")
	}

	// Test invalid tenant context
	invalidTenant := TenantContext{
		TenantID: "",
		SetAt:    time.Now(),
	}
	if invalidTenant.IsValid() {
		t.Error("Expected invalid tenant context to be invalid")
	}

	// Test null tenant ID
	nullTenant := TenantContext{
		TenantID: "null",
		SetAt:    time.Now(),
	}
	if nullTenant.IsValid() {
		t.Error("Expected null tenant ID to be invalid")
	}

	// Test undefined tenant ID
	undefinedTenant := TenantContext{
		TenantID: "undefined",
		SetAt:    time.Now(),
	}
	if undefinedTenant.IsValid() {
		t.Error("Expected undefined tenant ID to be invalid")
	}
}

func TestTenantContextStringRepresentation(t *testing.T) {
	tenant := TenantContext{
		TenantID: "test",
		SetAt:    time.Now(),
	}

	expected := "tenant:test"
	if tenant.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, tenant.String())
	}
}

// Test enhanced RLS functionality
func TestTenantIDValidation(t *testing.T) {
	config := &Config{
		MultitenancyEnabled: true,
		TenantIDPattern:     `^[a-zA-Z0-9_-]{3,50}$`,
	}

	db := &PostgreSQL{config: config}

	// Test valid tenant IDs
	validIDs := []string{
		"tenant123",
		"tenant_123",
		"tenant-123",
		"abc",
		"a" + strings.Repeat("b", 49), // 50 characters
	}

	for _, id := range validIDs {
		if err := db.validateTenantID(id); err != nil {
			t.Errorf("Expected valid tenant ID '%s' to pass validation: %v", id, err)
		}
	}

	// Test invalid tenant IDs
	invalidIDs := []string{
		"",                            // empty
		"ab",                          // too short
		"a" + strings.Repeat("b", 51), // too long
		"tenant..123",                 // invalid sequence
		"tenant--123",                 // invalid sequence
		"tenant@123",                  // invalid character
		"tenant 123",                  // space
	}

	for _, id := range invalidIDs {
		if err := db.validateTenantID(id); err == nil {
			t.Errorf("Expected invalid tenant ID '%s' to fail validation", id)
		}
	}
}

func TestRLSConfigurationOptions(t *testing.T) {
	config := NewConfig(
		WithRLSContextVarName("custom.tenant_id"),
		WithRLSContextTimeout(30*time.Minute),
		WithTenantIDPattern(`^[a-z]{3,10}$`),
		WithQueryStats(false),
	)

	if config.RLSContextVarName != "custom.tenant_id" {
		t.Errorf("Expected RLS context var name 'custom.tenant_id', got '%s'", config.RLSContextVarName)
	}

	if config.RLSContextTimeout != 30*time.Minute {
		t.Errorf("Expected RLS context timeout 30m, got %v", config.RLSContextTimeout)
	}

	if config.TenantIDPattern != `^[a-z]{3,10}$` {
		t.Errorf("Expected tenant ID pattern, got '%s'", config.TenantIDPattern)
	}

	if config.EnableQueryStats {
		t.Error("Expected query stats to be disabled")
	}
}

func TestTenantQueryStats(t *testing.T) {
	stats := TenantQueryStats{
		TenantID:        "test-tenant",
		TotalQueries:    100,
		TotalDuration:   5 * time.Second,
		AverageDuration: 50 * time.Millisecond,
		SlowQueries:     10,
		FailedQueries:   5,
		LastQueryAt:     time.Now(),
		TableStats: map[string]int64{
			"users":    50,
			"orders":   30,
			"products": 20,
		},
		QueryTypes: map[string]int64{
			"SELECT": 70,
			"INSERT": 20,
			"UPDATE": 10,
		},
	}

	if stats.TenantID != "test-tenant" {
		t.Errorf("Expected tenant ID 'test-tenant', got '%s'", stats.TenantID)
	}

	if stats.TotalQueries != 100 {
		t.Errorf("Expected total queries 100, got %d", stats.TotalQueries)
	}

	if stats.TotalDuration != 5*time.Second {
		t.Errorf("Expected total duration 5s, got %v", stats.TotalDuration)
	}

	if stats.AverageDuration != 50*time.Millisecond {
		t.Errorf("Expected average duration 50ms, got %v", stats.AverageDuration)
	}

	if stats.SlowQueries != 10 {
		t.Errorf("Expected slow queries 10, got %d", stats.SlowQueries)
	}

	if stats.FailedQueries != 5 {
		t.Errorf("Expected failed queries 5, got %d", stats.FailedQueries)
	}

	if len(stats.TableStats) != 3 {
		t.Errorf("Expected 3 table stats, got %d", len(stats.TableStats))
	}

	if len(stats.QueryTypes) != 3 {
		t.Errorf("Expected 3 query types, got %d", len(stats.QueryTypes))
	}
}

func TestRLSPolicyStruct(t *testing.T) {
	policy := RLSPolicy{
		TableName:        "users",
		PolicyName:       "tenant_isolation",
		PolicyDefinition: "FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::text)",
		IsActive:         true,
		CreatedAt:        time.Now(),
	}

	if policy.TableName != "users" {
		t.Errorf("Expected table name 'users', got '%s'", policy.TableName)
	}

	if policy.PolicyName != "tenant_isolation" {
		t.Errorf("Expected policy name 'tenant_isolation', got '%s'", policy.PolicyName)
	}

	if policy.PolicyDefinition != "FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::text)" {
		t.Errorf("Expected policy definition to match")
	}

	if !policy.IsActive {
		t.Error("Expected policy to be active")
	}
}

func TestQueryStatsInitialization(t *testing.T) {
	db := &PostgreSQL{
		config: &Config{
			MultitenancyEnabled: true,
			EnableQueryStats:    true,
		},
		queryStats: make(map[string]*TenantQueryStats),
	}

	// Test initialization
	db.initializeQueryStats("tenant1")

	if stats, exists := db.queryStats["tenant1"]; !exists {
		t.Error("Expected query stats to be initialized for tenant1")
	} else {
		if stats.TenantID != "tenant1" {
			t.Errorf("Expected tenant ID 'tenant1', got '%s'", stats.TenantID)
		}
		if stats.TotalQueries != 0 {
			t.Errorf("Expected total queries 0, got %d", stats.TotalQueries)
		}
	}

	// Test duplicate initialization doesn't overwrite
	originalStats := db.queryStats["tenant1"]
	db.initializeQueryStats("tenant1")
	if db.queryStats["tenant1"] != originalStats {
		t.Error("Expected duplicate initialization to not overwrite existing stats")
	}
}

func TestQueryStatsUpdate(t *testing.T) {
	db := &PostgreSQL{
		config: &Config{
			MultitenancyEnabled: true,
			EnableQueryStats:    true,
		},
		queryStats: make(map[string]*TenantQueryStats),
	}

	// Initialize stats
	db.initializeQueryStats("tenant1")

	// Update stats - use times that are clearly above/below the 100ms threshold
	db.updateQueryStats("tenant1", 50*time.Millisecond, "SELECT", "users", true)
	db.updateQueryStats("tenant1", 150*time.Millisecond, "INSERT", "orders", false)

	stats := db.queryStats["tenant1"]
	if stats.TotalQueries != 2 {
		t.Errorf("Expected total queries 2, got %d", stats.TotalQueries)
	}

	if stats.SlowQueries != 1 {
		t.Errorf("Expected slow queries 1 (only 150ms > 100ms), got %d", stats.SlowQueries)
	}

	if stats.FailedQueries != 1 {
		t.Errorf("Expected failed queries 1, got %d", stats.FailedQueries)
	}

	if stats.TableStats["users"] != 1 {
		t.Errorf("Expected users table queries 1, got %d", stats.TableStats["users"])
	}

	if stats.QueryTypes["SELECT"] != 1 {
		t.Errorf("Expected SELECT queries 1, got %d", stats.QueryTypes["SELECT"])
	}
}

func TestQueryStatsDisabled(t *testing.T) {
	db := &PostgreSQL{
		config: &Config{
			MultitenancyEnabled: true,
			EnableQueryStats:    false,
		},
		queryStats: make(map[string]*TenantQueryStats),
	}

	// initializeQueryStats still creates the structure even when disabled
	db.initializeQueryStats("tenant1")

	// updateQueryStats should be a no-op when disabled
	db.updateQueryStats("tenant1", 100*time.Millisecond, "SELECT", "users", true)

	// The structure should exist but no queries should be tracked
	if stats, exists := db.queryStats["tenant1"]; !exists {
		t.Error("Expected query stats structure to be created")
	} else if stats.TotalQueries != 0 {
		t.Errorf("Expected no queries to be tracked when disabled, got %d", stats.TotalQueries)
	}
}

func TestTenantContextExpiration(t *testing.T) {
	// Test non-expired context
	recentTenant := TenantContext{
		TenantID: "recent",
		SetAt:    time.Now(),
	}
	if recentTenant.IsExpired() {
		t.Error("Expected recent tenant context to not be expired")
	}

	// Test expired context
	expiredTenant := TenantContext{
		TenantID: "expired",
		SetAt:    time.Now().Add(-2 * time.Hour),
	}
	if !expiredTenant.IsExpired() {
		t.Error("Expected expired tenant context to be expired")
	}

	// Test boundary case (exactly 1 hour) - should be expired
	boundaryTenant := TenantContext{
		TenantID: "boundary",
		SetAt:    time.Now().Add(-1 * time.Hour),
	}
	if !boundaryTenant.IsExpired() {
		t.Error("Expected boundary tenant context to be expired")
	}
}

func TestTenantIDPatternValidation(t *testing.T) {
	// Test default pattern
	pattern := `^[a-zA-Z0-9_-]{3,50}$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Errorf("Failed to compile regex pattern: %v", err)
		return
	}

	matched := re.MatchString("valid-tenant_123")
	if !matched {
		t.Error("Expected 'valid-tenant_123' to match pattern")
	}

	// Test invalid patterns
	invalidIDs := []string{
		"ab",                          // too short
		"a" + strings.Repeat("b", 51), // too long
		"tenant@123",                  // invalid character
		"tenant 123",                  // space
		"tenant..123",                 // invalid sequence
	}

	for _, id := range invalidIDs {
		matched := re.MatchString(id)
		if matched {
			t.Errorf("Expected '%s' to not match pattern", id)
		}
	}
}

func TestMultitenancyDisabledBehavior(t *testing.T) {
	config := &Config{
		MultitenancyEnabled: false,
	}

	db := &PostgreSQL{config: config}

	// All multitenancy methods should return early when disabled
	if err := db.SetTenantContext(context.Background(), "tenant1"); err != nil {
		t.Errorf("Expected no error when multitenancy disabled: %v", err)
	}

	if err := db.ClearTenantContext(context.Background()); err != nil {
		t.Errorf("Expected no error when multitenancy disabled: %v", err)
	}

	if err := db.EnableRLS(context.Background(), "users"); err == nil {
		t.Error("Expected error when trying to enable RLS with multitenancy disabled")
	}

	if err := db.CreateRLSPolicy(context.Background(), "users", "policy", "definition"); err == nil {
		t.Error("Expected error when trying to create RLS policy with multitenancy disabled")
	}

	if err := db.VerifyRLSIsolation(context.Background(), "users"); err == nil {
		t.Error("Expected error when trying to verify RLS isolation with multitenancy disabled")
	}

	if _, err := db.GetTenantQueryStats(context.Background()); err == nil {
		t.Error("Expected error when trying to get query stats with multitenancy disabled")
	}
}
