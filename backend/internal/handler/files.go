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
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to list files."})
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
