package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"fanxian/internal/config"
	"fanxian/internal/handler"
	"fanxian/internal/jd"
	"fanxian/internal/middleware"
	"fanxian/internal/repository"
	"fanxian/internal/service"

	"github.com/gin-gonic/gin"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := repository.NewDB(cfg.Database.Path, cfg.Database.WAL)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	userRepo := &repository.UserRepo{DB: db}
	orderRepo := &repository.OrderRepo{DB: db}

	jdClient := jd.NewClient(
		cfg.JDUnion.AppKey,
		cfg.JDUnion.AppSecret,
		cfg.JDUnion.SiteID,
		cfg.JDUnion.PID,
	)

	authSvc := &service.AuthService{
		DB: db, UserRepo: userRepo, Config: cfg, JDClient: jdClient,
	}
	orderSvc := &service.OrderService{
		DB: db, OrderRepo: orderRepo, UserRepo: userRepo, Config: cfg,
	}
	productSvc := &service.ProductService{
		DB: db, JDClient: jdClient, CashbackRate: cfg.Cashback.DefaultRate,
	}
	cronSvc := &service.CronService{
		OrderService: orderSvc,
	}

	authH := &handler.AuthHandler{AuthService: authSvc, Config: cfg}
	productH := &handler.ProductHandler{ProductService: productSvc, UserRepo: userRepo}
	orderH := &handler.OrderHandler{OrderService: orderSvc}

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Templates
	templatesSub, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatalf("templates sub: %v", err)
	}
	tmpl := template.New("").Funcs(template.FuncMap{
		"statusText": func(status string) string {
			switch status {
			case "pending":
				return "待确认"
			case "confirmed":
				return "已确认"
			case "settled":
				return "已结算"
			case "invalid":
				return "已失效"
			default:
				return status
			}
		},
	})
	tmpl = template.Must(tmpl.ParseFS(templatesSub,
		"layout.html",
		"auth/login.html",
		"auth/register.html",
		"product/convert.html",
		"order/list.html",
	))
	r.SetHTMLTemplate(tmpl)

	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticSub))

	// CSRF protection
	csrf := middleware.CSRFMiddleware()

	// Public routes
	r.GET("/login", csrf, authH.ShowLogin)
	r.GET("/register", csrf, authH.ShowRegister)
	r.POST("/register", csrf, authH.Register)
	r.POST("/login", csrf, authH.Login)

	// Authenticated routes
	auth := r.Group("")
	auth.Use(middleware.AuthMiddleware(cfg))
	auth.GET("/", productH.ShowConvert)
	auth.POST("/convert", middleware.RateLimitMiddleware(cfg.RateLimit.ConvertPerMinute), productH.Convert)
	auth.GET("/orders", orderH.ListOrders)
	auth.GET("/logout", authH.Logout)

	// Start cron
	cronSvc.Start(1 * time.Hour)

	log.Printf("server starting on :%d", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
