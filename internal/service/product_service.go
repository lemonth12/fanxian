package service

import (
	"fanxian/internal/jd"

	"gorm.io/gorm"
)

type ProductService struct {
	DB            *gorm.DB
	JDClient      *jd.Client
	CashbackRate  float64
	MinCommission float64
}

func (s *ProductService) CalculateEstimate(price float64, commissionRate float64) float64 {
	if commissionRate == 0 {
		commissionRate = s.MinCommission
	}
	commission := price * commissionRate
	return commission * s.CashbackRate
}

func (s *ProductService) ConvertLink(subPID, productURL string) (string, float64, error) {
	sku, err := jd.ParseURL(productURL)
	if err != nil {
		return "", 0, err
	}

	affiliateURL, err := s.JDClient.ConvertLink(subPID, sku)
	if err != nil {
		return "", 0, err
	}

	estimate := s.CalculateEstimate(0, 0.03)
	return affiliateURL, estimate, nil
}
