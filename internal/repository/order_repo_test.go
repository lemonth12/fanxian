package repository

import (
	"fanxian/internal/model"
	"testing"
	"time"

	"gorm.io/gorm"
)

func setupOrderRepo(t *testing.T) (*OrderRepo, *gorm.DB) {
	t.Helper()
	db, err := NewDB(":memory:", false)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	return &OrderRepo{DB: db}, db
}

func TestOrderRepo_Upsert(t *testing.T) {
	repo, db := setupOrderRepo(t)
	db.Create(&model.User{
		Username: "u1", PasswordHash: "h", SubPID: "pid_1"})

	orderTime := time.Now()
	o := &model.Order{
		UserID:       1,
		SubPID:       "pid_1",
		JDOrderID:    "JD-001",
		ProductName:  "Test Product",
		CashbackRate: 0.7,
		Status:       "pending",
		OrderTime:    orderTime,
	}
	if err := repo.Upsert(o); err != nil {
		t.Fatalf("Upsert error: %v", err)
	}
	if o.ID == 0 {
		t.Error("ID should be set")
	}

	o.CommissionAmount = 50.0
	o.CashbackAmount = 35.0
	o.Status = "confirmed"
	if err := repo.Upsert(o); err != nil {
		t.Fatalf("Upsert update error: %v", err)
	}
	var found model.Order
	db.First(&found, o.ID)
	if found.Status != "confirmed" {
		t.Errorf("status = %s, want confirmed", found.Status)
	}
}

func TestOrderRepo_FindByUserID(t *testing.T) {
	repo, db := setupOrderRepo(t)
	db.Create(&model.User{
		Username: "u1", PasswordHash: "h", SubPID: "pid_1"})
	repo.Upsert(&model.Order{
		UserID: 1, SubPID: "pid_1", JDOrderID: "A",
		CashbackRate: 0.7, OrderTime: time.Now()})
	repo.Upsert(&model.Order{
		UserID: 1, SubPID: "pid_1", JDOrderID: "B",
		CashbackRate: 0.7, OrderTime: time.Now()})

	orders, err := repo.FindByUserID(1, 10, 0)
	if err != nil {
		t.Fatalf("FindByUserID error: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("count = %d, want 2", len(orders))
	}
}

func TestOrderRepo_FindByJDOrderID(t *testing.T) {
	repo, db := setupOrderRepo(t)
	db.Create(&model.User{
		Username: "u1", PasswordHash: "h", SubPID: "pid_1"})
	repo.Upsert(&model.Order{
		UserID: 1, SubPID: "pid_1", JDOrderID: "JD-UNIQUE",
		CashbackRate: 0.7, OrderTime: time.Now()})
	o, err := repo.FindByJDOrderID("JD-UNIQUE")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if o.JDOrderID != "JD-UNIQUE" {
		t.Errorf("jd_order_id = %s", o.JDOrderID)
	}
}
