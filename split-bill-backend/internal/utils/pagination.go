package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination holds pagination parameters
type Pagination struct {
	Page     int   `json:"page"`
	Limit    int   `json:"limit"`
	Total    int64 `json:"total"`
	LastPage int   `json:"last_page"`
}

// DefaultPage is the default page number
const DefaultPage = 1

// DefaultLimit is the default number of items per page
const DefaultLimit = 20

// MaxLimit is the maximum allowed items per page
const MaxLimit = 100

// ParsePagination extracts pagination parameters from the request query
func ParsePagination(c *gin.Context) Pagination {
	page := DefaultPage
	limit := DefaultLimit

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > MaxLimit {
				limit = MaxLimit
			}
		}
	}

	return Pagination{
		Page:  page,
		Limit: limit,
	}
}

// Offset calculates the offset for database queries
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

// Skip is an alias for Offset (MongoDB uses skip)
func (p Pagination) Skip() int64 {
	return int64(p.Offset())
}

// SetTotal sets the total count and calculates last page
func (p *Pagination) SetTotal(total int64) {
	p.Total = total
	if total == 0 {
		p.LastPage = 1
	} else {
		p.LastPage = int((total + int64(p.Limit) - 1) / int64(p.Limit))
	}
}

// HasNext returns true if there's a next page
func (p Pagination) HasNext() bool {
	return p.Page < p.LastPage
}

// HasPrev returns true if there's a previous page
func (p Pagination) HasPrev() bool {
	return p.Page > 1
}

// PaginatedResponse wraps data with pagination metadata
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(data interface{}, pagination Pagination) PaginatedResponse {
	return PaginatedResponse{
		Data:       data,
		Pagination: pagination,
	}
}

// RespondPaginated sends a paginated JSON response
func RespondPaginated(c *gin.Context, statusCode int, message string, data interface{}, pagination Pagination) {
	c.JSON(statusCode, APIResponse{
		Success: statusCode >= 200 && statusCode < 300,
		Message: message,
		Data: PaginatedResponse{
			Data:       data,
			Pagination: pagination,
		},
	})
}
