package model

import "time"

type Order struct {
	ID               uint       `gorm:"primaryKey"`
	UserID           uint       `gorm:"not null;index"`
	SubPID           string     `gorm:"column:sub_pid;not null"`
	JDOrderID        string     `gorm:"column:jd_order_id;uniqueIndex;not null"`
	SKUID            string     `gorm:"column:sku_id"`
	ProductName      string
	ProductURL       string
	EstimatedPrice   float64
	ActualPrice      float64
	JDCommissionRate float64    `gorm:"column:jd_commission_rate"`
	CommissionAmount float64    `gorm:"default:0"`
	CashbackAmount   float64    `gorm:"default:0"`
	CashbackRate     float64    `gorm:"not null"`
	Status           string     `gorm:"default:pending"`
	OrderTime        time.Time
	SettledAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
