package response

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func OK[T any](c *fiber.Ctx, data T) error {
	return c.Status(fiber.StatusOK).JSON(NewApiResponse(data))
}
func OKWithMessage[T any](c *fiber.Ctx, data T, msg string) error {
	return c.Status(fiber.StatusOK).JSON(NewApiResponseWithMessage(data, msg))
}
func Created[T any](c *fiber.Ctx, data T) error {
	return c.Status(fiber.StatusCreated).JSON(NewApiResponse(data))
}
func Paginated[T any](c *fiber.Ctx, data []T, total int64, page, limit int) error {
	return c.Status(fiber.StatusOK).JSON(NewPaginated(data, total, page, limit))
}
func NoContent(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }
func Success(c *fiber.Ctx, status string) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{Status: status})
}
func Deleted(c *fiber.Ctx, id any) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{Status: "deleted", ID: id})
}
func Deactivated(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{Status: "deactivated"})
}
func Error(c *fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(ErrorResponse{Error: code, Message: msg})
}
func ValidationError(c *fiber.Ctx, msg string) error {
	return Error(c, fiber.StatusBadRequest, "validation_error", msg)
}
func NotFound(c *fiber.Ctx, msg string) error {
	return Error(c, fiber.StatusNotFound, "not_found", msg)
}
func InternalError(c *fiber.Ctx, msg string) error {
	return Error(c, fiber.StatusInternalServerError, "internal_error", msg)
}
func Unauthorized(c *fiber.Ctx, msg string) error {
	return Error(c, fiber.StatusUnauthorized, "unauthorized", msg)
}
func Forbidden(c *fiber.Ctx, msg string) error {
	return Error(c, fiber.StatusForbidden, "forbidden", msg)
}
func Conflict(c *fiber.Ctx, msg string) error { return Error(c, fiber.StatusConflict, "conflict", msg) }
func BadRequest(c *fiber.Ctx, msg string) error {
	return Error(c, fiber.StatusBadRequest, "bad_request", msg)
}

func ParseIDParam(c *fiber.Ctx, param string) (uint, error) {
	v := c.Params(param)
	if v == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "ID obrigatório")
	}
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
func ParseIDParamOrRespond(c *fiber.Ctx, param string) (uint, bool) {
	id, err := ParseIDParam(c, param)
	if err != nil {
		_ = ValidationError(c, "ID deve ser número válido")
		return 0, false
	}
	return id, true
}
func GetQueryInt(c *fiber.Ctx, key string, def int) int {
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
func GetQueryString(c *fiber.Ctx, key, def string) string {
	v := c.Query(key)
	if v == "" {
		return def
	}
	return v
}
func GetPaginatedRequest(c *fiber.Ctx) PaginatedRequest {
	r := PaginatedRequest{Page: GetQueryInt(c, "page", 1), Limit: GetQueryInt(c, "limit", 20)}
	r.Normalize()
	return r
}
func GetTenantID(c *fiber.Ctx) (uint, error) {
	v := c.Get("X-Tenant-ID")
	if v == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "X-Tenant-ID obrigatório")
	}
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
func GetTenantIDOrRespond(c *fiber.Ctx) (uint, bool) {
	id, err := GetTenantID(c)
	if err != nil {
		_ = BadRequest(c, err.Error())
		return 0, false
	}
	return id, true
}
