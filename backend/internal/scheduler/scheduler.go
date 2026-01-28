package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"0xdomainsnapshot/internal/collector"
	"0xdomainsnapshot/internal/config"
	"0xdomainsnapshot/internal/service"
)

// Scheduler manages scheduled and on-demand sync jobs
type Scheduler struct {
	cron      *cron.Cron
	registry  *collector.Registry
	syncSvc   *service.SyncService
	exportSvc *service.ExportService
	lock      *SyncLock
	cfg       config.SchedulerConfig
	jobs      map[string]cron.EntryID
	mu        sync.Mutex
}

// New creates a new Scheduler
func New(
	registry *collector.Registry,
	syncSvc *service.SyncService,
	exportSvc *service.ExportService,
	lock *SyncLock,
	cfg config.SchedulerConfig,
) *Scheduler {
	return &Scheduler{
		cron:      cron.New(),
		registry:  registry,
		syncSvc:   syncSvc,
		exportSvc: exportSvc,
		lock:      lock,
		cfg:       cfg,
		jobs:      make(map[string]cron.EntryID),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	// Cleanup any stale locks from previous runs
	stale, err := s.lock.CleanupStale(ctx, 2*time.Hour)
	if err != nil {
		log.Printf("[Scheduler] Warning: failed to cleanup stale locks: %v", err)
	} else if stale > 0 {
		log.Printf("[Scheduler] Cleaned up %d stale sync records", stale)
	}

	if !s.cfg.Enabled {
		log.Println("[Scheduler] Scheduler disabled")
		return nil
	}

	// Schedule DNS collectors
	if s.cfg.DNSCron != "" {
		for _, c := range s.registry.GetByType(collector.CollectorTypeDNSRecords) {
			if err := s.scheduleCollector(c, s.cfg.DNSCron); err != nil {
				log.Printf("[Scheduler] Warning: failed to schedule %s: %v", c.Name(), err)
			}
		}
	}

	// Schedule domain collectors (if separate cron)
	if s.cfg.DomainsCron != "" && s.cfg.DomainsCron != s.cfg.DNSCron {
		for _, c := range s.registry.GetByType(collector.CollectorTypeDomains) {
			if err := s.scheduleCollector(c, s.cfg.DomainsCron); err != nil {
				log.Printf("[Scheduler] Warning: failed to schedule %s: %v", c.Name(), err)
			}
		}
	}

	s.cron.Start()
	log.Printf("[Scheduler] Started with %d scheduled jobs", len(s.jobs))

	// List scheduled jobs
	for name, entryID := range s.jobs {
		entry := s.cron.Entry(entryID)
		log.Printf("[Scheduler] %s: next run at %v", name, entry.Next)
	}

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("[Scheduler] Stopping...")
	cronCtx := s.cron.Stop()
	<-cronCtx.Done()
	log.Println("[Scheduler] Stopped")

	return nil
}

// scheduleCollector adds a collector to the cron scheduler
func (s *Scheduler) scheduleCollector(c collector.Collector, cronExpr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already scheduled
	if _, exists := s.jobs[c.Name()]; exists {
		return fmt.Errorf("collector %s already scheduled", c.Name())
	}

	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.runCollector(context.Background(), c, "scheduled")
	})

	if err != nil {
		return fmt.Errorf("add cron job: %w", err)
	}

	s.jobs[c.Name()] = entryID
	log.Printf("[Scheduler] Scheduled %s with cron: %s", c.Name(), cronExpr)

	return nil
}

// runCollector runs a collector with locking
func (s *Scheduler) runCollector(ctx context.Context, c collector.Collector, triggerType string) {
	// Try to acquire lock (non-blocking)
	syncID, acquired, err := s.lock.TryAcquire(ctx, c.Name(), string(c.Type()), triggerType)
	if err != nil {
		log.Printf("[Scheduler] Failed to acquire lock for %s: %v", c.Name(), err)
		return
	}

	if !acquired {
		log.Printf("[Scheduler] Skipping %s - already running", c.Name())
		return
	}

	log.Printf("[Scheduler] Starting %s sync (trigger: %s)", c.Name(), triggerType)

	// Run the sync
	stats, syncErr := s.syncSvc.RunCollector(ctx, c)

	// Prepare release stats
	releaseStats := SyncReleaseStats{}
	if stats != nil {
		releaseStats.Found = stats.Found
		releaseStats.Added = stats.Added
		releaseStats.Updated = stats.Updated
		releaseStats.Removed = stats.Removed
	}

	// Release lock with results
	if err := s.lock.Release(ctx, c.Name(), syncID, releaseStats, syncErr); err != nil {
		log.Printf("[Scheduler] Failed to release lock for %s: %v", c.Name(), err)
	}

	if syncErr != nil {
		log.Printf("[Scheduler] Sync %s failed: %v", c.Name(), syncErr)
		return
	}

	log.Printf("[Scheduler] Sync %s completed: found=%d added=%d updated=%d removed=%d",
		c.Name(), releaseStats.Found, releaseStats.Added, releaseStats.Updated, releaseStats.Removed)

	// Export JSON files after successful sync
	if err := s.exportSvc.ExportAll(ctx); err != nil {
		log.Printf("[Scheduler] Export failed after %s sync: %v", c.Name(), err)
	}
}

// TriggerSync manually triggers a collector sync (on-demand)
// Returns an error if the collector is not found
// Returns nil immediately - sync runs in background
func (s *Scheduler) TriggerSync(ctx context.Context, collectorName string) error {
	c, ok := s.registry.Get(collectorName)
	if !ok {
		return fmt.Errorf("collector not found: %s", collectorName)
	}

	// Run in background goroutine
	go s.runCollector(ctx, c, "manual")

	return nil
}

// TriggerSyncAll triggers all collectors
func (s *Scheduler) TriggerSyncAll(ctx context.Context) error {
	collectors := s.registry.All()
	if len(collectors) == 0 {
		return fmt.Errorf("no collectors registered")
	}

	for _, c := range collectors {
		go s.runCollector(ctx, c, "manual")
	}

	return nil
}

// GetNextRun returns the next scheduled run time for a collector
func (s *Scheduler) GetNextRun(collectorName string) *time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.jobs[collectorName]; ok {
		entry := s.cron.Entry(entryID)
		if !entry.Next.IsZero() {
			return &entry.Next
		}
	}
	return nil
}

// GetScheduledJobs returns information about all scheduled jobs
func (s *Scheduler) GetScheduledJobs() []ScheduledJobInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	var jobs []ScheduledJobInfo
	for name, entryID := range s.jobs {
		entry := s.cron.Entry(entryID)
		jobs = append(jobs, ScheduledJobInfo{
			Name:     name,
			NextRun:  entry.Next,
			PrevRun:  entry.Prev,
		})
	}
	return jobs
}

// ScheduledJobInfo holds information about a scheduled job
type ScheduledJobInfo struct {
	Name    string    `json:"name"`
	NextRun time.Time `json:"next_run"`
	PrevRun time.Time `json:"prev_run,omitempty"`
}

// IsCollectorRunning checks if a collector is currently running
func (s *Scheduler) IsCollectorRunning(ctx context.Context, collectorName string) (bool, error) {
	return s.lock.IsRunning(ctx, collectorName)
}

// GetCollectorStatus returns the status of a collector
func (s *Scheduler) GetCollectorStatus(ctx context.Context, collectorName string) (*CollectorStatusInfo, error) {
	return s.lock.GetCollectorStatus(ctx, collectorName)
}

// GetAllStatus returns the status of all collectors
func (s *Scheduler) GetAllStatus(ctx context.Context) ([]CollectorStatusInfo, error) {
	return s.lock.GetStatus(ctx)
}
