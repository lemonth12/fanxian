package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCSRFMiddleware_GET_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CSRFMiddleware())
	r.GET("/safe", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/safe", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCSRFMiddleware_POST_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CSRFMiddleware())
	r.POST("/unsafe", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/unsafe", strings.NewReader("data"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestCSRFMiddleware_POST_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CSRFMiddleware())
	r.POST("/unsafe", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First GET to obtain CSRF cookie
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w1, req1)
	cookies := w1.Result().Cookies()
	var csrfToken string
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			csrfToken = c.Value
		}
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/unsafe", strings.NewReader("data"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
