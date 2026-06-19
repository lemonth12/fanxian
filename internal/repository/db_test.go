package repository

import (
	"fanxian/internal/model"
	"testing"
)

func TestNewDB_AutoMigrate(t *testing.T) {
	db, err := NewDB(":memory:", true)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if !db.Migrator().HasTable(&model.User{}) {
		t.Error("users table not created")
	}
	if !db.Migrator().HasTable(&model.Order{}) {
		t.Error("orders table not created")
	}
}
