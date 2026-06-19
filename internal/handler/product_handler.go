package handler

import (
	"net/http"

	"fanxian/internal/repository"
	"fanxian/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	ProductService *service.ProductService
	UserRepo       *repository.UserRepo
}

func (h *ProductHandler) ShowConvert(c *gin.Context) {
	c.HTML(http.StatusOK, "product/convert.html", gin.H{
		"Title": "京东返现 - 链接转换",
	})
}

func (h *ProductHandler) Convert(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	productURL := c.PostForm("product_url")
	if productURL == "" {
		c.HTML(http.StatusOK, "product/convert.html", gin.H{
			"Title": "京东返现 - 链接转换",
			"Error": "请粘贴京东商品链接",
		})
		return
	}

	uid := userID.(uint)
	u, err := h.UserRepo.FindByID(uid)
	if err != nil {
		c.HTML(http.StatusOK, "product/convert.html", gin.H{
			"Title": "京东返现 - 链接转换",
			"Error": "用户信息获取失败",
		})
		return
	}

	affiliateURL, estimate, err := h.ProductService.ConvertLink(u.SubPID, productURL)
	if err != nil {
		c.HTML(http.StatusOK, "product/convert.html", gin.H{
			"Title": "京东返现 - 链接转换",
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "product/convert.html", gin.H{
		"Title":        "京东返现 - 链接转换",
		"AffiliateURL": affiliateURL,
		"Estimate":     estimate,
		"ProductURL":   productURL,
	})
}
