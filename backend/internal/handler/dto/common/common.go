package common

// Pagination Query

type PaginationQuery struct {
	Page           int    `form:"page" binding:"omitempty,min=1"`
	Limit          int    `form:"limit" binding:"omitempty,min=1,max=1000"`
	Search         string `form:"search"`
	OrderBy        string `form:"order_by"`
	Order          string `form:"order" binding:"omitempty,oneof=asc desc"`
	IncludeDeleted bool   `form:"include_deleted"`
}

// Error Response

type ErrorResponse struct {
	// Error and Fields are kept as compatibility aliases for existing clients.
	Error        string               `json:"error"`
	Code         string               `json:"code,omitempty"`
	Message      string               `json:"message,omitempty"`
	LocalizedKey string               `json:"localized_key,omitempty"`
	Details      any                  `json:"details,omitempty"`
	Fields       map[string]string    `json:"fields,omitempty"`
	FieldErrors  []FieldErrorResponse `json:"field_errors,omitempty"`
	RequestID    string               `json:"request_id,omitempty"`
}

type FieldErrorResponse struct {
	Path         string `json:"path"`
	Code         string `json:"code,omitempty"`
	Message      string `json:"message"`
	LocalizedKey string `json:"localized_key,omitempty"`
}
