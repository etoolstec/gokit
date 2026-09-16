package response

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/etoolstec/gokit/apperror"
	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// helper: roda um handler Gin e devolve (status, body)
func runGin(t *testing.T, handler gin.HandlerFunc) (int, string) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test/42", nil)
	handler(c)
	return w.Code, w.Body.String()
}

func TestGinOKGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		OKGin(c, map[string]any{"id": 1})
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

func TestGinCreatedGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		CreatedGin(c, map[string]any{"id": 99})
	})
	if status != 201 {
		t.Fatalf("status: want 201, got %d", status)
	}
	if !strings.Contains(body, `"data"`) {
		t.Errorf("envelope ausente: %s", body)
	}
}

func TestGinPaginatedGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		PaginatedGin(c, []int{1, 2, 3}, 100, 1, 3)
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	var resp PaginatedResponse[int]
	json.Unmarshal([]byte(body), &resp)
	if resp.Total != 100 || resp.TotalPages != 34 || len(resp.Data) != 3 {
		t.Errorf("paginated: %+v", resp)
	}
}

func TestGinDeletedGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		DeletedGin(c, 42)
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	if !strings.Contains(body, `"status":"deleted"`) || !strings.Contains(body, `"id":42`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestGinDeactivatedGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		DeactivatedGin(c)
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	if !strings.Contains(body, `"status":"deactivated"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestGinNotFoundGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		NotFoundGin(c, "cliente nao encontrado")
	})
	if status != 404 {
		t.Fatalf("status: want 404, got %d", status)
	}
	if !strings.Contains(body, `"error":"not_found"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestGinValidationErrorGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		ValidationErrorGin(c, "campo obrigatorio")
	})
	if status != 400 {
		t.Fatalf("status: want 400, got %d", status)
	}
	if !strings.Contains(body, `"error":"validation_error"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestGinInternalErrorGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		InternalErrorGin(c, "oops")
	})
	if status != 500 {
		t.Fatalf("status: want 500, got %d", status)
	}
	if !strings.Contains(body, `"internal_error"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestGinAppErrorGinNotFound(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		AppErrorGin(c, apperror.NewNotFoundError("nao encontrado"))
	})
	if status != 404 {
		t.Fatalf("status: want 404, got %d", status)
	}
	if !strings.Contains(body, `"not_found"`) {
		t.Errorf("ausente: %s", body)
	}
}

func TestGinFromErrorGinNil(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		FromErrorGin(c, nil)
	})
	// nada foi escrito, status default 200 do recorder
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	if body != "" {
		t.Errorf("esperado body vazio, got %s", body)
	}
}

func TestGinFromErrorGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		FromErrorGin(c, apperror.NewBadRequestError("invalido"))
	})
	if status != 400 {
		t.Fatalf("status: want 400, got %d", status)
	}
	if !strings.Contains(body, `"bad_request"`) && !strings.Contains(body, `"invalido"`) {
		t.Errorf("esperado erro mapeado: %s", body)
	}
}

func TestGinGetTenantIDGin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x", nil)
	c.Request.Header.Set("X-Tenant-ID", "7")

	id, err := GetTenantIDGin(c)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if id != 7 {
		t.Errorf("id: want 7, got %d", id)
	}
}

func TestGinGetTenantIDGinAusente(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	_, err := GetTenantIDGin(c)
	if err == nil {
		t.Fatal("esperava erro sem header")
	}
}

func TestGinGetTenantIDOrRespondGin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	_, ok := GetTenantIDOrRespondGin(c)
	if ok {
		t.Fatal("esperava false sem header")
	}
	if w.Code != 400 {
		t.Errorf("status: want 400, got %d", w.Code)
	}
}

func TestGinParseIDParamGin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "42"}}

	id, ok := ParseIDParamGin(c, "id")
	if !ok {
		t.Fatal("esperava true")
	}
	if id != 42 {
		t.Errorf("id: want 42, got %d", id)
	}
}

func TestGinParseIDParamGinInvalido(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}

	_, ok := ParseIDParamGin(c, "id")
	if ok {
		t.Fatal("esperava false com valor invalido")
	}
	if w.Code != 400 {
		t.Errorf("status: want 400, got %d", w.Code)
	}
}

func TestGinGetPaginatedRequestGin(t *testing.T) {
	cases := []struct {
		url                 string
		wantPage, wantLimit int
	}{
		{"/x", 1, 20},
		{"/x?page=3&limit=50", 3, 50},
		{"/x?page=-1&limit=999", 1, 20},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", c.url, nil)

		p := GetPaginatedRequestGin(ctx)
		if p.Page != c.wantPage || p.Limit != c.wantLimit {
			t.Errorf("%s: want (%d,%d), got (%d,%d)", c.url, c.wantPage, c.wantLimit, p.Page, p.Limit)
		}
	}
}

func TestGinGetQueryStringGin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x?q=joao", nil)

	if v := GetQueryStringGin(c, "q", "default"); v != "joao" {
		t.Errorf("want joao, got %s", v)
	}
	if v := GetQueryStringGin(c, "naoexiste", "default"); v != "default" {
		t.Errorf("want default, got %s", v)
	}
}

func TestGinNoContentGin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/x", nil)

	NoContentGin(c)
	c.Writer.WriteHeaderNow() // força flush do status

	if w.Code != 204 {
		t.Fatalf("status: want 204, got %d", w.Code)
	}
}

func TestGinOKWithMessageGin(t *testing.T) {
	status, body := runGin(t, func(c *gin.Context) {
		OKWithMessageGin(c, 42, "tudo certo")
	})
	if status != 200 {
		t.Fatalf("status: want 200, got %d", status)
	}
	if !strings.Contains(body, `"message":"tudo certo"`) {
		t.Errorf("ausente: %s", body)
	}
}
