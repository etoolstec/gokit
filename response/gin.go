package response

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func OKGin[T any](c *gin.Context, data T)      { c.JSON(http.StatusOK, NewApiResponse(data)) }
func CreatedGin[T any](c *gin.Context, data T) { c.JSON(http.StatusCreated, NewApiResponse(data)) }
func PaginatedGin[T any](c *gin.Context, data []T, total int64, page, limit int) {
	c.JSON(http.StatusOK, NewPaginated(data, total, page, limit))
}
func ErrorGin(c *gin.Context, status int, code, msg string) {
	c.JSON(status, ErrorResponse{Error: code, Message: msg})
}
func ValidationErrorGin(c *gin.Context, msg string) {
	ErrorGin(c, http.StatusBadRequest, "validation_error", msg)
}
func NotFoundGin(c *gin.Context, msg string) { ErrorGin(c, http.StatusNotFound, "not_found", msg) }

func ParseIDParamGin(c *gin.Context, param string) (int, bool) {
	v := c.Param(param)
	id, err := strconv.Atoi(v)
	if err != nil {
		ValidationErrorGin(c, "ID deve ser número válido")
		return 0, false
	}
	return id, true
}
func GetQueryIntGin(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
func GetPaginatedRequestGin(c *gin.Context) PaginatedRequest {
	r := PaginatedRequest{Page: GetQueryIntGin(c, "page", 1), Limit: GetQueryIntGin(c, "limit", 20)}
	r.Normalize()
	return r
}
