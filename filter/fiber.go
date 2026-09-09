package filter
import "github.com/gofiber/fiber/v2"
func QueryParamsToFiltersFiber(c *fiber.Ctx) map[string]any { m:=make(map[string]any); for k,v := range c.Queries() { if k!="page"&&k!="limit" { m[k]=v } }; return m }
