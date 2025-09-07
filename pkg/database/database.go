package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database interface defines the contract for database operations
type Database interface {
	// Core operations
	Connect() error
	Close() error
	GetDB() *sql.DB
	HealthCheck() error
	GetStats() ConnectionStats

	// Migration support
	Migrate(migrations []Migration) error
	GetMigrationVersion() (int, error)

	// RLS Multitenancy support
	WithTenant(tenantID string) Database
	SetTenantContext(ctx context.Context, tenantID string) error
	GetTenantContext(ctx context.Context) (TenantContext, error)
	ClearTenantContext(ctx context.Context) error

	// RLS Management
	EnableRLS(ctx context.Context, tableName string) error
	CreateRLSPolicy(ctx context.Context, tableName, policyName, policyDefinition string) error
	DropRLSPolicy(ctx context.Context, tableName, policyName string) error
	VerifyRLSIsolation(ctx context.Context, tableName string) error
	GetTenantQueryStats(ctx context.Context) (TenantQueryStats, error)
}

// ConnectionStats provides information about database connections
type ConnectionStats struct {
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
}

// PoolStats provides connection pool statistics
type PoolStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
}

// Migration represents a database schema migration
type Migration struct {
	Version     int
	Description string
	UpSQL       string
	DownSQL     string
}

// TenantContext holds tenant-specific information for RLS multitenancy
type TenantContext struct {
	TenantID string    `json:"tenantID"`
	SetAt    time.Time `json:"setAt,omitempty"` // When tenant context was set
}

// TenantQueryStats provides performance metrics for tenant-specific queries
type TenantQueryStats struct {
	TenantID        string           `json:"tenantID"`
	TotalQueries    int64            `json:"totalQueries"`
	TotalDuration   time.Duration    `json:"totalDuration"`
	AverageDuration time.Duration    `json:"averageDuration"`
	SlowQueries     int64            `json:"slowQueries"` // Queries > 100ms
	FailedQueries   int64            `json:"failedQueries"`
	LastQueryAt     time.Time        `json:"lastQueryAt"`
	TableStats      map[string]int64 `json:"tableStats"` // Queries per table
	QueryTypes      map[string]int64 `json:"queryTypes"` // SELECT, INSERT, etc.
}

// RLSPolicy represents a Row Level Security policy
type RLSPolicy struct {
	TableName        string    `json:"tableName"`
	PolicyName       string    `json:"policyName"`
	PolicyDefinition string    `json:"policyDefinition"`
	IsActive         bool      `json:"isActive"`
	CreatedAt        time.Time `json:"createdAt"`
}

// String returns a string representation of the tenant context
func (tc TenantContext) String() string {
	return fmt.Sprintf("tenant:%s", tc.TenantID)
}

// IsValid checks if the tenant context is valid
func (tc TenantContext) IsValid() bool {
	return tc.TenantID != "" && tc.TenantID != "null" && tc.TenantID != "undefined"
}

// IsExpired checks if the tenant context has expired (older than 1 hour)
func (tc TenantContext) IsExpired() bool {
	return time.Since(tc.SetAt) > time.Hour
}

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
	RetryAttempts   int
	RetryDelay      time.Duration

	// RLS Multitenancy configuration
	MultitenancyEnabled bool
	RLSContextVarName   string        // Default: "app.current_tenant_id"
	RLSContextTimeout   time.Duration // Default: 1 hour
	TenantIDPattern     string        // Regex pattern for tenant ID validation
	EnableQueryStats    bool          // Enable tenant query performance tracking
}

// DefaultConfig returns a secure default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		Database:        "postgres",
		SSLMode:         "require",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnectTimeout:  10 * time.Second,
		QueryTimeout:    30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,

		// RLS Multitenancy defaults
		MultitenancyEnabled: false,
		RLSContextVarName:   "app.current_tenant_id",
		RLSContextTimeout:   time.Hour,
		TenantIDPattern:     `^[a-zA-Z0-9_-]{3,50}$`, // Alphanumeric, underscore, hyphen, 3-50 chars
		EnableQueryStats:    true,
	}
}

// Option is a functional option for configuring the database
type Option func(*Config)

// WithHost sets the database host
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the database port
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithUser sets the database user
func WithUser(user string) Option {
	return func(c *Config) {
		c.User = user
	}
}

// WithPassword sets the database password
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDatabase sets the database name
func WithDatabase(database string) Option {
	return func(c *Config) {
		c.Database = database
	}
}

// WithSSLMode sets the SSL mode
func WithSSLMode(sslMode string) Option {
	return func(c *Config) {
		c.SSLMode = sslMode
	}
}

// WithMaxOpenConns sets the maximum number of open connections
func WithMaxOpenConns(maxOpen int) Option {
	return func(c *Config) {
		c.MaxOpenConns = maxOpen
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(maxIdle int) Option {
	return func(c *Config) {
		c.MaxIdleConns = maxIdle
	}
}

// WithConnMaxLifetime sets the maximum lifetime of connections
func WithConnMaxLifetime(lifetime time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = lifetime
	}
}

// WithConnMaxIdleTime sets the maximum idle time of connections
func WithConnMaxIdleTime(idleTime time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxIdleTime = idleTime
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = timeout
	}
}

// WithQueryTimeout sets the query timeout
func WithQueryTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.QueryTimeout = timeout
	}
}

// WithRetryAttempts sets the number of retry attempts
func WithRetryAttempts(attempts int) Option {
	return func(c *Config) {
		c.RetryAttempts = attempts
	}
}

// WithRetryDelay sets the retry delay
func WithRetryDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.RetryDelay = delay
	}
}

// WithMultitenancy enables RLS multitenancy support
func WithMultitenancy(enabled bool) Option {
	return func(c *Config) {
		c.MultitenancyEnabled = enabled
	}
}

// WithRLSContextVarName sets the PostgreSQL variable name for RLS context
func WithRLSContextVarName(varName string) Option {
	return func(c *Config) {
		c.RLSContextVarName = varName
	}
}

// WithRLSContextTimeout sets the timeout for RLS context
func WithRLSContextTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.RLSContextTimeout = timeout
	}
}

// WithTenantIDPattern sets the regex pattern for tenant ID validation
func WithTenantIDPattern(pattern string) Option {
	return func(c *Config) {
		c.TenantIDPattern = pattern
	}
}

// WithQueryStats enables or disables tenant query performance tracking
func WithQueryStats(enabled bool) Option {
	return func(c *Config) {
		c.EnableQueryStats = enabled
	}
}

// NewConfig creates a new database configuration with options
func NewConfig(options ...Option) *Config {
	config := DefaultConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

// PostgreSQL implementation
type PostgreSQL struct {
	config *Config
	db     *sql.DB
	mu     sync.RWMutex
	closed bool

	// RLS Multitenancy support
	currentTenant *TenantContext
	tenantMu      sync.RWMutex

	// Query statistics tracking
	queryStats map[string]*TenantQueryStats
	statsMu    sync.RWMutex
}

// NewPostgreSQL creates a new PostgreSQL database instance
func NewPostgreSQL(config *Config) *PostgreSQL {
	return &PostgreSQL{
		config:     config,
		queryStats: make(map[string]*TenantQueryStats),
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgreSQL) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("database connection is closed")
	}

	dsn := p.buildDSN()

	var err error
	p.db, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	p.db.SetMaxOpenConns(p.config.MaxOpenConns)
	p.db.SetMaxIdleConns(p.config.MaxIdleConns)
	p.db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
	p.db.SetConnMaxIdleTime(p.config.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), p.config.ConnectTimeout)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("### 🗄️ Database: Connected to PostgreSQL at %s:%d/%s",
		p.config.Host, p.config.Port, p.config.Database)

	return nil
}

// Close closes the database connection
func (p *PostgreSQL) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || p.db == nil {
		return nil
	}

	if err := p.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	p.closed = true
	log.Printf("### 🗄️ Database: Connection closed")

	return nil
}

// GetDB returns the underlying sql.DB instance
func (p *PostgreSQL) GetDB() *sql.DB {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db
}

// HealthCheck verifies the database connection is healthy
func (p *PostgreSQL) HealthCheck() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed || p.db == nil {
		return fmt.Errorf("database connection is closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.config.QueryTimeout)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats returns connection pool statistics
func (p *PostgreSQL) GetStats() ConnectionStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.db == nil {
		return ConnectionStats{}
	}

	return ConnectionStats{
		OpenConnections:   p.db.Stats().OpenConnections,
		InUse:             p.db.Stats().InUse,
		Idle:              p.db.Stats().Idle,
		WaitCount:         p.db.Stats().WaitCount,
		WaitDuration:      p.db.Stats().WaitDuration,
		MaxIdleClosed:     p.db.Stats().MaxIdleClosed,
		MaxLifetimeClosed: p.db.Stats().MaxLifetimeClosed,
	}
}

// Migrate runs database migrations
func (p *PostgreSQL) Migrate(migrations []Migration) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed || p.db == nil {
		return fmt.Errorf("database connection is closed")
	}

	// Create migrations table if it doesn't exist
	if err := p.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion, err := p.GetMigrationVersion()
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	// Sort migrations by version
	sortedMigrations := p.sortMigrations(migrations)

	// Apply pending migrations
	for _, migration := range sortedMigrations {
		if migration.Version > currentVersion {
			if err := p.applyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
			log.Printf("### 🗄️ Database: Applied migration %d: %s",
				migration.Version, migration.Description)
		}
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func (p *PostgreSQL) GetMigrationVersion() (int, error) {
	if p.db == nil {
		return 0, fmt.Errorf("database connection is closed")
	}

	var version int
	query := `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`

	ctx, cancel := context.WithTimeout(context.Background(), p.config.QueryTimeout)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, nil
}

// buildDSN builds the PostgreSQL connection string
func (p *PostgreSQL) buildDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host, p.config.Port, p.config.User, p.config.Password,
		p.config.Database, p.config.SSLMode)
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func (p *PostgreSQL) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	ctx, cancel := context.WithTimeout(context.Background(), p.config.QueryTimeout)
	defer cancel()

	_, err := p.db.ExecContext(ctx, query)
	return err
}

// applyMigration applies a single migration
func (p *PostgreSQL) applyMigration(migration Migration) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.config.QueryTimeout)
	defer cancel()

	// Start transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			// Log rollback error but don't fail the migration
			log.Printf("Warning: failed to rollback transaction: %v", err)
		}
	}()

	// Execute migration
	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	recordQuery := `INSERT INTO schema_migrations (version, description) VALUES ($1, $2)`
	if _, err := tx.ExecContext(ctx, recordQuery, migration.Version, migration.Description); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// sortMigrations sorts migrations by version
func (p *PostgreSQL) sortMigrations(migrations []Migration) []Migration {
	// Simple bubble sort for small migration lists
	sorted := make([]Migration, len(migrations))
	copy(sorted, migrations)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Version > sorted[j+1].Version {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// NewPostgreSQLWithOptions creates a new PostgreSQL instance with options
func NewPostgreSQLWithOptions(options ...Option) *PostgreSQL {
	config := NewConfig(options...)
	return NewPostgreSQL(config)
}

// RLS Multitenancy methods

// WithTenant returns a new database instance configured for the specified tenant
func (p *PostgreSQL) WithTenant(tenantID string) Database {
	if !p.config.MultitenancyEnabled {
		return p
	}

	tenant := &TenantContext{
		TenantID: tenantID,
		SetAt:    time.Now(),
	}

	// Create a new instance with the tenant context
	newDB := &PostgreSQL{
		config:        p.config,
		db:            p.db,
		mu:            sync.RWMutex{},
		closed:        p.closed,
		currentTenant: tenant,
		tenantMu:      sync.RWMutex{},
		queryStats:    make(map[string]*TenantQueryStats),
		statsMu:       sync.RWMutex{},
	}

	return newDB
}

// SetTenantContext sets the tenant context for the current database session
func (p *PostgreSQL) SetTenantContext(ctx context.Context, tenantID string) error {
	if !p.config.MultitenancyEnabled {
		return nil
	}

	// Validate tenant ID
	if err := p.validateTenantID(tenantID); err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	p.tenantMu.Lock()
	defer p.tenantMu.Unlock()

	// Set RLS context variable
	query := `SELECT set_config($1, $2, false)`
	_, err := p.db.ExecContext(ctx, query, p.config.RLSContextVarName, tenantID)
	if err != nil {
		return fmt.Errorf("failed to set RLS tenant context: %w", err)
	}

	p.currentTenant = &TenantContext{
		TenantID: tenantID,
		SetAt:    time.Now(),
	}

	// Track query statistics if enabled
	if p.config.EnableQueryStats {
		p.initializeQueryStats(tenantID)
	}

	log.Printf("### 🗄️ Database: Set RLS tenant context: %s", tenantID)
	return nil
}

// GetTenantContext returns the current tenant context
func (p *PostgreSQL) GetTenantContext(ctx context.Context) (TenantContext, error) {
	if !p.config.MultitenancyEnabled {
		return TenantContext{}, nil
	}

	p.tenantMu.RLock()
	defer p.tenantMu.RUnlock()

	if p.currentTenant != nil {
		// Check if context has expired
		if p.currentTenant.IsExpired() {
			log.Printf("Warning: tenant context expired for tenant %s", p.currentTenant.TenantID)
		}
		return *p.currentTenant, nil
	}

	// Try to get from database context
	query := `SELECT current_setting($1, true)`
	var tenantID string
	err := p.db.QueryRowContext(ctx, query, p.config.RLSContextVarName).Scan(&tenantID)
	if err != nil {
		return TenantContext{}, fmt.Errorf("failed to get RLS tenant context: %w", err)
	}

	if tenantID == "" {
		return TenantContext{}, nil
	}

	return TenantContext{
		TenantID: tenantID,
		SetAt:    time.Now(),
	}, nil
}

// ClearTenantContext clears the current tenant context
func (p *PostgreSQL) ClearTenantContext(ctx context.Context) error {
	if !p.config.MultitenancyEnabled {
		return nil
	}

	p.tenantMu.Lock()
	defer p.tenantMu.Unlock()

	// Clear RLS context variable
	query := `SELECT set_config($1, '', false)`
	_, err := p.db.ExecContext(ctx, query, p.config.RLSContextVarName)
	if err != nil {
		return fmt.Errorf("failed to clear RLS tenant context: %w", err)
	}

	p.currentTenant = nil

	log.Printf("### 🗄️ Database: Cleared RLS tenant context")
	return nil
}

// RLS Management methods

// EnableRLS enables Row Level Security on a table
func (p *PostgreSQL) EnableRLS(ctx context.Context, tableName string) error {
	if !p.config.MultitenancyEnabled {
		return fmt.Errorf("multitenancy is not enabled")
	}

	query := fmt.Sprintf(`ALTER TABLE %s ENABLE ROW LEVEL SECURITY`, tableName)

	ctx, cancel := context.WithTimeout(ctx, p.config.QueryTimeout)
	defer cancel()

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to enable RLS on table %s: %w", tableName, err)
	}

	log.Printf("### 🗄️ Database: Enabled RLS on table: %s", tableName)
	return nil
}

// CreateRLSPolicy creates a new RLS policy on a table
func (p *PostgreSQL) CreateRLSPolicy(ctx context.Context, tableName, policyName, policyDefinition string) error {
	if !p.config.MultitenancyEnabled {
		return fmt.Errorf("multitenancy is not enabled")
	}

	query := fmt.Sprintf(`CREATE POLICY %s ON %s %s`, policyName, tableName, policyDefinition)

	ctx, cancel := context.WithTimeout(ctx, p.config.QueryTimeout)
	defer cancel()

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create RLS policy %s on table %s: %w", policyName, tableName, err)
	}

	log.Printf("### 🗄️ Database: Created RLS policy %s on table: %s", policyName, tableName)
	return nil
}

// DropRLSPolicy drops an RLS policy from a table
func (p *PostgreSQL) DropRLSPolicy(ctx context.Context, tableName, policyName string) error {
	if !p.config.MultitenancyEnabled {
		return fmt.Errorf("multitenancy is not enabled")
	}

	query := fmt.Sprintf(`DROP POLICY IF EXISTS %s ON %s`, policyName, tableName)

	ctx, cancel := context.WithTimeout(ctx, p.config.QueryTimeout)
	defer cancel()

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop RLS policy %s from table %s: %w", policyName, tableName, err)
	}

	log.Printf("### 🗄️ Database: Dropped RLS policy %s from table: %s", policyName, tableName)
	return nil
}

// VerifyRLSIsolation verifies that RLS is working correctly for the current tenant
func (p *PostgreSQL) VerifyRLSIsolation(ctx context.Context, tableName string) error {
	if !p.config.MultitenancyEnabled {
		return fmt.Errorf("multitenancy is not enabled")
	}

	p.tenantMu.RLock()
	tenant := p.currentTenant
	p.tenantMu.RUnlock()

	if tenant == nil || tenant.TenantID == "" {
		return fmt.Errorf("no tenant context set")
	}

	// Test query to verify RLS is working
	testQuery := `SELECT COUNT(*) FROM ` + tableName + ` LIMIT 1`

	ctx, cancel := context.WithTimeout(ctx, p.config.QueryTimeout)
	defer cancel()

	var count int
	err := p.db.QueryRowContext(ctx, testQuery).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to verify RLS isolation: %w", err)
	}

	log.Printf("### 🗄️ Database: Verified RLS isolation for tenant %s on table %s", tenant.TenantID, tableName)
	return nil
}

// GetTenantQueryStats returns performance statistics for the current tenant
func (p *PostgreSQL) GetTenantQueryStats(ctx context.Context) (TenantQueryStats, error) {
	if !p.config.MultitenancyEnabled || !p.config.EnableQueryStats {
		return TenantQueryStats{}, fmt.Errorf("query statistics not enabled")
	}

	p.tenantMu.RLock()
	tenant := p.currentTenant
	p.tenantMu.RUnlock()

	if tenant == nil || tenant.TenantID == "" {
		return TenantQueryStats{}, fmt.Errorf("no tenant context set")
	}

	p.statsMu.RLock()
	stats, exists := p.queryStats[tenant.TenantID]
	p.statsMu.RUnlock()

	if !exists {
		return TenantQueryStats{
			TenantID:        tenant.TenantID,
			TotalQueries:    0,
			TotalDuration:   0,
			AverageDuration: 0,
			SlowQueries:     0,
			FailedQueries:   0,
			LastQueryAt:     time.Time{},
			TableStats:      make(map[string]int64),
			QueryTypes:      make(map[string]int64),
		}, nil
	}

	return *stats, nil
}

// Utility methods

// validateTenantID validates a tenant ID against the configured pattern
func (p *PostgreSQL) validateTenantID(tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}

	if len(tenantID) < 3 || len(tenantID) > 50 {
		return fmt.Errorf("tenant ID must be between 3 and 50 characters")
	}

	// Check against regex pattern if configured
	if p.config.TenantIDPattern != "" {
		matched, err := regexp.MatchString(p.config.TenantIDPattern, tenantID)
		if err != nil {
			return fmt.Errorf("failed to validate tenant ID pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("tenant ID '%s' does not match pattern '%s'", tenantID, p.config.TenantIDPattern)
		}
	}

	// Additional security checks
	if strings.Contains(tenantID, "..") || strings.Contains(tenantID, "--") {
		return fmt.Errorf("tenant ID contains invalid sequences")
	}

	return nil
}

// initializeQueryStats initializes query statistics tracking for a tenant
func (p *PostgreSQL) initializeQueryStats(tenantID string) {
	p.statsMu.Lock()
	defer p.statsMu.Unlock()

	if _, exists := p.queryStats[tenantID]; !exists {
		p.queryStats[tenantID] = &TenantQueryStats{
			TenantID:        tenantID,
			TotalQueries:    0,
			TotalDuration:   0,
			AverageDuration: 0,
			SlowQueries:     0,
			FailedQueries:   0,
			LastQueryAt:     time.Time{},
			TableStats:      make(map[string]int64),
			QueryTypes:      make(map[string]int64),
		}
	}
}

// updateQueryStats updates query statistics for the current tenant
func (p *PostgreSQL) updateQueryStats(tenantID string, duration time.Duration, queryType, tableName string,
	success bool) {
	if !p.config.EnableQueryStats {
		return
	}

	p.statsMu.Lock()
	defer p.statsMu.Unlock()

	stats, exists := p.queryStats[tenantID]
	if !exists {
		stats = &TenantQueryStats{
			TenantID:        tenantID,
			TotalQueries:    0,
			TotalDuration:   0,
			AverageDuration: 0,
			SlowQueries:     0,
			FailedQueries:   0,
			LastQueryAt:     time.Time{},
			TableStats:      make(map[string]int64),
			QueryTypes:      make(map[string]int64),
		}
		p.queryStats[tenantID] = stats
	}

	// Update statistics
	stats.TotalQueries++
	stats.TotalDuration += duration
	stats.AverageDuration = stats.TotalDuration / time.Duration(stats.TotalQueries)
	stats.LastQueryAt = time.Now()

	// Track slow queries (> 100ms)
	if duration > 100*time.Millisecond {
		stats.SlowQueries++
	}

	// Track failed queries
	if !success {
		stats.FailedQueries++
	}

	// Track table usage
	if tableName != "" {
		stats.TableStats[tableName]++
	}

	// Track query types
	if queryType != "" {
		stats.QueryTypes[queryType]++
	}
}
