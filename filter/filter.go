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

// FilterConfig define como um filtro deve ser aplicado
type FilterConfig struct {
	Column   string // Nome da coluna no banco
	Operator string // Operador SQL: =, LIKE, >, <, etc.
	Value    any    // Valor a ser filtrado
}

// ============================================================
// FUNÇÕES
// ============================================================

func QueryParamsToFilters(c *gin.Context) map[string]any {
	filters := make(map[string]any)

	// Parâmetros que não são campos do model
	excludedParams := map[string]bool{
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
		// Opcional: pular parâmetros de paginação/ordenação
		if excludedParams[key] {
			continue
		}
		filters[key] = values[0]
	}
	return filters
}

// ApplyFilters aplica filtros dinamicamente a uma query usando reflection
func ApplyFilters(query *gorm.DB, model interface{}, filters map[string]interface{}) *gorm.DB {
	if len(filters) == 0 {
		return query
	}

	// Obter o tipo da struct
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Mapear nome do campo → nome da coluna
	fieldToColumn := buildFieldToColumnMap(t)

	// Aplicar cada filtro
	for key, value := range filters {
		if value == nil || value == "" {
			continue
		}

		// Determinar a coluna e o operador
		column, operator, ok := parseFilterKey(key, fieldToColumn)
		if !ok {
			continue
		}

		// Aplicar o filtro na query
		applyFilter(query, column, operator, value)
	}

	return query
}

// ============================================================
// FUNÇÕES AUXILIARES
// ============================================================

// buildFieldToColumnMap constrói um mapa: nome do campo → nome da coluna
func buildFieldToColumnMap(t reflect.Type) map[string]string {
	fieldMap := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Se for struct aninhada, não processar (evitar recursão infinita)
		if field.Type.Kind() == reflect.Struct && field.Anonymous == false {
			continue
		}

		// Verificar se o campo tem a tag gorm
		tag := field.Tag.Get("gorm")
		if tag == "" {
			continue
		}

		// Extrair o nome da coluna da tag
		column := extractColumnFromTag(tag)
		if column == "" {
			continue
		}

		// Mapear lowercase para facilitar a busca
		fieldMap[strings.ToLower(field.Name)] = column
		fieldMap[strings.ToLower(column)] = column
		fieldMap[strings.ToLower(field.Name)] = column
	}

	return fieldMap
}

// extractColumnFromTag extrai o nome da coluna da tag gorm
func extractColumnFromTag(tag string) string {
	// Tag exemplo: "column:ent_id;primaryKey;autoIncrement"
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// parseFilterKey analisa a chave do filtro e retorna a coluna e o operador
func parseFilterKey(key string, fieldMap map[string]string) (column string, operator string, ok bool) {
	// Verificar se a chave contém um operador: "nome__like", "data__gte", etc.
	parts := strings.Split(key, "__")
	fieldName := parts[0]
	operator = "=" // Operador padrão

	if len(parts) > 1 {
		operator = parts[1]
	}

	// Mapear operadores para SQL
	operator = mapOperator(operator)

	// Buscar a coluna no mapa
	col, exists := fieldMap[strings.ToLower(fieldName)]
	if !exists {
		return "", "", false
	}

	return col, operator, true
}

// mapOperator mapeia operadores amigáveis para SQL
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

// applyFilter aplica o filtro na query
func applyFilter(query *gorm.DB, column, operator string, value interface{}) {
	switch operator {
	case "=", "!=", ">", ">=", "<", "<=":
		query = query.Where(fmt.Sprintf("%s %s ?", column, operator), value)

	case "LIKE", "ILIKE":
		query = query.Where(fmt.Sprintf("%s %s ?", column, operator), "%"+value.(string)+"%")

	case "IN":
		query = query.Where(fmt.Sprintf("%s IN (?)", column), value)

	case "NOT IN":
		query = query.Where(fmt.Sprintf("%s NOT IN (?)", column), value)

	case "BETWEEN":
		// value deve ser um slice com 2 elementos
		if v, ok := value.([]interface{}); ok && len(v) == 2 {
			query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", column), v[0], v[1])
		}

	case "IS NULL":
		query = query.Where(fmt.Sprintf("%s IS NULL", column))

	case "IS NOT NULL":
		query = query.Where(fmt.Sprintf("%s IS NOT NULL", column))

	default:
		query = query.Where(fmt.Sprintf("%s = ?", column), value)
	}
}

// ============================================================
// FUNÇÕES DE ALTO NÍVEL (para usar no Repository)
// ============================================================

// FilterBuilder é um builder para criar filtros de forma fluente
type FilterBuilder struct {
	filters map[string]interface{}
}

// NewFilterBuilder cria um novo builder de filtros
func NewFilterBuilder(c *gin.Context) *FilterBuilder {
	return &FilterBuilder{
		filters: make(map[string]interface{}),
	}
}

// Eq adiciona um filtro de igualdade
func (fb *FilterBuilder) Eq(field string, value interface{}) *FilterBuilder {
	fb.filters[field+"__eq"] = value
	return fb
}

// Like adiciona um filtro LIKE
func (fb *FilterBuilder) Like(field string, value string) *FilterBuilder {
	fb.filters[field+"__like"] = value
	return fb
}

// Gt adiciona um filtro > (maior que)
func (fb *FilterBuilder) Gt(field string, value interface{}) *FilterBuilder {
	fb.filters[field+"__gt"] = value
	return fb
}

// Gte adiciona um filtro >= (maior ou igual)
func (fb *FilterBuilder) Gte(field string, value interface{}) *FilterBuilder {
	fb.filters[field+"__gte"] = value
	return fb
}

// Lt adiciona um filtro < (menor que)
func (fb *FilterBuilder) Lt(field string, value interface{}) *FilterBuilder {
	fb.filters[field+"__lt"] = value
	return fb
}

// Lte adiciona um filtro <= (menor ou igual)
func (fb *FilterBuilder) Lte(field string, value interface{}) *FilterBuilder {
	fb.filters[field+"__lte"] = value
	return fb
}

// Between adiciona um filtro BETWEEN
func (fb *FilterBuilder) Between(field string, from, to interface{}) *FilterBuilder {
	fb.filters[field+"__between"] = []interface{}{from, to}
	return fb
}

// In adiciona um filtro IN
func (fb *FilterBuilder) In(field string, values []interface{}) *FilterBuilder {
	fb.filters[field+"__in"] = values
	return fb
}

// Build retorna os filtros
func (fb *FilterBuilder) Build() map[string]interface{} {
	return fb.filters
}
