package filter

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FilterConfig struct {
	Column   string
	Operator string
	Value    any
}

func QueryParamsToFilters(c *gin.Context) map[string]any {
	filters := make(map[string]any)
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return filters
	}
	excluded := map[string]bool{"limit": true, "offset": true, "page": true, "sort": true, "order": true}
	for key, values := range c.Request.URL.Query() {
		if len(values) == 0 || excluded[key] {
			continue
		}
		filters[key] = values[0]
	}
	return filters
}

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
		key = normalizeFilterKey(key)
		column, operator, ok := parseFilterKey(key, fieldToColumn)
		if !ok {
			continue
		}
		query = applyFilter(query, column, operator, value)
	}
	return query
}

func normalizeFilterKey(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "data_inicio", "inicio", "from", "created_from":
		return "created_at__gte"
	case "data_fim", "fim", "to", "created_to":
		return "created_at__lte"
	default:
		return key
	}
}

func buildFieldToColumnMap(t reflect.Type) map[string]string {
	fieldMap := make(map[string]string)
	collectFields(t, fieldMap)
	return fieldMap
}

func collectFields(t reflect.Type, fieldMap map[string]string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if field.Anonymous && ft.Kind() == reflect.Struct {
			collectFields(ft, fieldMap)
			continue
		}
		column := extractColumnFromTag(field.Tag.Get("gorm"))
		if column == "" {
			column = toSnake(field.Name)
		}
		if column == "" || column == "-" {
			continue
		}
		fieldMap[strings.ToLower(field.Name)] = column
		fieldMap[strings.ToLower(column)] = column
		if j := field.Tag.Get("json"); j != "" {
			name := strings.Split(j, ",")[0]
			if name != "" && name != "-" {
				fieldMap[strings.ToLower(name)] = column
			}
		}
	}
}

func extractColumnFromTag(tag string) string {
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

func toSnake(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(prev) || (nextLower && unicode.IsUpper(prev)) {
					b.WriteByte('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func parseFilterKey(key string, fieldMap map[string]string) (column string, operator string, ok bool) {
	parts := strings.Split(key, "__")
	fieldName := parts[0]
	operator = "="
	explicit := false
	if len(parts) > 1 {
		operator = mapOperator(parts[1])
		explicit = true
	}
	col, exists := fieldMap[strings.ToLower(fieldName)]
	if !exists {
		return "", "", false
	}
	if !explicit && operator == "=" && isTextSearchColumn(col) {
		operator = "LIKE"
	}
	return col, operator, true
}

func isTextSearchColumn(column string) bool {
	c := strings.ToLower(column)
	switch c {
	case "nome", "name", "descricao", "description", "email",
		"telefone", "phone", "documento", "cnpj", "cpf",
		"razao_social", "fantasia", "endereco", "cidade", "bairro",
		"logradouro", "complemento":
		return true
	default:
		return strings.Contains(c, "nome") || strings.HasSuffix(c, "_name")
	}
}

func mapOperator(op string) string {
	operators := map[string]string{
		"eq": "=", "neq": "!=", "gt": ">", "gte": ">=", "lt": "<", "lte": "<=",
		"like": "LIKE", "ilike": "LIKE",
		"in": "IN", "nin": "NOT IN", "between": "BETWEEN",
		"isnull": "IS NULL", "isnotnull": "IS NOT NULL",
	}
	if sqlOp, exists := operators[op]; exists {
		return sqlOp
	}
	return "="
}

func applyFilter(query *gorm.DB, column, operator string, value interface{}) *gorm.DB {
	switch operator {
	case "=", "!=", ">", ">=", "<", "<=":
		return query.Where(fmt.Sprintf("%s %s ?", column, operator), value)
	case "LIKE", "ILIKE":
		s, ok := value.(string)
		if !ok {
			return query
		}
		return query.Where(fmt.Sprintf("%s LIKE ?", column), "%"+s+"%")
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
