package response

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/etoolstec/gokit/apperror"
	"github.com/gofiber/fiber/v2"
)

// helper: roda um handler Fiber e devolve (status, body)
func runFiber(t *testing.T, handler fiber.Handler) (int, string) {
	t.Helper()
	app := fiber.New()
	app.Get("/test/:id", handler)
	req := httptest.NewRequest("GET", "/test/42", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestFiberOK(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return OK(c, map[string]any{"id": 1})
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	var resp ApiResponse[map[string]any]
	json.Unmarshal([]byte(body), &resp)
	if resp.Data["id"].(float64) != 1 {
		t.Errorf("data.id: want 1, got %v", resp.Data["id"])
	}
}

func TestFiberCreated(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return Created(c, map[string]any{"id": 99})
	})
	if status != 201 {
		t.Fatalf("status: want 201, got %d", status)
	}
	if !strings.Contains(body, `"data"`) {
		t.Errorf("envelope ausente: %s", body)
	}
}

func TestFiberPaginated(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return Paginated(c, []int{1, 2, 3}, 100, 1, 3)
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	var resp PaginatedResponse[int]
	json.Unmarshal([]byte(body), &resp)
	if resp.Total != 100 {
		t.Errorf("total: want 100, got %d", resp.Total)
	}
	if resp.TotalPages != 34 {
		t.Errorf("total_pages: want 34, got %d", resp.TotalPages)
	}
	if len(resp.Data) != 3 {
		t.Errorf("len(data): want 3, got %d", len(resp.Data))
	}
}

func TestFiberDeleted(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return Deleted(c, 42)
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	if !strings.Contains(body, `"status":"deleted"`) {
		t.Errorf("status deleted ausente: %s", body)
	}
	if !strings.Contains(body, `"id":42`) {
		t.Errorf("id ausente: %s", body)
	}
}

func TestFiberDeactivated(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return Deactivated(c)
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	if !strings.Contains(body, `"status":"deactivated"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestFiberNotFound(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return NotFound(c, "cliente nao encontrado")
	})
	if status != 404 {
		t.Fatalf("status: want 404, got %d", status)
	}
	if !strings.Contains(body, `"error":"not_found"`) {
		t.Errorf("error code ausente: %s", body)
	}
}

func TestFiberValidationError(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return ValidationError(c, "campo obrigatorio")
	})
	if status != 400 {
		t.Fatalf("status: want 400, got %d", status)
	}
	if !strings.Contains(body, `"error":"validation_error"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestFiberAppErrorNotFound(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return AppError(c, apperror.NewNotFoundError("nao encontrado"))
	})
	if status != 404 {
		t.Fatalf("status: want 404, got %d", status)
	}
	if !strings.Contains(body, `"not_found"`) {
		t.Errorf("code ausente: %s", body)
	}
}

func TestFiberAppErrorBadRequest(t *testing.T) {
	status, _ := runFiber(t, func(c *fiber.Ctx) error {
		return AppError(c, apperror.NewBadRequestError("invalido"))
	})
	if status != 400 {
		t.Fatalf("status: want 400, got %d", status)
	}
}

func TestFiberAppErrorFallback(t *testing.T) {
	status, body := runFiber(t, func(c *fiber.Ctx) error {
		return AppError(c, io.EOF) // erro generico, nao AppError
	})
	if status != 500 {
		t.Fatalf("status: want 500, got %d", status)
	}
	if !strings.Contains(body, `"internal_error"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestFiberParseIDParam(t *testing.T) {
	_, body := runFiber(t, func(c *fiber.Ctx) error {
		id, err := ParseIDParam(c, "id")
		if err != nil {
			return err
		}
		return OK(c, id)
	})
	if !strings.Contains(body, `"data":42`) {
		t.Errorf("id parse errado: %s", body)
	}
}

func TestFiberParseIDParamOrRespondInvalido(t *testing.T) {
	app := fiber.New()
	app.Get("/test/:id", func(c *fiber.Ctx) error {
		_, ok := ParseIDParamOrRespond(c, "id")
		if !ok {
			return nil // já respondeu
		}
		return OK(c, "ok")
	})
	req := httptest.NewRequest("GET", "/test/abc", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatalf("status: want 400, got %d", resp.StatusCode)
	}
}

func TestFiberGetPaginatedRequest(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		p := GetPaginatedRequest(c)
		return OK(c, map[string]int{"page": p.Page, "limit": p.Limit})
	})

	cases := []struct {
		url       string
		wantPage  int
		wantLimit int
	}{
		{"/x", 1, 20},
		{"/x?page=3&limit=50", 3, 50},
		{"/x?page=-1&limit=999", 1, 20},
	}
	for _, c := range cases {
		req := httptest.NewRequest("GET", c.url, nil)
		resp, _ := app.Test(req, -1)
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var r ApiResponse[map[string]int]
		json.Unmarshal(b, &r)
		if r.Data["page"] != c.wantPage || r.Data["limit"] != c.wantLimit {
			t.Errorf("%s: want (%d,%d), got (%d,%d)",
				c.url, c.wantPage, c.wantLimit, r.Data["page"], r.Data["limit"])
		}
	}
}

func TestFiberGetTenantID(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		id, err := GetTenantID(c)
		if err != nil {
			return BadRequest(c, err.Error())
		}
		return OK(c, id)
	})

	// com header
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-Tenant-ID", "7")
	resp, _ := app.Test(req, -1)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), `"data":7`) {
		t.Errorf("id esperado 7: %s", string(b))
	}

	// sem header
	req2 := httptest.NewRequest("GET", "/x", nil)
	resp2, _ := app.Test(req2, -1)
	resp2.Body.Close()
	if resp2.StatusCode != 400 {
		t.Errorf("status: want 400, got %d", resp2.StatusCode)
	}
}

func TestFiberNoContent(t *testing.T) {
	status, _ := runFiber(t, func(c *fiber.Ctx) error {
		return NoContent(c)
	})
	if status != 204 {
		t.Fatalf("status: want 204, got %d", status)
	}
}

func TestFiberGetQueryString(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		return OK(c, GetQueryString(c, "q", "default"))
	})
	req := httptest.NewRequest("GET", "/x?q=joao", nil)
	resp, _ := app.Test(req, -1)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), `"data":"joao"`) {
		t.Errorf("q errado: %s", string(b))
	}

	req2 := httptest.NewRequest("GET", "/x", nil)
	resp2, _ := app.Test(req2, -1)
	b2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if !strings.Contains(string(b2), `"data":"default"`) {
		t.Errorf("default errado: %s", string(b2))
	}
}
