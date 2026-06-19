package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fanxian/internal/config"
	"fanxian/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test",
			AccessExpire:  2 * time.Hour,
			RefreshExpire: 168 * time.Hour,
		},
	}
	r.Use(AuthMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 302 {
		t.Errorf("status = %d, want 302 redirect", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:        "testsecret",
			AccessExpire:  2 * time.Hour,
			RefreshExpire: 168 * time.Hour,
		},
	}
	r.Use(AuthMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(200, gin.H{"user_id": userID})
	})

	claims := model.JWTClaims{
		UserID:   42,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("testsecret"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: tokenStr,
	})
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
