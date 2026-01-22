package dto

// Pagination Query

type PaginationQuery struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Limit   int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search"`
	OrderBy string `form:"order_by"`
	Order   string `form:"order" binding:"omitempty,oneof=asc desc"`
}

// Error Response

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
