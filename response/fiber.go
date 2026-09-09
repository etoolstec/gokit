package response
import ("strconv"; "github.com/gofiber/fiber/v2")
func OK[T any](c *fiber.Ctx, data T) error { return c.Status(200).JSON(NewApiResponse(data)) }
func Paginated[T any](c *fiber.Ctx, data []T, total int64, page, limit int) error { return c.Status(200).JSON(NewPaginated(data, total, page, limit)) }
func Error(c *fiber.Ctx, status int, code, msg string) error { return c.Status(status).JSON(ErrorResponse{Error: code, Message: msg}) }
func ValidationError(c *fiber.Ctx, msg string) error { return Error(c, 400, "validation_error", msg) }
func NotFound(c *fiber.Ctx, msg string) error { return Error(c, 404, "not_found", msg) }
func BadRequest(c *fiber.Ctx, msg string) error { return Error(c, 400, "bad_request", msg) }
func ParseIDParamOrRespond(c *fiber.Ctx, param string) (uint, bool) { v := c.Params(param); id, err := strconv.ParseUint(v,10,64); if err!=nil { _ = ValidationError(c,"ID invalido"); return 0,false }; return uint(id), true }
func GetQueryInt(c *fiber.Ctx, key string, def int) int { v:=c.Query(key); if v==""{return def}; i,_:=strconv.Atoi(v); return i }
func GetPaginatedRequest(c *fiber.Ctx) PaginatedRequest { r:=PaginatedRequest{Page:GetQueryInt(c,"page",1), Limit:GetQueryInt(c,"limit",20)}; r.Normalize(); return r }
func GetTenantIDOrRespond(c *fiber.Ctx) (uint, bool) { v:=c.Get("X-Tenant-ID"); id,err:=strconv.ParseUint(v,10,64); if err!=nil { _=BadRequest(c,"X-Tenant-ID obrigatorio"); return 0,false }; return uint(id), true }
