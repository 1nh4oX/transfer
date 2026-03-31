package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"transfer/backend/internal/middleware"
	"transfer/backend/internal/model"
	"transfer/backend/internal/service"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) Upload(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "file is required."})
		return
	}

	var folderID *string
	if v := strings.TrimSpace(c.PostForm("folderId")); v != "" {
		folderID = &v
	}

	fileItem, err := h.fileService.Upload(c.Request.Context(), currentUser, fileHeader, folderID)
	if err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FOLDER_NOT_FOUND", Message: "Folder not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to upload file."})
		return
	}

	c.JSON(http.StatusCreated, fileItem)
}

func (h *FileHandler) DownloadByID(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	fileID := c.Param("fileId")
	fileItem, err := h.fileService.GetFileByOwner(c.Request.Context(), currentUser, fileID)
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FILE_NOT_FOUND", Message: "File not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Unexpected server error."})
		return
	}

	c.FileAttachment(fileItem.StoragePath, fileItem.Name)
}

func (h *FileHandler) DownloadPublic(c *gin.Context, fileID string) {
	fileItem, err := h.fileService.GetPublicFile(c.Request.Context(), fileID)
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FILE_NOT_FOUND", Message: "File not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Unexpected server error."})
		return
	}

	c.FileAttachment(fileItem.StoragePath, fileItem.Name)
}

func (h *FileHandler) List(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	page := 1
	if v := c.Query("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid page."})
			return
		}
		page = p
	}

	pageSize := 20
	if v := c.Query("pageSize"); v != "" {
		ps, err := strconv.Atoi(v)
		if err != nil || ps < 1 || ps > 100 {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid pageSize."})
			return
		}
		pageSize = ps
	}

	var folderID *string
	if v := strings.TrimSpace(c.Query("folderId")); v != "" {
		folderID = &v
	}

	resp, err := h.fileService.ListByOwner(c.Request.Context(), currentUser, folderID, page, pageSize)
	if err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FOLDER_NOT_FOUND", Message: "Folder not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to list files."})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) MoveByID(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	fileID := c.Param("fileId")
	var req model.MoveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid request."})
		return
	}

	item, err := h.fileService.MoveFileByOwner(c.Request.Context(), currentUser, fileID, req.TargetFolderID)
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FILE_NOT_FOUND", Message: "File not found."})
			return
		}
		if errors.Is(err, service.ErrFolderNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FOLDER_NOT_FOUND", Message: "Folder not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to move file."})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *FileHandler) CreateFolder(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	var req model.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid request."})
		return
	}

	item, err := h.fileService.CreateFolder(c.Request.Context(), currentUser, req.Name, req.ParentID)
	if err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FOLDER_NOT_FOUND", Message: "Folder not found."})
			return
		}
		if errors.Is(err, service.ErrFolderConflict) {
			c.JSON(http.StatusConflict, model.ErrorResponse{Code: "FOLDER_CONFLICT", Message: "Folder with same name already exists."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to create folder."})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *FileHandler) MoveFolder(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	folderID := c.Param("folderId")
	var req model.MoveFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid request."})
		return
	}

	item, err := h.fileService.MoveFolderByOwner(c.Request.Context(), currentUser, folderID, req.TargetParentID)
	if err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FOLDER_NOT_FOUND", Message: "Folder not found."})
			return
		}
		if errors.Is(err, service.ErrFolderConflict) {
			c.JSON(http.StatusConflict, model.ErrorResponse{Code: "FOLDER_CONFLICT", Message: "Folder with same name already exists."})
			return
		}
		if errors.Is(err, service.ErrInvalidFolderMove) {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_MOVE", Message: "Cannot move folder into itself or its descendant."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to move folder."})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *FileHandler) RenameFolder(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	folderID := c.Param("folderId")
	var req model.RenameFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid request."})
		return
	}

	item, err := h.fileService.RenameFolderByOwner(c.Request.Context(), currentUser, folderID, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrFolderNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FOLDER_NOT_FOUND", Message: "Folder not found."})
			return
		}
		if errors.Is(err, service.ErrFolderConflict) {
			c.JSON(http.StatusConflict, model.ErrorResponse{Code: "FOLDER_CONFLICT", Message: "Folder with same name already exists."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to rename folder."})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *FileHandler) GetTree(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	resp, err := h.fileService.GetTreeByOwner(c.Request.Context(), currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to load tree."})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *FileHandler) DeleteByID(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}
	fileID := c.Param("fileId")
	if err := h.fileService.DeleteByOwner(c.Request.Context(), currentUser, fileID); err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "FILE_NOT_FOUND", Message: "File not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to delete file."})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FileHandler) NotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, model.ErrorResponse{
		Code:    "NOT_IMPLEMENTED",
		Message: "This endpoint is planned but not implemented yet.",
	})
}
