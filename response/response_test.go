package response

import (
	"encoding/json"
	"testing"
)

type payload struct {
	Nome string `json:"nome"`
}

func TestNewApiResponse(t *testing.T) {
	r := NewApiResponse(payload{Nome: "X"})
	b, _ := json.Marshal(r)
	want := `{"data":{"nome":"X"}}`
	if string(b) != want {
		t.Errorf("want %s, got %s", want, string(b))
	}
}

func TestNewApiResponseWithMessage(t *testing.T) {
	r := NewApiResponseWithMessage(42, "ok")
	b, _ := json.Marshal(r)
	want := `{"data":42,"message":"ok"}`
	if string(b) != want {
		t.Errorf("want %s, got %s", want, string(b))
	}
}

func TestNewPaginatedTotalPages(t *testing.T) {
	cases := []struct {
		total      int64
		limit      int
		totalPages int64
	}{
		{0, 20, 0},
		{1, 20, 1},
		{20, 20, 1},
		{21, 20, 2},
		{100, 10, 10},
		{101, 10, 11},
		{0, 0, 0},
	}
	for _, c := range cases {
		r := NewPaginated([]int{}, c.total, 1, c.limit)
		if r.TotalPages != c.totalPages {
			t.Errorf("total=%d limit=%d: want %d, got %d", c.total, c.limit, c.totalPages, r.TotalPages)
		}
	}
}

func TestNewPaginatedNilData(t *testing.T) {
	r := NewPaginated[int](nil, 0, 1, 20)
	if r.Data == nil {
		t.Fatal("Data nil deveria virar slice vazio")
	}
	if len(r.Data) != 0 {
		t.Errorf("want 0, got %d", len(r.Data))
	}
}

func TestPaginatedRequestNormalize(t *testing.T) {
	cases := []struct {
		page, limit         int
		wantPage, wantLimit int
	}{
		{0, 0, 1, 20},
		{-1, -5, 1, 20},
		{2, 50, 2, 50},
		{2, 200, 2, 20},
		{1, 100, 1, 100},
	}
	for _, c := range cases {
		r := PaginatedRequest{Page: c.page, Limit: c.limit}
		r.Normalize()
		if r.Page != c.wantPage || r.Limit != c.wantLimit {
			t.Errorf("page=%d limit=%d: want (%d,%d), got (%d,%d)",
				c.page, c.limit, c.wantPage, c.wantLimit, r.Page, r.Limit)
		}
	}
}

func TestPaginatedRequestOffset(t *testing.T) {
	cases := []struct {
		page, limit, offset int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{5, 25, 100},
	}
	for _, c := range cases {
		r := PaginatedRequest{Page: c.page, Limit: c.limit}
		if r.Offset() != c.offset {
			t.Errorf("page=%d limit=%d: want %d, got %d", c.page, c.limit, c.offset, r.Offset())
		}
	}
}

func TestSuccessResponseJSON(t *testing.T) {
	r := SuccessResponse{Status: "deleted", ID: 42}
	b, _ := json.Marshal(r)
	want := `{"status":"deleted","id":42}`
	if string(b) != want {
		t.Errorf("want %s, got %s", want, string(b))
	}
}

func TestSuccessResponseOmitEmpty(t *testing.T) {
	r := SuccessResponse{Status: "deactivated"}
	b, _ := json.Marshal(r)
	want := `{"status":"deactivated"}`
	if string(b) != want {
		t.Errorf("want %s, got %s", want, string(b))
	}
}

func TestErrorResponseJSON(t *testing.T) {
	r := ErrorResponse{Error: "not_found", Message: "cliente nao encontrado"}
	b, _ := json.Marshal(r)
	want := `{"error":"not_found","message":"cliente nao encontrado"}`
	if string(b) != want {
		t.Errorf("want %s, got %s", want, string(b))
	}
}