package repository

import (
	"fanxian/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	DB *gorm.DB
}

func (r *UserRepo) Create(u *model.User) error {
	return r.DB.Create(u).Error
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.DB.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdateTotalEarned(tx *gorm.DB, userID uint, amount float64) error {
	if tx == nil {
		tx = r.DB
	}
	return tx.Model(&model.User{}).Where("id = ?", userID).
		Update("total_earned", gorm.Expr("total_earned + ?", amount)).Error
}
