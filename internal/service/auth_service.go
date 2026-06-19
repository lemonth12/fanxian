package service

import (
	"errors"
	"fmt"
	"time"

	"fanxian/internal/config"
	"fanxian/internal/jd"
	"fanxian/internal/model"
	"fanxian/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB       *gorm.DB
	UserRepo *repository.UserRepo
	Config   *config.Config
	JDClient *jd.Client
}

func (s *AuthService) Register(username, password string) (*model.User, error) {
	if len(password) < 6 {
		return nil, errors.New("密码至少6位")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := s.UserRepo.Create(u); err != nil {
		return nil, errors.New("用户名已存在")
	}

	subPID := fmt.Sprintf("%s_%d", s.Config.JDUnion.PID, u.ID)
	if s.JDClient != nil {
		if _, err := s.JDClient.CreateSubPID(s.Config.JDUnion.PID, subPID); err != nil {
			// Degrade gracefully: JD API call failure shouldn't block registration
		}
	}
	u.SubPID = subPID
	s.DB.Model(u).Update("sub_pid", subPID)

	return u, nil
}

func (s *AuthService) Login(username, password string) (*model.User, string, string, error) {
	u, err := s.UserRepo.FindByUsername(username)
	if err != nil {
		return nil, "", "", errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", errors.New("用户名或密码错误")
	}

	access, err := s.issueToken(u, s.Config.JWT.AccessExpire)
	if err != nil {
		return nil, "", "", fmt.Errorf("issue access token: %w", err)
	}
	refresh, err := s.issueToken(u, s.Config.JWT.RefreshExpire)
	if err != nil {
		return nil, "", "", fmt.Errorf("issue refresh token: %w", err)
	}
	return u, access, refresh, nil
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*model.JWTClaims, error) {
	return s.parseToken(tokenStr)
}

func (s *AuthService) issueToken(u *model.User, expire time.Duration) (string, error) {
	claims := model.JWTClaims{
		UserID:   u.ID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Config.JWT.Secret))
}

func (s *AuthService) parseToken(tokenStr string) (*model.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &model.JWTClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(s.Config.JWT.Secret), nil
		})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*model.JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
