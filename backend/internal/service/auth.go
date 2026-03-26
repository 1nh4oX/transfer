package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"transfer/backend/internal/config"
	"transfer/backend/internal/model"
	"transfer/backend/internal/repo"
)

var ErrInvalidCredential = errors.New("invalid credential")
var ErrUsernameTaken = errors.New("username taken")

type AuthService struct {
	cfg      config.Config
	userRepo repo.UserRepository
}

func NewAuthService(cfg config.Config, userRepo repo.UserRepository) *AuthService {
	return &AuthService{cfg: cfg, userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (model.LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.LoginResponse{}, err
	}

	err = s.userRepo.CreateUser(ctx, repo.CreateUserParams{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, repo.ErrAlreadyExists) {
			return model.LoginResponse{}, ErrUsernameTaken
		}
		return model.LoginResponse{}, err
	}

	return s.issueToken(username)
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password

	rec, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		// fallback to demo account for backward compatibility
		if username == s.cfg.DemoUsername && password == s.cfg.DemoPassword {
			return s.issueToken(username)
		}
		if errors.Is(err, repo.ErrNotFound) {
			return model.LoginResponse{}, ErrInvalidCredential
		}
		return model.LoginResponse{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(rec.PasswordHash), []byte(password)) != nil {
		return model.LoginResponse{}, ErrInvalidCredential
	}
	return s.issueToken(username)
}

func (s *AuthService) issueToken(username string) (model.LoginResponse, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(s.cfg.TokenExpireIn) * time.Second)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
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
