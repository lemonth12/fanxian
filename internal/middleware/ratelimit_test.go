package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimit_UnderLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(10))
	r.POST("/convert", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		if userID == nil {
			userID = uint(0)
		}
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/convert", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRateLimit_OverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(2))
	r.POST("/convert", func(c *gin.Context) {
		c.String(200, "ok")
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/convert", nil)
		r.ServeHTTP(w, req)
		if i >= 2 && w.Code != 429 {
			t.Errorf("request %d: status = %d, want 429", i+1, w.Code)
		}
	}
}
