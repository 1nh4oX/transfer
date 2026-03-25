package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"transfer/backend/internal/model"
	"transfer/backend/internal/repo"
)

type NoteService struct {
	repo repo.NoteRepository
}

func NewNoteService(noteRepo repo.NoteRepository) *NoteService {
	return &NoteService{repo: noteRepo}
}

func (s *NoteService) CreateNote(ctx context.Context, ownerID, content string) (model.NoteItem, error) {
	id, err := generateNoteID()
	if err != nil {
		return model.NoteItem{}, err
	}

	content = strings.TrimSpace(content)
	createdAt := time.Now().UTC()
	if err := s.repo.CreateNote(ctx, repo.CreateNoteParams{
		ID:        id,
		OwnerID:   ownerID,
		Content:   content,
		CreatedAt: createdAt,
	}); err != nil {
		return model.NoteItem{}, err
	}

	return model.NoteItem{ID: id, Content: content, CreatedAt: createdAt}, nil
}

func (s *NoteService) ListNotes(ctx context.Context, ownerID string, page, pageSize int) (model.ListNotesResponse, error) {
	offset := (page - 1) * pageSize
	recs, total, err := s.repo.ListNotesByOwner(ctx, ownerID, pageSize, offset)
	if err != nil {
		return model.ListNotesResponse{}, err
	}

	items := make([]model.NoteItem, 0, len(recs))
	for _, rec := range recs {
		items = append(items, model.NoteItem{
			ID:        rec.ID,
			Content:   rec.Content,
			CreatedAt: rec.CreatedAt,
		})
	}

	return model.ListNotesResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func generateNoteID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate note id: %w", err)
	}
	return "n_" + base64.RawURLEncoding.EncodeToString(b), nil
}
