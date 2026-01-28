package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"0xdomainsnapshot/internal/database"
)

// SyncLock manages exclusive access for sync operations
// Uses both in-memory mutex (for same-process) and database record (for multi-instance)
type SyncLock struct {
	db       *database.DB
	locks    map[string]*sync.Mutex
	mapMutex sync.RWMutex
}

// NewSyncLock creates a new SyncLock
func NewSyncLock(db *database.DB) *SyncLock {
	return &SyncLock{
		db:    db,
		locks: make(map[string]*sync.Mutex),
	}
}

// getLock returns the mutex for a collector, creating it if needed
func (s *SyncLock) getLock(collectorName string) *sync.Mutex {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	if _, exists := s.locks[collectorName]; !exists {
		s.locks[collectorName] = &sync.Mutex{}
	}
	return s.locks[collectorName]
}

// TryAcquire attempts to acquire a lock for the collector
// Returns (syncID, acquired, error)
// - syncID: ID of the sync_status record (use for Release)
// - acquired: true if lock was acquired, false if already running
// - error: any error that occurred
func (s *SyncLock) TryAcquire(ctx context.Context, collectorName, serviceType, triggerType string) (string, bool, error) {
	lock := s.getLock(collectorName)

	// Try to acquire in-memory lock (non-blocking)
	acquired := lock.TryLock()
	if !acquired {
		return "", false, nil
	}

	// Check database for running sync (handle multi-instance case)
	var existingID string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM sync_status
		WHERE collector_name = $1 AND status = 'running'
	`, collectorName).Scan(&existingID)

	if err == nil {
		// Another instance is running
		lock.Unlock()
		return "", false, nil
	} else if err != sql.ErrNoRows {
		lock.Unlock()
		return "", false, fmt.Errorf("check running sync: %w", err)
	}

	// Create new sync record
	var syncID string
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO sync_status (collector_name, service_type, status, trigger_type, started_at)
		VALUES ($1, $2, 'running', $3, NOW())
		RETURNING id
	`, collectorName, serviceType, triggerType).Scan(&syncID)

	if err != nil {
		lock.Unlock()
		return "", false, fmt.Errorf("create sync record: %w", err)
	}

	return syncID, true, nil
}

// Release releases the lock and updates sync status
func (s *SyncLock) Release(ctx context.Context, collectorName, syncID string, stats SyncReleaseStats, syncError error) error {
	lock := s.getLock(collectorName)
	defer lock.Unlock()

	status := "completed"
	var errMsg *string
	if syncError != nil {
		status = "failed"
		msg := syncError.Error()
		errMsg = &msg
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE sync_status
		SET status = $1, completed_at = NOW(),
		    records_found = $2, records_added = $3,
		    records_updated = $4, records_removed = $5,
		    error_message = $6
		WHERE id = $7
	`, status, stats.Found, stats.Added, stats.Updated, stats.Removed, errMsg, syncID)

	if err != nil {
		return fmt.Errorf("update sync status: %w", err)
	}

	return nil
}

// SyncReleaseStats holds stats for releasing a sync lock
type SyncReleaseStats struct {
	Found   int
	Added   int
	Updated int
	Removed int
}

// IsRunning checks if a collector is currently running
func (s *SyncLock) IsRunning(ctx context.Context, collectorName string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sync_status
		WHERE collector_name = $1 AND status = 'running'
	`, collectorName).Scan(&count)

	return count > 0, err
}

// GetStatus returns the latest status for all collectors
func (s *SyncLock) GetStatus(ctx context.Context) ([]CollectorStatusInfo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT ON (collector_name)
			collector_name, service_type, status, trigger_type,
			started_at, completed_at,
			records_found, records_added, records_updated, records_removed,
			error_message
		FROM sync_status
		ORDER BY collector_name, started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []CollectorStatusInfo
	for rows.Next() {
		var s CollectorStatusInfo
		var completedAt sql.NullTime
		var errMsg sql.NullString
		var found, added, updated, removed sql.NullInt64

		err := rows.Scan(
			&s.Name, &s.ServiceType, &s.Status, &s.TriggerType,
			&s.StartedAt, &completedAt,
			&found, &added, &updated, &removed,
			&errMsg,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			s.CompletedAt = &completedAt.Time
		}
		if errMsg.Valid {
			s.ErrorMessage = errMsg.String
		}
		if found.Valid {
			s.RecordsFound = int(found.Int64)
		}
		if added.Valid {
			s.RecordsAdded = int(added.Int64)
		}
		if updated.Valid {
			s.RecordsUpdated = int(updated.Int64)
		}
		if removed.Valid {
			s.RecordsRemoved = int(removed.Int64)
		}

		statuses = append(statuses, s)
	}

	return statuses, rows.Err()
}

// GetCollectorStatus returns the latest status for a specific collector
func (s *SyncLock) GetCollectorStatus(ctx context.Context, collectorName string) (*CollectorStatusInfo, error) {
	var status CollectorStatusInfo
	var completedAt sql.NullTime
	var errMsg sql.NullString
	var found, added, updated, removed sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT collector_name, service_type, status, trigger_type,
		       started_at, completed_at,
		       records_found, records_added, records_updated, records_removed,
		       error_message
		FROM sync_status
		WHERE collector_name = $1
		ORDER BY started_at DESC
		LIMIT 1
	`, collectorName).Scan(
		&status.Name, &status.ServiceType, &status.Status, &status.TriggerType,
		&status.StartedAt, &completedAt,
		&found, &added, &updated, &removed,
		&errMsg,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		status.CompletedAt = &completedAt.Time
	}
	if errMsg.Valid {
		status.ErrorMessage = errMsg.String
	}
	if found.Valid {
		status.RecordsFound = int(found.Int64)
	}
	if added.Valid {
		status.RecordsAdded = int(added.Int64)
	}
	if updated.Valid {
		status.RecordsUpdated = int(updated.Int64)
	}
	if removed.Valid {
		status.RecordsRemoved = int(removed.Int64)
	}

	return &status, nil
}

// CollectorStatusInfo holds status information for a collector
type CollectorStatusInfo struct {
	Name           string     `json:"name"`
	ServiceType    string     `json:"service_type"`
	Status         string     `json:"status"`
	TriggerType    string     `json:"trigger_type"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	RecordsFound   int        `json:"records_found"`
	RecordsAdded   int        `json:"records_added"`
	RecordsUpdated int        `json:"records_updated"`
	RecordsRemoved int        `json:"records_removed"`
	ErrorMessage   string     `json:"error_message,omitempty"`
}

// CleanupStale marks any stale "running" records as failed
// This handles cases where the process crashed without releasing the lock
func (s *SyncLock) CleanupStale(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge)

	result, err := s.db.ExecContext(ctx, `
		UPDATE sync_status
		SET status = 'failed', completed_at = NOW(), error_message = 'Process terminated unexpectedly'
		WHERE status = 'running' AND started_at < $1
	`, cutoff)

	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}
