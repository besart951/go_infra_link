package dto

import (
	"time"

	"github.com/google/uuid"
)

// User DTOs

type CreateUserRequest struct {
	FirstName   string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName    string     `json:"last_name" binding:"required,min=1,max=100"`
	Email       string     `json:"email" binding:"required,email"`
	Password    string     `json:"password" binding:"required,min=8"`
	IsActive    bool       `json:"is_active"`
	CreatedByID *uuid.UUID `json:"created_by_id"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email     string `json:"email" binding:"omitempty,email"`
	Password  string `json:"password" binding:"omitempty,min=8"`
	IsActive  *bool  `json:"is_active"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponse struct {
	Items      []UserResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
}
