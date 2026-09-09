package filter
import "github.com/gin-gonic/gin"
func QueryParamsToFiltersGin(c *gin.Context) map[string]any { m:=make(map[string]any); for k,v := range c.Request.URL.Query() { if k!="page"&&k!="limit"&&len(v)>0 { m[k]=v[0] } }; return m }
