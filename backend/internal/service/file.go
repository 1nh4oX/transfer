package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"transfer/backend/internal/model"
	"transfer/backend/internal/repo"
)

var ErrFileNotFound = errors.New("file not found")

type FileService struct {
	repo      repo.FileRepository
	uploadDir string
}

func NewFileService(fileRepo repo.FileRepository, uploadDir string) (*FileService, error) {
	if strings.TrimSpace(uploadDir) == "" {
		uploadDir = "uploads"
	}
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &FileService{repo: fileRepo, uploadDir: uploadDir}, nil
}

func (s *FileService) Upload(ctx context.Context, ownerID string, fileHeader *multipart.FileHeader, folderID *string) (model.FileItem, error) {
	fileID, err := generateID("f")
	if err != nil {
		return model.FileItem{}, err
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fileHeader.Filename))
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	safeName := filepath.Base(fileHeader.Filename)
	storedName := fileID + "_" + safeName
	storedPath := filepath.Join(s.uploadDir, storedName)

	src, err := fileHeader.Open()
	if err != nil {
		return model.FileItem{}, fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(storedPath)
	if err != nil {
		return model.FileItem{}, fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	n, err := dst.ReadFrom(src)
	if err != nil {
		return model.FileItem{}, fmt.Errorf("write destination file: %w", err)
	}

	createdAt := time.Now().UTC()
	if err := s.repo.CreateFile(ctx, repo.CreateFileParams{
		ID:          fileID,
		OwnerID:     ownerID,
		FolderID:    folderID,
		Name:        safeName,
		StoragePath: storedPath,
		MimeType:    mimeType,
		Ext:         ext,
		Size:        n,
		CreatedAt:   createdAt,
	}); err != nil {
		_ = os.Remove(storedPath)
		return model.FileItem{}, err
	}

	return model.FileItem{
		ID:          fileID,
		Name:        safeName,
		Size:        n,
		MimeType:    mimeType,
		Ext:         ext,
		FolderID:    folderID,
		CreatedAt:   createdAt,
		DownloadURL: "/api/files/" + fileID + "/download",
		StoragePath: storedPath,
		OwnerID:     ownerID,
	}, nil
}

func (s *FileService) GetFileByOwner(ctx context.Context, ownerID, fileID string) (model.FileItem, error) {
	rec, err := s.repo.GetFileByIDAndOwner(ctx, fileID, ownerID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.FileItem{}, ErrFileNotFound
		}
		return model.FileItem{}, err
	}
	return toModelFile(rec), nil
}

func (s *FileService) GetPublicFile(ctx context.Context, fileID string) (model.FileItem, error) {
	rec, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.FileItem{}, ErrFileNotFound
		}
		return model.FileItem{}, err
	}
	return toModelFile(rec), nil
}

func (s *FileService) ListByOwner(ctx context.Context, ownerID string, folderID *string, page, pageSize int) (model.ListFilesResponse, error) {
	offset := (page - 1) * pageSize
	recs, total, err := s.repo.ListFilesByOwner(ctx, ownerID, folderID, pageSize, offset)
	if err != nil {
		return model.ListFilesResponse{}, err
	}

	items := make([]model.FileItem, 0, len(recs))
	for _, rec := range recs {
		items = append(items, toModelFile(rec))
	}

	return model.ListFilesResponse{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *FileService) DeleteByOwner(ctx context.Context, ownerID, fileID string) error {
	rec, err := s.repo.DeleteFileByIDAndOwner(ctx, fileID, ownerID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrFileNotFound
		}
		return err
	}

	if err := os.Remove(rec.StoragePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove file from disk: %w", err)
	}
	return nil
}

func toModelFile(rec repo.FileRecord) model.FileItem {
	return model.FileItem{
		ID:          rec.ID,
		Name:        rec.Name,
		Size:        rec.Size,
		MimeType:    rec.MimeType,
		Ext:         rec.Ext,
		FolderID:    rec.FolderID,
		CreatedAt:   rec.CreatedAt,
		DownloadURL: "/api/files/" + rec.ID + "/download",
		StoragePath: rec.StoragePath,
		OwnerID:     rec.OwnerID,
	}
}

func generateID(prefix string) (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(b), nil
}
