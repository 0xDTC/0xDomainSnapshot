-- 001_initial_schema.down.sql
-- Rollback initial database schema

DROP INDEX IF EXISTS idx_sync_running;
DROP INDEX IF EXISTS idx_sync_status_started;
DROP INDEX IF EXISTS idx_sync_status_status;
DROP INDEX IF EXISTS idx_sync_status_collector;

DROP INDEX IF EXISTS idx_dns_records_subdomain;
DROP INDEX IF EXISTS idx_dns_records_last_seen;
DROP INDEX IF EXISTS idx_dns_records_type;
DROP INDEX IF EXISTS idx_dns_records_source;
DROP INDEX IF EXISTS idx_dns_records_status;
DROP INDEX IF EXISTS idx_dns_records_domain;

DROP INDEX IF EXISTS idx_domains_domain;
DROP INDEX IF EXISTS idx_domains_last_seen;
DROP INDEX IF EXISTS idx_domains_registrar;
DROP INDEX IF EXISTS idx_domains_status;

DROP TABLE IF EXISTS sync_status;
DROP TABLE IF EXISTS dns_records;
DROP TABLE IF EXISTS domains;
