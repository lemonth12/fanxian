package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	SubPID       string    `gorm:"column:sub_pid;not null"`
	TotalEarned  float64   `gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
