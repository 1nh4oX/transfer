package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"transfer/backend/internal/config"
	"transfer/backend/internal/model"
	"transfer/backend/internal/repo"
)

var (
	ErrShareNotFound = errors.New("share not found")
	ErrShareExpired  = errors.New("share expired")
	ErrItemNotFound  = errors.New("item not found")
)

type ShareService struct {
	baseURL string
	repo    repo.ShareRepository
	files   repo.FileRepository
}

func NewShareService(cfg config.Config, repo repo.ShareRepository, files repo.FileRepository) *ShareService {
	return &ShareService{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		repo:    repo,
		files:   files,
	}
}

// CreateShare generates and persists a new share token for a file or folder.
func (s *ShareService) CreateShare(ctx context.Context, ownerID string, req model.CreateShareRequest) (model.ShareInfo, error) {
	now := time.Now().UTC()
	itemName := req.ItemId
	var itemSize *int64

	if req.ItemType == "file" {
		rec, err := s.files.GetFileByIDAndOwner(ctx, req.ItemId, ownerID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return model.ShareInfo{}, ErrItemNotFound
			}
			return model.ShareInfo{}, err
		}
		itemName = rec.Name
		size := rec.Size
		itemSize = &size
	}

	var token string
	for i := 0; i < 3; i++ {
		generated, err := generateToken()
		if err != nil {
			return model.ShareInfo{}, fmt.Errorf("generate token: %w", err)
		}
		token = generated

		err = s.repo.CreateShare(ctx, repo.CreateShareParams{
			Token:     token,
			OwnerID:   ownerID,
			ItemType:  req.ItemType,
			ItemID:    req.ItemId,
			ItemName:  itemName,
			ItemSize:  itemSize,
			CreatedAt: now,
			ExpiresAt: req.ExpiresAt,
		})
		if err == nil {
			return model.ShareInfo{
				Token:     token,
				ShareURL:  fmt.Sprintf("%s/api/s/%s", s.baseURL, token),
				ItemType:  req.ItemType,
				ItemId:    req.ItemId,
				CreatedAt: now,
				ExpiresAt: req.ExpiresAt,
			}, nil
		}

		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			continue
		}
		return model.ShareInfo{}, err
	}

	return model.ShareInfo{}, errors.New("failed to allocate unique share token")
}

// GetShare looks up a token from persistent storage and returns the public metadata.
func (s *ShareService) GetShare(ctx context.Context, token string) (model.ShareMeta, error) {
	rec, err := s.repo.GetShareByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.ShareMeta{}, ErrShareNotFound
		}
		return model.ShareMeta{}, err
	}

	if rec.ExpiresAt != nil && time.Now().After(*rec.ExpiresAt) {
		return model.ShareMeta{}, ErrShareExpired
	}

	return model.ShareMeta{
		Token:     rec.Token,
		ItemType:  rec.ItemType,
		Name:      rec.ItemName,
		Size:      rec.ItemSize,
		CreatedAt: rec.CreatedAt,
		ExpiresAt: rec.ExpiresAt,
		ItemId:    rec.ItemID,
	}, nil
}

// generateToken creates a 16-byte random URL-safe base64 token (~22 chars).
func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
