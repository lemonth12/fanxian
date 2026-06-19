package service

import (
	"fanxian/internal/config"
	"fanxian/internal/model"
	"fanxian/internal/repository"
	"testing"
	"time"

	"gorm.io/gorm"
)

func setupOrderService(t *testing.T) (*OrderService, *gorm.DB) {
	t.Helper()
	db, _ := repository.NewDB(":memory:", false)
	orderRepo := &repository.OrderRepo{DB: db}
	userRepo := &repository.UserRepo{DB: db}
	cfg := &config.Config{
		Cashback: config.CashbackConfig{DefaultRate: 0.7},
	}
	return &OrderService{
		DB:        db,
		OrderRepo: orderRepo,
		UserRepo:  userRepo,
		Config:    cfg,
	}, db
}

func TestOrderService_SettleOrder(t *testing.T) {
	svc, db := setupOrderService(t)
	db.Create(&model.User{
		Username: "u1", PasswordHash: "h", SubPID: "p1"})
	db.Create(&model.Order{
		UserID: 1, SubPID: "p1", JDOrderID: "JD-SETTLE",
		CommissionAmount: 50.0, CashbackAmount: 35.0,
		CashbackRate: 0.7, Status: "confirmed",
		OrderTime: time.Now(),
	})

	err := svc.SettleOrder("JD-SETTLE", 50.0, 35.0)
	if err != nil {
		t.Fatalf("SettleOrder error: %v", err)
	}

	var ord model.Order
	db.Where("jd_order_id = ?", "JD-SETTLE").First(&ord)
	if ord.Status != "settled" {
		t.Errorf("status = %s, want settled", ord.Status)
	}
	if ord.CashbackAmount != 35.0 {
		t.Errorf("cashback = %f, want 35.0", ord.CashbackAmount)
	}

	var usr model.User
	db.First(&usr, 1)
	if usr.TotalEarned != 35.0 {
		t.Errorf("total_earned = %f, want 35.0", usr.TotalEarned)
	}
}

func TestOrderService_SettleOrder_NotFound(t *testing.T) {
	svc, _ := setupOrderService(t)
	err := svc.SettleOrder("NONEXISTENT", 50.0, 35.0)
	if err == nil {
		t.Error("expected error settling nonexistent order")
	}
}

func TestOrderService_GetUserOrders(t *testing.T) {
	svc, db := setupOrderService(t)
	db.Create(&model.User{
		Username: "u2", PasswordHash: "h", SubPID: "p2"})
	db.Create(&model.Order{
		UserID: 1, SubPID: "p2", JDOrderID: "A1",
		CashbackRate: 0.7, Status: "pending",
		OrderTime: time.Now()})
	db.Create(&model.Order{
		UserID: 1, SubPID: "p2", JDOrderID: "B1",
		CashbackRate: 0.7, Status: "confirmed",
		OrderTime: time.Now()})

	orders, err := svc.GetUserOrders(1, 10, 0)
	if err != nil {
		t.Fatalf("GetUserOrders error: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("count = %d, want 2", len(orders))
	}
}
