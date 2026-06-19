package repository

import (
	"fanxian/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepo struct {
	DB *gorm.DB
}

func (r *OrderRepo) Upsert(o *model.Order) error {
	return r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "jd_order_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"sub_pid", "sku_id", "product_name", "product_url",
			"estimated_price", "actual_price", "jd_commission_rate",
			"commission_amount", "cashback_amount", "cashback_rate",
			"status", "order_time", "settled_at", "updated_at",
		}),
	}).Create(o).Error
}

func (r *OrderRepo) FindByUserID(userID uint, limit, offset int) ([]model.Order, error) {
	var orders []model.Order
	err := r.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepo) FindByJDOrderID(jdOrderID string) (*model.Order, error) {
	var o model.Order
	err := r.DB.Where("jd_order_id = ?", jdOrderID).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) FindByStatus(status string) ([]model.Order, error) {
	var orders []model.Order
	err := r.DB.Where("status = ?", status).Find(&orders).Error
	return orders, err
}
