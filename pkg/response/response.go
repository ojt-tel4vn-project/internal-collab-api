package response

import "time"

type APIResponse[T any] struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      T      `json:"data,omitempty"`
	Timestamp string `json:"timestamp"`
}

type PaginatedData[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       []T          `json:"data"`
	Pagination *Pagination  `json:"pagination"`
	Timestamp  string       `json:"timestamp"`
}

type Pagination struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func Success[T any](message string, data T) *APIResponse[T] {
	return &APIResponse[T]{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func SuccessMessage(message string) *APIResponse[any] {
	return &APIResponse[any]{
		Success:   true,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func Paginated[T any](message string, items []T, total int64, page, pageSize int) *PaginatedResponse[T] {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResponse[T]{
		Success: true,
		Message: message,
		Data:    items,
		Pagination: &Pagination{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

type BodyWrapper[T any] struct {
	Body T
}

func Wrap[T any](data T) *BodyWrapper[T] {
	return &BodyWrapper[T]{Body: data}
}

func WrapSuccess[T any](message string, data T) *BodyWrapper[*APIResponse[T]] {
	return &BodyWrapper[*APIResponse[T]]{
		Body: Success(message, data),
	}
}

func WrapPaginated[T any](message string, items []T, total int64, page, pageSize int) *BodyWrapper[*PaginatedResponse[T]] {
	return &BodyWrapper[*PaginatedResponse[T]]{
		Body: Paginated(message, items, total, page, pageSize),
	}
}

func WrapMessage(message string) *BodyWrapper[*APIResponse[any]] {
	return &BodyWrapper[*APIResponse[any]]{
		Body: SuccessMessage(message),
	}
}

type CreatedResponse[T any] struct {
	Body T `json:"body"`
}

func Created[T any](message string, data T) *CreatedResponse[*APIResponse[T]] {
	return &CreatedResponse[*APIResponse[T]]{
		Body: Success(message, data),
	}
}
