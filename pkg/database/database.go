package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
}

// NewPostgreSQL creates a new PostgreSQL database instance
func NewPostgreSQL(config *Config) *PostgreSQL {
	return &PostgreSQL{
		config: config,
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
