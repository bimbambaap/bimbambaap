package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"` // omitempty = niet meesturen in publieke responses
	Password  string    `json:"-"`               // nooit meesturen
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Caption   string    `json:"caption"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	Author    *User     `json:"author,omitempty"`
}

// Request structs (wat binnenkomt van de client)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

type CreatePostRequest struct {
	Caption string `json:"caption"`
}
