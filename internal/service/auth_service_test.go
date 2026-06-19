package service

import (
	"fanxian/internal/config"
	"fanxian/internal/repository"
	"strings"
	"testing"
	"time"
)

func setupAuthService(t *testing.T) *AuthService {
	t.Helper()
	db, err := repository.NewDB(":memory:", false)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	return &AuthService{
		DB:       db,
		UserRepo: &repository.UserRepo{DB: db},
		Config: &config.Config{
			JWT: config.JWTConfig{
				AccessExpire:  2 * time.Hour,
				RefreshExpire: 168 * time.Hour,
				Secret:        "testsecret",
			},
			JDUnion: config.JDUnionConfig{PID: "jd_test"},
		},
	}
}

func TestAuthService_Register(t *testing.T) {
	svc := setupAuthService(t)
	u, err := svc.Register("alice", "secret123")
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if u.ID == 0 {
		t.Error("user ID should be set")
	}
	if u.SubPID != "jd_test_1" {
		t.Errorf("sub_pid = %s, want jd_test_1", u.SubPID)
	}
	if !strings.HasPrefix(u.PasswordHash, "$2a$") {
		t.Error("password should be bcrypt hashed")
	}
}

func TestAuthService_Register_Duplicate(t *testing.T) {
	svc := setupAuthService(t)
	svc.Register("bob", "password")
	_, err := svc.Register("bob", "another")
	if err == nil {
		t.Error("expected duplicate error")
	}
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	svc := setupAuthService(t)
	_, err := svc.Register("eve", "12345")
	if err == nil {
		t.Error("expected password too short error")
	}
}

func TestAuthService_Login(t *testing.T) {
	svc := setupAuthService(t)
	svc.Register("carol", "mypassword")
	u, access, refresh, err := svc.Login("carol", "mypassword")
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if u.Username != "carol" {
		t.Errorf("username = %s", u.Username)
	}
	if access == "" || refresh == "" {
		t.Error("tokens should not be empty")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc := setupAuthService(t)
	svc.Register("dave", "correct")
	_, _, _, err := svc.Login("dave", "wrong")
	if err == nil {
		t.Error("expected wrong password error")
	}
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	svc := setupAuthService(t)
	svc.Register("tokenuser", "password")
	_, access, _, _ := svc.Login("tokenuser", "password")
	claims, err := svc.ValidateAccessToken(access)
	if err != nil {
		t.Fatalf("ValidateAccessToken error: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("userID = %d, want 1", claims.UserID)
	}
	if claims.Username != "tokenuser" {
		t.Errorf("username = %s, want tokenuser", claims.Username)
	}
}
