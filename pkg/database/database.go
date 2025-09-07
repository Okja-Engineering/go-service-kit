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

	// RLS Multitenancy support - simple tenant context switching
	SetTenantContext(ctx context.Context, tenantID string) error
	ClearTenantContext(ctx context.Context) error
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

// TenantContext holds tenant-specific information for RLS multitenancy
type TenantContext struct {
	TenantID string    `json:"tenantID"`
	SetAt    time.Time `json:"setAt,omitempty"`
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

	// RLS Multitenancy configuration
	RLSContextVarName string // Default: "app.current_tenant_id"
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

		// RLS Multitenancy defaults
		RLSContextVarName: "app.current_tenant_id",
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
func WithMaxOpenConns(maxOpenConns int) Option {
	return func(c *Config) {
		c.MaxOpenConns = maxOpenConns
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *Config) {
		c.MaxIdleConns = maxIdleConns
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection
func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = connMaxLifetime
	}
}

// WithConnMaxIdleTime sets the maximum idle time of a connection
func WithConnMaxIdleTime(connMaxIdleTime time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxIdleTime = connMaxIdleTime
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(connectTimeout time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = connectTimeout
	}
}

// WithQueryTimeout sets the query timeout
func WithQueryTimeout(queryTimeout time.Duration) Option {
	return func(c *Config) {
		c.QueryTimeout = queryTimeout
	}
}

// WithRLSContextVarName sets the RLS context variable name
func WithRLSContextVarName(varName string) Option {
	return func(c *Config) {
		c.RLSContextVarName = varName
	}
}

// NewConfig creates a new configuration with the provided options
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

	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.config.ConnectTimeout)
	defer cancel()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(p.config.MaxOpenConns)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(p.config.ConnMaxIdleTime)

	p.db = db
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

	stats := p.db.Stats()
	return ConnectionStats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

// SetTenantContext sets the tenant context for RLS
func (p *PostgreSQL) SetTenantContext(ctx context.Context, tenantID string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed || p.db == nil {
		return fmt.Errorf("database connection is closed")
	}

	if tenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}

	// Set RLS context variable
	query := `SELECT set_config($1, $2, false)`
	_, err := p.db.ExecContext(ctx, query, p.config.RLSContextVarName, tenantID)
	if err != nil {
		return fmt.Errorf("failed to set RLS tenant context: %w", err)
	}

	return nil
}

// ClearTenantContext clears the tenant context
func (p *PostgreSQL) ClearTenantContext(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed || p.db == nil {
		return fmt.Errorf("database connection is closed")
	}

	// Clear RLS context variable
	query := `SELECT set_config($1, '', false)`
	_, err := p.db.ExecContext(ctx, query, p.config.RLSContextVarName)
	if err != nil {
		return fmt.Errorf("failed to clear RLS tenant context: %w", err)
	}

	return nil
}

// buildDSN builds the PostgreSQL connection string
func (p *PostgreSQL) buildDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host, p.config.Port, p.config.User, p.config.Password,
		p.config.Database, p.config.SSLMode)
}

// NewPostgreSQLWithOptions creates a new PostgreSQL instance with options
func NewPostgreSQLWithOptions(options ...Option) *PostgreSQL {
	config := NewConfig(options...)
	return NewPostgreSQL(config)
}
