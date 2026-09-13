package filter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================================
// TYPES
// ============================================================

// FilterConfig define como um filtro deve ser aplicado.
// Mantido para compatibilidade; o motor atual usa map[string]any + reflection.
type FilterConfig struct {
	Column   string // Nome da coluna no banco
	Operator string // Operador SQL: =, LIKE, >, <, etc.
	Value    any    // Valor a ser filtrado
}

// ============================================================
// ENTRADA (Gin)
// ============================================================

// QueryParamsToFilters extrai os query params do request como filtros.
// Parâmetros de paginação/ordenação são excluídos.
func QueryParamsToFilters(c *gin.Context) map[string]any {
	filters := make(map[string]any)
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return filters
	}

	excluded := map[string]bool{
		"limit":  true,
		"offset": true,
		"page":   true,
		"sort":   true,
		"order":  true,
	}

	for key, values := range c.Request.URL.Query() {
		if len(values) == 0 {
			continue
		}
		if excluded[key] {
			continue
		}
		filters[key] = values[0]
	}
	return filters
}

// ============================================================
// APLICAÇÃO
// ============================================================

// ApplyFilters aplica filtros dinamicamente a uma query usando reflection.
//
// Formato da chave:
//   - "campo"            → campo = valor
//   - "campo__like"      → campo LIKE %valor%
//   - "campo__gte"       → campo >= valor
//   - "campo__in"        → campo IN (v1, v2, ...)  — aceita "a,b,c" ou []any
//   - "campo__between"   → campo BETWEEN a AND b   — aceita "a,b" ou []any{a,b}
//
// Chaves sem correspondência no model são silenciosamente ignoradas.
func ApplyFilters(query *gorm.DB, model interface{}, filters map[string]interface{}) *gorm.DB {
	if query == nil || len(filters) == 0 {
		return query
	}

	t := reflect.TypeOf(model)
	if t == nil {
		return query
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return query
	}

	fieldToColumn := buildFieldToColumnMap(t)

	for key, value := range filters {
		if value == nil {
			continue
		}
		if s, ok := value.(string); ok && s == "" {
			continue
		}

		column, operator, ok := parseFilterKey(key, fieldToColumn)
		if !ok {
			continue
		}

		query = applyFilter(query, column, operator, value)
	}

	return query
}

// ============================================================
// AUXILIARES
// ============================================================

// buildFieldToColumnMap constrói o mapa nome_do_campo → coluna,
// usando a tag gorm:"column:...". Mapeia em lowercase para lookup case-insensitive.
func buildFieldToColumnMap(t reflect.Type) map[string]string {
	fieldMap := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Pula structs aninhadas nomeadas (não-anônimas).
		// Embedded anônimos (ex: gorm.Model) também são pulados aqui,
		// porque não têm tag gorm:"column:" própria.
		if field.Type.Kind() == reflect.Struct && !field.Anonymous {
			continue
		}

		tag := field.Tag.Get("gorm")
		if tag == "" {
			continue
		}

		column := extractColumnFromTag(tag)
		if column == "" {
			continue
		}

		fieldMap[strings.ToLower(field.Name)] = column
		fieldMap[strings.ToLower(column)] = column
	}

	return fieldMap
}

// extractColumnFromTag extrai o nome da coluna de uma tag gorm.
// Ex: "column:ent_id;primaryKey" → "ent_id"
func extractColumnFromTag(tag string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// parseFilterKey extrai coluna e operador de uma chave de filtro.
// Retorna ok=false quando o campo não existe no model.
func parseFilterKey(key string, fieldMap map[string]string) (column string, operator string, ok bool) {
	parts := strings.Split(key, "__")
	fieldName := parts[0]
	operator = "="

	if len(parts) > 1 {
		operator = mapOperator(parts[1])
	}

	col, exists := fieldMap[strings.ToLower(fieldName)]
	if !exists {
		return "", "", false
	}
	return col, operator, true
}

// mapOperator traduz operadores amigáveis para SQL.
func mapOperator(op string) string {
	operators := map[string]string{
		"eq":        "=",
		"neq":       "!=",
		"gt":        ">",
		"gte":       ">=",
		"lt":        "<",
		"lte":       "<=",
		"like":      "LIKE",
		"ilike":     "ILIKE",
		"in":        "IN",
		"nin":       "NOT IN",
		"between":   "BETWEEN",
		"isnull":    "IS NULL",
		"isnotnull": "IS NOT NULL",
	}

	if sqlOp, exists := operators[op]; exists {
		return sqlOp
	}
	return "="
}

// applyFilter aplica UM filtro e devolve a query modificada.
// GORM é imutável: query.Where(...) retorna um *gorm.DB novo.
// Por isso o retorno é obrigatório — o caller precisa reatribuir.
func applyFilter(query *gorm.DB, column, operator string, value interface{}) *gorm.DB {
	switch operator {
	case "=", "!=", ">", ">=", "<", "<=":
		return query.Where(fmt.Sprintf("%s %s ?", column, operator), value)

	case "LIKE", "ILIKE":
		s, ok := value.(string)
		if !ok {
			return query
		}
		return query.Where(fmt.Sprintf("%s %s ?", column, operator), "%"+s+"%")

	case "IN":
		values, ok := toSlice(value)
		if !ok || len(values) == 0 {
			return query
		}
		return query.Where(fmt.Sprintf("%s IN ?", column), values)

	case "NOT IN":
		values, ok := toSlice(value)
		if !ok || len(values) == 0 {
			return query
		}
		return query.Where(fmt.Sprintf("%s NOT IN ?", column), values)

	case "BETWEEN":
		values, ok := toSlice(value)
		if !ok || len(values) != 2 {
			return query
		}
		return query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", column), values[0], values[1])

	case "IS NULL":
		return query.Where(fmt.Sprintf("%s IS NULL", column))

	case "IS NOT NULL":
		return query.Where(fmt.Sprintf("%s IS NOT NULL", column))

	default:
		return query.Where(fmt.Sprintf("%s = ?", column), value)
	}
}

// toSlice normaliza value para []interface{}.
// Aceita []interface{} direto, ou string CSV ("a,b,c") que vira slice.
func toSlice(value interface{}) ([]interface{}, bool) {
	switch v := value.(type) {
	case []interface{}:
		return v, true
	case []string:
		out := make([]interface{}, len(v))
		for i, s := range v {
			out[i] = s
		}
		return out, true
	case string:
		if v == "" {
			return nil, false
		}
		parts := strings.Split(v, ",")
		out := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out, true
	default:
		return nil, false
	}
}
