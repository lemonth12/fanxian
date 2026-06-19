package handler

import (
	"net/http"

	"fanxian/internal/config"
	"fanxian/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService *service.AuthService
	Config      *config.Config
}

func (h *AuthHandler) Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	confirm := c.PostForm("confirm_password")

	if username == "" || password == "" {
		c.HTML(http.StatusOK, "auth/register.html", gin.H{
			"Error": "用户名和密码不能为空",
			"Title": "注册",
		})
		return
	}
	if password != confirm {
		c.HTML(http.StatusOK, "auth/register.html", gin.H{
			"Error": "两次密码不一致",
			"Title": "注册",
		})
		return
	}

	_, err := h.AuthService.Register(username, password)
	if err != nil {
		c.HTML(http.StatusOK, "auth/register.html", gin.H{
			"Error": err.Error(),
			"Title": "注册",
		})
		return
	}
	c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.HTML(http.StatusOK, "auth/login.html", gin.H{
			"Error": "用户名和密码不能为空",
			"Title": "登录",
		})
		return
	}

	_, access, refresh, err := h.AuthService.Login(username, password)
	if err != nil {
		c.HTML(http.StatusOK, "auth/login.html", gin.H{
			"Error": err.Error(),
			"Title": "登录",
		})
		return
	}

	c.SetCookie("access_token", access,
		int(h.Config.JWT.AccessExpire.Seconds()), "/", "", false, true)
	c.SetCookie("refresh_token", refresh,
		int(h.Config.JWT.RefreshExpire.Seconds()), "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/login.html", gin.H{
		"Title": "登录",
	})
}

func (h *AuthHandler) ShowRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/register.html", gin.H{
		"Title": "注册",
	})
}
