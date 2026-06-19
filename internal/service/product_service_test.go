package service

import "testing"

func TestProductService_CalculateEstimate(t *testing.T) {
	svc := &ProductService{CashbackRate: 0.7}
	est := svc.CalculateEstimate(100.0, 0.05)
	// commission = 100 * 0.05 = 5.0, cashback = 5.0 * 0.7 = 3.5
	if est != 3.5 {
		t.Errorf("estimate = %f, want 3.5", est)
	}
}

func TestProductService_CalculateEstimate_Zero(t *testing.T) {
	svc := &ProductService{CashbackRate: 0.7}
	est := svc.CalculateEstimate(0, 0.05)
	if est != 0 {
		t.Errorf("estimate = %f, want 0", est)
	}
}
