package filter

import "github.com/gofiber/fiber/v2"

// QueryParamsToFiltersFiber extrai os query params do request Fiber como filtros.
// Parâmetros de paginação/ordenação são excluídos.
func QueryParamsToFiltersFiber(c *fiber.Ctx) map[string]any {
	m := make(map[string]any)
	if c == nil {
		return m
	}

	excluded := map[string]bool{
		"limit":  true,
		"offset": true,
		"page":   true,
		"sort":   true,
		"order":  true,
	}

	for k, v := range c.Queries() {
		if excluded[k] {
			continue
		}
		m[k] = v
	}
	return m

}
