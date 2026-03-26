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

type FolderItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parentId"`
	CreatedAt time.Time `json:"createdAt"`
}

type TreeNode struct {
	Folder   FolderItem `json:"folder"`
	Children []TreeNode `json:"children"`
	Files    []FileItem `json:"files"`
}

type TreeResponse struct {
	RootFolders []TreeNode `json:"rootFolders"`
}

type CreateFolderRequest struct {
	Name     string  `json:"name" binding:"required,min=1,max=128"`
	ParentID *string `json:"parentId"`
}

type MoveFolderRequest struct {
	TargetParentID *string `json:"targetParentId"`
}

type MoveFileRequest struct {
	TargetFolderID *string `json:"targetFolderId"`
}

type ListFilesResponse struct {
	Items    []FileItem `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int64      `json:"total"`
}
