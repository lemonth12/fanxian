package handler

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fanxian/internal/config"
	"fanxian/internal/model"
	"fanxian/internal/repository"
	"fanxian/internal/service"

	"github.com/gin-gonic/gin"
)

func setupOrderHandler(t *testing.T) (*OrderHandler, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, _ := repository.NewDB(":memory:", false)
	orderRepo := &repository.OrderRepo{DB: db}
	userRepo := &repository.UserRepo{DB: db}
	cfg := &config.Config{
		Cashback: config.CashbackConfig{DefaultRate: 0.7},
	}
	orderSvc := &service.OrderService{
		DB: db, OrderRepo: orderRepo,
		UserRepo: userRepo, Config: cfg,
	}

	r := gin.New()
	tmpl := template.Must(template.New("order/list.html").Parse("{{.Title}}{{.Error}}{{range .Orders}}{{.ProductName}}{{end}}{{.TotalEarned}}"))
	r.SetHTMLTemplate(tmpl)

	h := &OrderHandler{OrderService: orderSvc}
	r.GET("/orders", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("username", "testuser")
		c.Next()
	}, h.ListOrders)
	return h, r
}

func TestListOrders_Empty(t *testing.T) {
	_, r := setupOrderHandler(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/orders", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestListOrders_WithData(t *testing.T) {
	h, r := setupOrderHandler(t)
	h.OrderService.UpsertOrder(&model.Order{
		UserID: 1, SubPID: "p1", JDOrderID: "JD-001",
		ProductName: "测试商品", CashbackRate: 0.7,
		Status: "pending", OrderTime: time.Now(),
	}, 0.7)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/orders", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
