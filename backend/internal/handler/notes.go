package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"transfer/backend/internal/middleware"
	"transfer/backend/internal/model"
	"transfer/backend/internal/service"
)

type NoteHandler struct {
	noteService *service.NoteService
}

func NewNoteHandler(noteService *service.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

func (h *NoteHandler) CreateNote(c *gin.Context) {
	currentUser, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: "UNAUTHORIZED", Message: "Authentication required."})
		return
	}

	var req model.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "Invalid request."})
		return
	}

	item, err := h.noteService.CreateNote(c.Request.Context(), currentUser, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to create note."})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *NoteHandler) ListNotes(c *gin.Context) {
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

	resp, err := h.noteService.ListNotes(c.Request.Context(), currentUser, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "Failed to list notes."})
		return
	}

	c.JSON(http.StatusOK, resp)
}
