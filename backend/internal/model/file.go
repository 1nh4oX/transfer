package model

import "time"

type FileItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mimeType"`
	Ext         string    `json:"ext"`
	FolderID    *string   `json:"folderId"`
	CreatedAt   time.Time `json:"createdAt"`
	DownloadURL string    `json:"downloadUrl"`
	StoragePath string    `json:"-"`
	OwnerID     string    `json:"-"`
}

type ListFilesResponse struct {
	Items    []FileItem `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int64      `json:"total"`
}
