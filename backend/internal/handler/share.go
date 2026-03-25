package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"transfer/backend/internal/middleware"
	"transfer/backend/internal/model"
	"transfer/backend/internal/service"
)

type ShareHandler struct {
	shareService *service.ShareService
	fileHandler  *FileHandler
}

func NewShareHandler(shareService *service.ShareService, fileHandler *FileHandler) *ShareHandler {
	return &ShareHandler{shareService: shareService, fileHandler: fileHandler}
}

// CreateShare handles POST /api/shares (requires JWT).
func (h *ShareHandler) CreateShare(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	var req model.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: err.Error()})
		return
	}

	info, err := h.shareService.CreateShare(c.Request.Context(), currentUser, req)
	if err != nil {
		if errors.Is(err, service.ErrItemNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "NOT_FOUND", Message: "Item not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to create share."})
		return
	}

	c.JSON(http.StatusCreated, info)
}

// GetShareInfo handles GET /api/s/:token (public, no auth needed).
func (h *ShareHandler) GetShareInfo(c *gin.Context) {
	token := c.Param("token")

	meta, err := h.shareService.GetShare(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrShareNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "SHARE_NOT_FOUND", Message: "Share link does not exist."})
			return
		}
		if errors.Is(err, service.ErrShareExpired) {
			c.JSON(http.StatusGone, model.ErrorResponse{Code: "SHARE_EXPIRED", Message: "This share link has expired."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Unexpected server error."})
		return
	}

	c.JSON(http.StatusOK, meta)
}

// DownloadShare handles GET /api/s/:token/download (public, no auth needed).
func (h *ShareHandler) DownloadShare(c *gin.Context) {
	token := c.Param("token")

	meta, err := h.shareService.GetShare(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrShareNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "SHARE_NOT_FOUND", Message: "Share link does not exist."})
			return
		}
		if errors.Is(err, service.ErrShareExpired) {
			c.JSON(http.StatusGone, model.ErrorResponse{Code: "SHARE_EXPIRED", Message: "This share link has expired."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Unexpected server error."})
		return
	}

	// Folder download: not implemented in phase 1.
	if meta.ItemType == "folder" {
		c.JSON(http.StatusNotImplemented, model.ErrorResponse{
			Code:    "NOT_IMPLEMENTED",
			Message: "Folder download (zip) is not yet supported.",
		})
		return
	}

	h.fileHandler.DownloadPublic(c, meta.ItemId)
}
