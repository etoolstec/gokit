package response
import ("net/http"; "github.com/gin-gonic/gin")
func OKGin[T any](c *gin.Context, data T) { c.JSON(http.StatusOK, NewApiResponse(data)) }
