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

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrFolderNotFound    = errors.New("folder not found")
	ErrFolderConflict    = errors.New("folder name conflict")
	ErrInvalidFolderMove = errors.New("invalid folder move")
)

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
	if folderID != nil {
		if _, err := s.repo.GetFolderByIDAndOwner(ctx, *folderID, ownerID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return model.FileItem{}, ErrFolderNotFound
			}
			return model.FileItem{}, err
		}
	}

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
	if folderID != nil {
		if _, err := s.repo.GetFolderByIDAndOwner(ctx, *folderID, ownerID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return model.ListFilesResponse{}, ErrFolderNotFound
			}
			return model.ListFilesResponse{}, err
		}
	}

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

func (s *FileService) MoveFileByOwner(ctx context.Context, ownerID, fileID string, targetFolderID *string) (model.FileItem, error) {
	if targetFolderID != nil {
		if _, err := s.repo.GetFolderByIDAndOwner(ctx, *targetFolderID, ownerID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return model.FileItem{}, ErrFolderNotFound
			}
			return model.FileItem{}, err
		}
	}

	rec, err := s.repo.MoveFileByIDAndOwner(ctx, fileID, ownerID, targetFolderID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.FileItem{}, ErrFileNotFound
		}
		return model.FileItem{}, err
	}
	return toModelFile(rec), nil
}

func (s *FileService) CreateFolder(ctx context.Context, ownerID, name string, parentID *string) (model.FolderItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.FolderItem{}, errors.New("folder name is required")
	}
	if parentID != nil {
		if _, err := s.repo.GetFolderByIDAndOwner(ctx, *parentID, ownerID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return model.FolderItem{}, ErrFolderNotFound
			}
			return model.FolderItem{}, err
		}
	}

	exists, err := s.repo.FolderNameExistsInParent(ctx, ownerID, parentID, name, nil)
	if err != nil {
		return model.FolderItem{}, err
	}
	if exists {
		return model.FolderItem{}, ErrFolderConflict
	}

	folderID, err := generateID("d")
	if err != nil {
		return model.FolderItem{}, err
	}
	createdAt := time.Now().UTC()
	if err := s.repo.CreateFolder(ctx, repo.CreateFolderParams{
		ID:        folderID,
		OwnerID:   ownerID,
		ParentID:  parentID,
		Name:      name,
		CreatedAt: createdAt,
	}); err != nil {
		return model.FolderItem{}, err
	}
	return model.FolderItem{ID: folderID, Name: name, ParentID: parentID, CreatedAt: createdAt}, nil
}

func (s *FileService) MoveFolderByOwner(ctx context.Context, ownerID, folderID string, targetParentID *string) (model.FolderItem, error) {
	current, err := s.repo.GetFolderByIDAndOwner(ctx, folderID, ownerID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.FolderItem{}, ErrFolderNotFound
		}
		return model.FolderItem{}, err
	}

	if targetParentID != nil {
		target, err := s.repo.GetFolderByIDAndOwner(ctx, *targetParentID, ownerID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return model.FolderItem{}, ErrFolderNotFound
			}
			return model.FolderItem{}, err
		}
		if target.ID == folderID {
			return model.FolderItem{}, ErrInvalidFolderMove
		}
		folders, err := s.repo.ListFoldersByOwner(ctx, ownerID)
		if err != nil {
			return model.FolderItem{}, err
		}
		if isDescendant(folders, folderID, target.ID) {
			return model.FolderItem{}, ErrInvalidFolderMove
		}
	}

	exists, err := s.repo.FolderNameExistsInParent(ctx, ownerID, targetParentID, current.Name, &folderID)
	if err != nil {
		return model.FolderItem{}, err
	}
	if exists {
		return model.FolderItem{}, ErrFolderConflict
	}

	rec, err := s.repo.MoveFolderByIDAndOwner(ctx, folderID, ownerID, targetParentID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.FolderItem{}, ErrFolderNotFound
		}
		return model.FolderItem{}, err
	}
	return toModelFolder(rec), nil
}

func (s *FileService) RenameFolderByOwner(ctx context.Context, ownerID, folderID, name string) (model.FolderItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.FolderItem{}, errors.New("folder name is required")
	}

	current, err := s.repo.GetFolderByIDAndOwner(ctx, folderID, ownerID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.FolderItem{}, ErrFolderNotFound
		}
		return model.FolderItem{}, err
	}

	exists, err := s.repo.FolderNameExistsInParent(ctx, ownerID, current.ParentID, name, &folderID)
	if err != nil {
		return model.FolderItem{}, err
	}
	if exists {
		return model.FolderItem{}, ErrFolderConflict
	}

	rec, err := s.repo.RenameFolderByIDAndOwner(ctx, folderID, ownerID, name)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return model.FolderItem{}, ErrFolderNotFound
		}
		return model.FolderItem{}, err
	}
	return toModelFolder(rec), nil
}

func (s *FileService) GetTreeByOwner(ctx context.Context, ownerID string) (model.TreeResponse, error) {
	folders, err := s.repo.ListFoldersByOwner(ctx, ownerID)
	if err != nil {
		return model.TreeResponse{}, err
	}
	files, err := s.repo.ListAllFilesByOwner(ctx, ownerID)
	if err != nil {
		return model.TreeResponse{}, err
	}

	folderByID := make(map[string]repo.FolderRecord, len(folders))
	children := make(map[string][]repo.FolderRecord)
	rootFolders := make([]repo.FolderRecord, 0)
	for _, f := range folders {
		folderByID[f.ID] = f
		if f.ParentID == nil {
			rootFolders = append(rootFolders, f)
			continue
		}
		children[*f.ParentID] = append(children[*f.ParentID], f)
	}

	filesByFolder := make(map[string][]model.FileItem)
	rootFiles := make([]model.FileItem, 0)
	for _, rec := range files {
		item := toModelFile(rec)
		if rec.FolderID == nil {
			rootFiles = append(rootFiles, item)
			continue
		}
		if _, ok := folderByID[*rec.FolderID]; !ok {
			rootFiles = append(rootFiles, item)
			continue
		}
		filesByFolder[*rec.FolderID] = append(filesByFolder[*rec.FolderID], item)
	}

	_ = rootFiles // reserved for future root files in contract if needed

	var build func(repo.FolderRecord) model.TreeNode
	build = func(folder repo.FolderRecord) model.TreeNode {
		node := model.TreeNode{
			Folder:   toModelFolder(folder),
			Children: make([]model.TreeNode, 0),
			Files:    filesByFolder[folder.ID],
		}
		for _, child := range children[folder.ID] {
			node.Children = append(node.Children, build(child))
		}
		return node
	}

	resp := model.TreeResponse{RootFolders: make([]model.TreeNode, 0, len(rootFolders))}
	for _, root := range rootFolders {
		resp.RootFolders = append(resp.RootFolders, build(root))
	}
	return resp, nil
}

func isDescendant(folders []repo.FolderRecord, ancestorID, maybeDescendantID string) bool {
	parentByID := make(map[string]*string, len(folders))
	for _, f := range folders {
		parentByID[f.ID] = f.ParentID
	}
	current := maybeDescendantID
	for {
		parent, ok := parentByID[current]
		if !ok || parent == nil {
			return false
		}
		if *parent == ancestorID {
			return true
		}
		current = *parent
	}
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

func toModelFolder(rec repo.FolderRecord) model.FolderItem {
	return model.FolderItem{
		ID:        rec.ID,
		Name:      rec.Name,
		ParentID:  rec.ParentID,
		CreatedAt: rec.CreatedAt,
	}
}

func generateID(prefix string) (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(b), nil
}
