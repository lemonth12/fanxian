package repository

import (
	"fanxian/internal/model"
	"testing"

	"gorm.io/gorm"
)

func setupUserRepo(t *testing.T) (*UserRepo, *gorm.DB) {
	t.Helper()
	db, err := NewDB(":memory:", false)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	return &UserRepo{DB: db}, db
}

func TestUserRepo_Create(t *testing.T) {
	repo, _ := setupUserRepo(t)
	u := &model.User{
		Username:     "testuser",
		PasswordHash: "hashed",
		SubPID:       "pid_1",
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if u.ID == 0 {
		t.Error("ID should be set after create")
	}
}

func TestUserRepo_Create_Duplicate(t *testing.T) {
	repo, _ := setupUserRepo(t)
	repo.Create(&model.User{
		Username: "alice", PasswordHash: "h", SubPID: "p_1"})
	err := repo.Create(&model.User{
		Username: "alice", PasswordHash: "h2", SubPID: "p_2"})
	if err == nil {
		t.Error("expected unique constraint error")
	}
}

func TestUserRepo_FindByUsername(t *testing.T) {
	repo, _ := setupUserRepo(t)
	repo.Create(&model.User{
		Username: "bob", PasswordHash: "hash", SubPID: "p"})
	u, err := repo.FindByUsername("bob")
	if err != nil {
		t.Fatalf("FindByUsername error: %v", err)
	}
	if u.Username != "bob" {
		t.Errorf("username = %s, want bob", u.Username)
	}
}

func TestUserRepo_FindByUsername_NotFound(t *testing.T) {
	repo, _ := setupUserRepo(t)
	_, err := repo.FindByUsername("nobody")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUserRepo_UpdateTotalEarned(t *testing.T) {
	repo, db := setupUserRepo(t)
	repo.Create(&model.User{
		Username: "earn", PasswordHash: "h", SubPID: "p",
		TotalEarned: 10.0})
	if err := repo.UpdateTotalEarned(db, 1, 5.0); err != nil {
		t.Fatalf("UpdateTotalEarned error: %v", err)
	}
	u, _ := repo.FindByUsername("earn")
	if u.TotalEarned != 15.0 {
		t.Errorf("total = %f, want 15.0", u.TotalEarned)
	}
}
