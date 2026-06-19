package handler

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"fanxian/internal/jd"
	"fanxian/internal/middleware"
	"fanxian/internal/repository"
	"fanxian/internal/service"

	"github.com/gin-gonic/gin"
)

func setupProductHandler(t *testing.T) (*ProductHandler, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, _ := repository.NewDB(":memory:", false)
	userRepo := &repository.UserRepo{DB: db}
	jdClient := jd.NewClient("key", "secret", "site", "pid")

	r := gin.New()
	tmpl := template.Must(template.New("product/convert.html").Parse("{{.Error}}{{.AffiliateURL}}{{.Estimate}}"))
	r.SetHTMLTemplate(tmpl)

	h := &ProductHandler{
		ProductService: &service.ProductService{
			DB: db, JDClient: jdClient, CashbackRate: 0.7,
		},
		UserRepo: userRepo,
	}
	r.GET("/", h.ShowConvert)
	r.POST("/convert", middleware.RateLimitMiddleware(10), h.Convert)
	return h, r
}

func TestShowConvert_NoAuth(t *testing.T) {
	_, r := setupProductHandler(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (public page)", w.Code)
	}
}

func TestConvert_NotLoggedIn(t *testing.T) {
	_, r := setupProductHandler(t)
	form := url.Values{"product_url": {"https://item.jd.com/123456.html"}}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/convert", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	if w.Code != 302 {
		t.Errorf("status = %d, want 302 redirect to login", w.Code)
	}
}
