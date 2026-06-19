package service

import (
	"log"
	"sync"
	"time"
)

type CronService struct {
	OrderService *OrderService
	JDClient     interface{} // JD client for order sync, injected in main
	mu           sync.Mutex
	lastSyncTime time.Time
}

func (s *CronService) LastSyncTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSyncTime
}

func (s *CronService) SetLastSyncTime(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSyncTime = t
}

func (s *CronService) RunOnce() error {
	lastSync := s.LastSyncTime()
	now := time.Now()
	if lastSync.IsZero() {
		lastSync = now.Add(-1 * time.Hour)
	}

	log.Printf("[cron] syncing orders from %s to %s",
		lastSync.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"))

	s.SetLastSyncTime(now)
	return nil
}

func (s *CronService) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		log.Printf("[cron] started with interval %s", interval)
		for range ticker.C {
			if err := s.RunOnce(); err != nil {
				log.Printf("[cron] sync error: %v", err)
			}
		}
	}()
}
