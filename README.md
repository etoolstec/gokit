# gokit - Etools Common Kit

Kit compartilhado para todos os projetos Go da Etools. Substitui a cópia de `utils/`, `apperrors/` e `appvalidation/` em cada projeto.

## Instalação

```bash
go get github.com/etoolstec/go-kit@v1.0.0
```

Para desenvolvimento local com go.work:

```go
// /home/opc/prj/go.work
go 1.22
use (
    ./go-kit
    ./mcp-server-openerp
    ./erp-completo
)
```

## Pacotes

| Pacote | O que faz | Antes |
|--------|-----------|-------|
| `response` | Envelope padrão {data, total, page, ...} + helpers Gin/Fiber | `utils/response.go` |
| `apperror` | AppError com HTTP status | `apperrors/errors.go` |
| `converter` | Ponteiros, datas, números | `utils/converter.go` |
| `validator` | CPF/CNPJ, validação struct | `appvalidation/*.go` + `utils/validators.go` |
| `jwt` | Geração e validação JWT | `utils/jwt.go` |
| `mapper` | DTO <-> Model com copier (substitui reflection manual) | `utils/mapper.go` |
| `filter` | Filtros dinâmicos + GORM | `utils/filter.go` |
| `document` | Formatação CPF/CNPJ | `appvalidation/document.go` |

## Uso

### Response - sempre {data: ...}

```go
import "github.com/etoolstec/go-kit/response"

// Fiber (mcp-server-openerp)
func (h *Handler) List(c *fiber.Ctx) error {
    tenantID, ok := response.GetTenantIDOrRespond(c)
    if !ok { return nil }
    pag := response.GetPaginatedRequest(c)
    items, total, _ := h.service.List(pag.Page, pag.Limit)
    return response.Paginated(c, items, total, pag.Page, pag.Limit)
    // -> {"data": [...], "total": 100, "page": 1, "limit": 20, "total_pages": 5}
}

// Gin (erp-completo)
func (h *Handler) Get(c *gin.Context) error {
    id, ok := response.ParseIDParamGin(c, "id")
    if !ok { return nil }
    item, _ := h.service.Get(id)
    response.OKGin(c, item)
}
```

### Mapper - novo, estável

```go
import "github.com/etoolstec/go-kit/mapper"

// Antes: MapToModel com reflection que dava panic
// Agora: usa jinzhu/copier que é battle-tested
var model ProdutoModel
err := mapper.MapToModel(&dto, &model)

var dto ProdutoDTO
err = mapper.MapToDTO(&model, &dto)
```

### Validator

```go
import "github.com/etoolstec/go-kit/validator"

v := validator.New()
err := v.ValidateStruct(dto) // usa go-playground/validator + cpf/cnpj

docV := validator.NewDocumentValidator()
docV.IsValidCPF("529.982.247-25")
docV.FormatarDocumento("52998224725") // 529.982.247-25
```

## Versionamento

- v1.0.0 - primeira versão estável com os 6 helpers + apperror + document
- Use tags semânticas: v1.0.1, v1.1.0, etc.
