package response

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/etoolstec/gokit/apperror"
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

func OKWithMessageGin[T any](c *gin.Context, data T, msg string) {
	c.JSON(http.StatusOK, NewApiResponseWithMessage(data, msg))
}

func NoContentGin(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func SuccessGin(c *gin.Context, status string) {
	c.JSON(http.StatusOK, SuccessResponse{Status: status})
}

func DeletedGin(c *gin.Context, id any) {
	c.JSON(http.StatusOK, SuccessResponse{Status: "deleted", ID: id})
}

func DeactivatedGin(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse{Status: "deactivated"})
}

func InternalErrorGin(c *gin.Context, msg string) {
	ErrorGin(c, http.StatusInternalServerError, "internal_error", msg)
}

func UnauthorizedGin(c *gin.Context, msg string) {
	ErrorGin(c, http.StatusUnauthorized, "unauthorized", msg)
}

func ForbiddenGin(c *gin.Context, msg string) {
	ErrorGin(c, http.StatusForbidden, "forbidden", msg)
}

func ConflictGin(c *gin.Context, msg string) {
	ErrorGin(c, http.StatusConflict, "conflict", msg)
}

func BadRequestGin(c *gin.Context, msg string) {
	ErrorGin(c, http.StatusBadRequest, "bad_request", msg)
}

func GetQueryStringGin(c *gin.Context, key, def string) string {
	v := c.Query(key)
	if v == "" {
		return def
	}
	return v
}

func GetTenantIDGin(c *gin.Context) (uint, error) {
	v := c.GetHeader("X-Tenant-ID")
	if v == "" {
		return 0, gin.Error{Err: nil}
	}
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func GetTenantIDOrRespondGin(c *gin.Context) (uint, bool) {
	id, err := GetTenantIDGin(c)
	if err != nil {
		BadRequestGin(c, "X-Tenant-ID obrigatorio")
		return 0, false
	}
	return id, true
}

func AppErrorGin(c *gin.Context, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		RespondWithAppErrorGin(c, appErr)
		return
	}
	InternalErrorGin(c, err.Error())
}

func RespondWithAppErrorGin(c *gin.Context, err *apperror.AppError) {
	switch err.Code {
	case http.StatusBadRequest:
		ValidationErrorGin(c, err.Message)
	case http.StatusConflict:
		ConflictGin(c, err.Message)
	case http.StatusNotFound:
		NotFoundGin(c, err.Message)
	case http.StatusInternalServerError:
		InternalErrorGin(c, err.Message)
	case http.StatusUnauthorized:
		UnauthorizedGin(c, err.Message)
	case http.StatusForbidden:
		ForbiddenGin(c, err.Message)
	default:
		ErrorGin(c, err.Code, "error", err.Message)
	}
}

func FromErrorGin(c *gin.Context, err error) {
	if err == nil {
		return
	}
	status := apperror.GetHTTPStatus(err)
	msg := err.Error()
	var appErr *apperror.AppError
	if errors.As(err, &appErr) && appErr.Message != "" {
		msg = appErr.Message
	}
	switch status {
	case 400:
		BadRequestGin(c, msg)
	case 401:
		UnauthorizedGin(c, msg)
	case 403:
		ForbiddenGin(c, msg)
	case 404:
		NotFoundGin(c, msg)
	case 409:
		ConflictGin(c, msg)
	default:
		if status >= 400 && status < 600 {
			ErrorGin(c, status, "error", msg)
		} else {
			InternalErrorGin(c, msg)
		}
	}
}

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
