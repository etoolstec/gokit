package filter
import "gorm.io/gorm"
func ApplyFilters(q *gorm.DB, model any, filters map[string]any) *gorm.DB { return q }
