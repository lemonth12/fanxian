package middleware

import (
	"net/http"

	"fanxian/internal/config"
	"fanxian/internal/service"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	authSvc := &service.AuthService{Config: cfg}
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		claims, err := authSvc.ValidateAccessToken(tokenStr)
		if err != nil {
			c.SetCookie("access_token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	id, _ := c.Get("userID")
	return id.(uint)
}

func OptionalAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, _ := c.Cookie("access_token")
		if tokenStr == "" {
			c.Next()
			return
		}
		authSvc := &service.AuthService{Config: cfg}
		claims, err := authSvc.ValidateAccessToken(tokenStr)
		if err != nil {
			c.Next()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
