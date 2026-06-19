package service

import (
	"fanxian/internal/config"
	"fanxian/internal/repository"
	"testing"
	"time"
)

func TestCronService_LastSyncTime(t *testing.T) {
	svc := &CronService{}
	now := time.Now()
	svc.SetLastSyncTime(now)
	if svc.LastSyncTime().IsZero() {
		t.Error("last sync time should not be zero")
	}
}

func TestCronService_RunOnce(t *testing.T) {
	db, _ := repository.NewDB(":memory:", false)
	orderRepo := &repository.OrderRepo{DB: db}
	userRepo := &repository.UserRepo{DB: db}
	cfg := &config.Config{
		Cashback: config.CashbackConfig{DefaultRate: 0.7},
	}
	orderSvc := &OrderService{
		DB: db, OrderRepo: orderRepo,
		UserRepo: userRepo, Config: cfg,
	}
	svc := &CronService{
		OrderService: orderSvc,
	}

	now := time.Now()
	svc.SetLastSyncTime(now.Add(-1 * time.Hour))

	err := svc.RunOnce()
	if err != nil {
		t.Fatalf("RunOnce unexpected error: %v", err)
	}
	if svc.LastSyncTime().Before(now) {
		t.Error("sync time should have advanced")
	}
}
