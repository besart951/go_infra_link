package dto

// Pagination Query

type PaginationQuery struct {
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search string `form:"search"`
}

// User Filter Query with advanced filtering options
type UserFilterQuery struct {
	PaginationQuery
	Role        string `form:"role" binding:"omitempty,oneof=user admin superadmin"`
	IsActive    string `form:"is_active" binding:"omitempty,oneof=true false"`
	CompanyName string `form:"company_name"`
}

// Team Filter Query
type TeamFilterQuery struct {
	PaginationQuery
	// Can be extended with team-specific filters in the future
}

// Error Response

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
