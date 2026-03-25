package model

import "time"

type CreateNoteRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

type NoteItem struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type ListNotesResponse struct {
	Items    []NoteItem `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int64      `json:"total"`
}
