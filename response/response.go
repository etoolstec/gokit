package response
type ApiResponse[T any] struct { Data T `json:"data"`; Message string `json:"message,omitempty"` }
type PaginatedResponse[T any] struct { Data []T `json:"data"`; Total int64 `json:"total"`; Page int `json:"page"`; Limit int `json:"limit"`; TotalPages int64 `json:"total_pages"` }
type PaginatedRequest struct { Page int; Limit int }
func (r *PaginatedRequest) Normalize() { if r.Page <= 0 { r.Page = 1 }; if r.Limit <= 0 || r.Limit > 100 { r.Limit = 20 } }
func (r PaginatedRequest) Offset() int { return (r.Page - 1) * r.Limit }
type ErrorResponse struct { Error string `json:"error"`; Message string `json:"message"` }
type SuccessResponse struct { Status string `json:"status"` }
func NewPaginated[T any](data []T, total int64, page, limit int) PaginatedResponse[T] { if data == nil { data = []T{} }; tp := int64(0); if limit>0 { tp = (total + int64(limit) -1)/ int64(limit) }; return PaginatedResponse[T]{Data: data, Total: total, Page: page, Limit: limit, TotalPages: tp} }
func NewApiResponse[T any](data T) ApiResponse[T] { return ApiResponse[T]{Data: data} }
