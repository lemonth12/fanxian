package handler

import (
	"net/http"
	"strconv"

	"fanxian/internal/middleware"
	"fanxian/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	OrderService *service.OrderService
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 20
	offset := (page - 1) * limit

	orders, err := h.OrderService.GetUserOrders(userID, limit, offset)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "order/list.html", gin.H{
			"Title": "我的订单",
			"Error": "加载订单失败",
		})
		return
	}

	var earned float64
	u, err := h.OrderService.UserRepo.FindByID(userID)
	if err == nil {
		earned = u.TotalEarned
	}

	c.HTML(http.StatusOK, "order/list.html", gin.H{
		"Title":       "我的订单",
		"Orders":      orders,
		"TotalEarned": earned,
		"Page":        page,
	})
}
