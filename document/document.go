package document

import (
	"regexp"
	"strings"

	"github.com/paemuri/brdoc"
)

type DocumentValidator interface {
	IsValidDocumento(documento string) bool
	IsValidCPF(documento string) bool
	IsValidCNPJ(documento string) bool
	LimparDocumento(documento string) string
	FormatarDocumento(documento string) string
}

// documentValidator implementa a interface DocumentValidator
type documentValidator struct{}

// NewDocumentValidator cria um novo validador de documentos
func NewDocumentValidator() DocumentValidator {
	return &documentValidator{}
}

// IsValidDocumento verifica se o documento é um CPF ou CNPJ válido
func (d *documentValidator) IsValidDocumento(documento string) bool {
	return brdoc.IsCPF(documento) || brdoc.IsCNPJ(documento)
}

// IsValidCPF verifica se é um CPF válido
func (d *documentValidator) IsValidCPF(documento string) bool {
	return brdoc.IsCPF(documento)
}

// IsValidCNPJ verifica se é um CNPJ válido
func (d *documentValidator) IsValidCNPJ(documento string) bool {
	return brdoc.IsCNPJ(documento)
}

// LimparDocumento remove caracteres especiais do documento
func (d *documentValidator) LimparDocumento(documento string) string {
	documento = strings.TrimSpace(documento)
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(documento, "")
}

// FormatarDocumento formata CPF ou CNPJ automaticamente
func (d *documentValidator) FormatarDocumento(documento string) string {
	documento = d.LimparDocumento(documento)

	if len(documento) == 11 {
		return formatCPF(documento)
	}
	if len(documento) == 14 {
		return formatCNPJ(documento)
	}
	return documento
}

func formatCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}
	return cpf[:3] + "." + cpf[3:6] + "." + cpf[6:9] + "-" + cpf[9:]
}

func formatCNPJ(cnpj string) string {
	if len(cnpj) != 14 {
		return cnpj
	}
	return cnpj[:2] + "." + cnpj[2:5] + "." + cnpj[5:8] + "/" + cnpj[8:12] + "-" + cnpj[12:]
}
