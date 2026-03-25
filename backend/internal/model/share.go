package model

import "time"

// CreateShareRequest is sent by the authenticated user to create a share link.
type CreateShareRequest struct {
	ItemType  string     `json:"itemType" binding:"required,oneof=file folder"`
	ItemId    string     `json:"itemId" binding:"required"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// ShareInfo is the response body after successfully creating a share.
type ShareInfo struct {
	Token     string     `json:"token"`
	ShareURL  string     `json:"shareUrl"`
	ItemType  string     `json:"itemType"`
	ItemId    string     `json:"itemId"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// ShareMeta is the public metadata returned when anybody visits /api/s/:token.
type ShareMeta struct {
	Token     string     `json:"token"`
	ItemType  string     `json:"itemType"`
	Name      string     `json:"name"`
	Size      *int64     `json:"size"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
	// Internal field — not serialised, used by download handler.
	ItemId string `json:"-"`
}
