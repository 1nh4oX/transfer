package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"transfer/backend/internal/config"
	"transfer/backend/internal/model"
)

var ErrInvalidCredential = errors.New("invalid credential")

type AuthService struct {
	cfg config.Config
}

func NewAuthService(cfg config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (s *AuthService) Login(req model.LoginRequest) (model.LoginResponse, error) {
	if req.Username != s.cfg.DemoUsername || req.Password != s.cfg.DemoPassword {
		return model.LoginResponse{}, ErrInvalidCredential
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(s.cfg.TokenExpireIn) * time.Second)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
	})

	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return model.LoginResponse{}, err
	}

	return model.LoginResponse{
		Token:     signed,
		TokenType: "Bearer",
		ExpiresIn: s.cfg.TokenExpireIn,
	}, nil
}
