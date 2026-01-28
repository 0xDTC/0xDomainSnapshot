package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"0xdomainsnapshot/internal/config"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// RunMigrations runs the database migrations
func (db *DB) RunMigrations(ctx context.Context) error {
	// Check if tables exist
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'domains'
		)
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if tables exist: %w", err)
	}

	if exists {
		return nil // Tables already exist
	}

	// Run migration
	_, err = db.ExecContext(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// migrationSQL contains the initial database schema
const migrationSQL = `
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Domains table (registered domains from registrars)
CREATE TABLE IF NOT EXISTS domains (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain          VARCHAR(255) NOT NULL,
    registrar       VARCHAR(50) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    expiry_date     DATE,
    discovery_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    last_seen       DATE NOT NULL DEFAULT CURRENT_DATE,
    raw_data        JSONB,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(domain, registrar)
);

-- DNS Records table (subdomains/records)
CREATE TABLE IF NOT EXISTS dns_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain          VARCHAR(255) NOT NULL,
    subdomain       VARCHAR(255) NOT NULL DEFAULT '',
    record_type     VARCHAR(20) NOT NULL,
    data            TEXT NOT NULL,
    ttl             INTEGER,
    priority        INTEGER,
    source          VARCHAR(50) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    discovery_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    last_seen       DATE NOT NULL DEFAULT CURRENT_DATE,
    raw_data        JSONB,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(domain, subdomain, record_type, data, source)
);

-- Sync status table (tracks collector runs)
CREATE TABLE IF NOT EXISTS sync_status (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collector_name  VARCHAR(100) NOT NULL,
    service_type    VARCHAR(50) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    started_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMP WITH TIME ZONE,
    records_found   INTEGER DEFAULT 0,
    records_added   INTEGER DEFAULT 0,
    records_updated INTEGER DEFAULT 0,
    records_removed INTEGER DEFAULT 0,
    error_message   TEXT,
    trigger_type    VARCHAR(20) NOT NULL
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_domains_status ON domains(status);
CREATE INDEX IF NOT EXISTS idx_domains_registrar ON domains(registrar);
CREATE INDEX IF NOT EXISTS idx_domains_last_seen ON domains(last_seen);

CREATE INDEX IF NOT EXISTS idx_dns_records_domain ON dns_records(domain);
CREATE INDEX IF NOT EXISTS idx_dns_records_status ON dns_records(status);
CREATE INDEX IF NOT EXISTS idx_dns_records_source ON dns_records(source);
CREATE INDEX IF NOT EXISTS idx_dns_records_type ON dns_records(record_type);
CREATE INDEX IF NOT EXISTS idx_dns_records_last_seen ON dns_records(last_seen);

CREATE INDEX IF NOT EXISTS idx_sync_status_collector ON sync_status(collector_name);
CREATE INDEX IF NOT EXISTS idx_sync_status_status ON sync_status(status);
CREATE INDEX IF NOT EXISTS idx_sync_status_started ON sync_status(started_at DESC);
`
