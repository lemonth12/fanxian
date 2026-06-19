package service

import (
	"errors"
	"fmt"
	"time"

	"fanxian/internal/config"
	"fanxian/internal/model"
	"fanxian/internal/repository"

	"gorm.io/gorm"
)

type OrderService struct {
	DB        *gorm.DB
	OrderRepo *repository.OrderRepo
	UserRepo  *repository.UserRepo
	Config    *config.Config
}

func (s *OrderService) GetUserOrders(userID uint, limit, offset int) ([]model.Order, error) {
	return s.OrderRepo.FindByUserID(userID, limit, offset)
}

func (s *OrderService) SettleOrder(jdOrderID string, commissionAmount, cashbackAmount float64) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var ord model.Order
		if err := tx.Where("jd_order_id = ?", jdOrderID).First(&ord).Error; err != nil {
			return fmt.Errorf("order not found: %w", err)
		}
		if ord.Status != "confirmed" {
			return errors.New("order not in confirmed status")
		}

		now := time.Now()
		if err := tx.Model(&model.Order{}).
			Where("jd_order_id = ?", jdOrderID).
			Updates(map[string]interface{}{
				"status":            "settled",
				"commission_amount": commissionAmount,
				"cashback_amount":   cashbackAmount,
				"settled_at":        &now,
			}).Error; err != nil {
			return err
		}

		if err := s.UserRepo.UpdateTotalEarned(tx, ord.UserID, cashbackAmount); err != nil {
			return err
		}
		return nil
	})
}

func (s *OrderService) UpsertOrder(o *model.Order, cashbackRate float64) error {
	if o.CashbackRate == 0 {
		o.CashbackRate = cashbackRate
	}
	return s.OrderRepo.Upsert(o)
}
