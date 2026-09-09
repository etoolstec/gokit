package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/etoolstec/gokit/document"
	"github.com/go-playground/validator/v10"
	"github.com/paemuri/brdoc"
)

type Validator interface {
	ValidateStruct(s interface{}) error
	ValidateField(field interface{}, tag string) error
	GetMandatoryFields(structure interface{}) []string
	ValidateMandatoryFields(structure interface{}) error
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

func isFieldMandatoryByType(fieldValue reflect.Value) bool {
	if fieldValue.Kind() == reflect.Ptr {
		return false
	}
	if fieldValue.Kind() == reflect.Interface {
		return false
	}
	if fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Map || fieldValue.Kind() == reflect.Array {
		return false
	}
	switch fieldValue.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Struct:
		return true
	default:
		return false
	}
}

func CleanString(value string) string {
	if value == "" {
		return ""
	}
	value = strings.TrimSpace(value)
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(value, " ")
}

type documentValidator struct{}

func NewDocumentValidator() document.DocumentValidator { return &documentValidator{} }
func (d *documentValidator) IsValidDocumento(documento string) bool {
	return brdoc.IsCPF(documento) || brdoc.IsCNPJ(documento)
}
func (d *documentValidator) IsValidCPF(documento string) bool  { return brdoc.IsCPF(documento) }
func (d *documentValidator) IsValidCNPJ(documento string) bool { return brdoc.IsCNPJ(documento) }
func (d *documentValidator) LimparDocumento(documento string) string {
	documento = strings.TrimSpace(documento)
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(documento, "")
}
func (d *documentValidator) FormatarDocumento(documento string) string {
	documento = d.LimparDocumento(documento)
	if len(documento) == 11 {
		return documento[:3] + "." + documento[3:6] + "." + documento[6:9] + "-" + documento[9:]
	}
	if len(documento) == 14 {
		return documento[:2] + "." + documento[2:5] + "." + documento[5:8] + "/" + documento[8:12] + "-" + documento[12:]
	}
	return documento
}

type playValidator struct{ validate *validator.Validate }

func NewPlayValidator() Validator {
	v := validator.New()
	v.RegisterValidation("cpf", func(fl validator.FieldLevel) bool { return brdoc.IsCPF(fl.Field().String()) })
	v.RegisterValidation("cnpj", func(fl validator.FieldLevel) bool { return brdoc.IsCNPJ(fl.Field().String()) })
	v.RegisterValidation("documento", func(fl validator.FieldLevel) bool {
		doc := fl.Field().String()
		return brdoc.IsCPF(doc) || brdoc.IsCNPJ(doc)
	})
	return &playValidator{validate: v}
}

func (v *playValidator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		return fmt.Errorf("erro de validação: %w", err)
	}
	return nil
}
func (v *playValidator) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}
func (v *playValidator) GetMandatoryFields(structure interface{}) []string {
	var fields []string
	val := reflect.ValueOf(structure)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fields
	}
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		if v.isFieldMandatory(field, fieldValue) {
			fields = append(fields, field.Name)
		}
	}
	return fields
}
func (v *playValidator) ValidateMandatoryFields(structure interface{}) error {
	mandatoryFields := v.GetMandatoryFields(structure)
	val := reflect.ValueOf(structure)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for _, fieldName := range mandatoryFields {
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			return fmt.Errorf("campo obrigatório %s não encontrado", fieldName)
		}
		if isEmptyValue(field) {
			return fmt.Errorf("campo obrigatório %s está vazio", fieldName)
		}
	}
	return nil
}
func (v *playValidator) isFieldMandatory(field reflect.StructField, fieldValue reflect.Value) bool {
	tag := field.Tag.Get("validate")
	if strings.Contains(tag, "required") {
		return true
	}
	if tag := field.Tag.Get("mandatory"); tag == "true" {
		return true
	}
	if tag := field.Tag.Get("binding"); strings.Contains(tag, "required") {
		return true
	}
	return isFieldMandatoryByType(fieldValue)
}
