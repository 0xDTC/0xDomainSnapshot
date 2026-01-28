-- 001_initial_schema.up.sql
-- Initial database schema for 0xDomainSnapshot

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Domains table (registered domains from registrars)
CREATE TABLE IF NOT EXISTS domains (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain          VARCHAR(255) NOT NULL,
    registrar       VARCHAR(50) NOT NULL,  -- 'GoDaddy', 'Cloudflare'
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- 'active', 'removed'
    expiry_date     DATE,
    discovery_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    last_seen       DATE NOT NULL DEFAULT CURRENT_DATE,
    raw_data        JSONB,  -- Store original API response
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(domain, registrar)
);

-- DNS Records table (subdomains/records)
CREATE TABLE IF NOT EXISTS dns_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain          VARCHAR(255) NOT NULL,  -- Parent domain
    subdomain       VARCHAR(255) NOT NULL DEFAULT '',  -- Empty string for root (@)
    record_type     VARCHAR(20) NOT NULL,  -- A, AAAA, CNAME, MX, TXT, NS, etc.
    data            TEXT NOT NULL,  -- Record value
    ttl             INTEGER,
    priority        INTEGER,  -- For MX records
    source          VARCHAR(50) NOT NULL,  -- 'GoDaddy', 'Cloudflare'
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- 'active', 'removed'
    discovery_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    last_seen       DATE NOT NULL DEFAULT CURRENT_DATE,
    raw_data        JSONB,  -- Store original API response
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Unique constraint on the record signature
    UNIQUE(domain, subdomain, record_type, data, source)
);

-- Sync status table (tracks collector runs and provides locking)
CREATE TABLE IF NOT EXISTS sync_status (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collector_name  VARCHAR(100) NOT NULL,  -- 'godaddy_dns', 'cloudflare_dns'
    service_type    VARCHAR(50) NOT NULL,   -- 'domains', 'dns_records'
    status          VARCHAR(20) NOT NULL,   -- 'running', 'completed', 'failed'
    started_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMP WITH TIME ZONE,
    records_found   INTEGER DEFAULT 0,
    records_added   INTEGER DEFAULT 0,
    records_updated INTEGER DEFAULT 0,
    records_removed INTEGER DEFAULT 0,
    error_message   TEXT,
    trigger_type    VARCHAR(20) NOT NULL    -- 'scheduled', 'manual'
);

-- Indexes for domains table
CREATE INDEX IF NOT EXISTS idx_domains_status ON domains(status);
CREATE INDEX IF NOT EXISTS idx_domains_registrar ON domains(registrar);
CREATE INDEX IF NOT EXISTS idx_domains_last_seen ON domains(last_seen);
CREATE INDEX IF NOT EXISTS idx_domains_domain ON domains(domain);

-- Indexes for dns_records table
CREATE INDEX IF NOT EXISTS idx_dns_records_domain ON dns_records(domain);
CREATE INDEX IF NOT EXISTS idx_dns_records_status ON dns_records(status);
CREATE INDEX IF NOT EXISTS idx_dns_records_source ON dns_records(source);
CREATE INDEX IF NOT EXISTS idx_dns_records_type ON dns_records(record_type);
CREATE INDEX IF NOT EXISTS idx_dns_records_last_seen ON dns_records(last_seen);
CREATE INDEX IF NOT EXISTS idx_dns_records_subdomain ON dns_records(subdomain);

-- Indexes for sync_status table
CREATE INDEX IF NOT EXISTS idx_sync_status_collector ON sync_status(collector_name);
CREATE INDEX IF NOT EXISTS idx_sync_status_status ON sync_status(status);
CREATE INDEX IF NOT EXISTS idx_sync_status_started ON sync_status(started_at DESC);

-- Partial index for finding running syncs (used for locking)
CREATE UNIQUE INDEX IF NOT EXISTS idx_sync_running
    ON sync_status(collector_name)
    WHERE status = 'running';
