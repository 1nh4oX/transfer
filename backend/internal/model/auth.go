package model

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
	ExpiresIn int64  `json:"expiresIn"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
