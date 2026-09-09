package apperror

import (
	"errors"
	"fmt"
)

// AppError representa um erro da aplicação com código HTTP
type AppError struct {
	Code    int    // HTTP status code
	Message string // Mensagem amigável
	Err     error  // Erro original (opcional)
}

// Error implementa a interface error
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap permite que errors.Is/As funcionem
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus retorna o código HTTP apropriado
func (e *AppError) HTTPStatus() int {
	return e.Code
}

// Funções helpers para criar erros comuns

// NewValidationError cria erro de validação (400)
func NewValidationError(msg string) *AppError {
	return &AppError{
		Code:    400,
		Message: msg,
	}
}

// NewConflictError cria erro de conflito (409)
func NewConflictError(msg string) *AppError {
	return &AppError{
		Code:    409,
		Message: msg,
	}
}

// NewNotFoundError cria erro de não encontrado (404)
func NewNotFoundError(msg string) *AppError {
	return &AppError{
		Code:    404,
		Message: msg,
	}
}

// NewInternalError cria erro interno (500)
func NewInternalError(msg string, err error) *AppError {
	return &AppError{
		Code:    500,
		Message: msg,
		Err:     err,
	}
}

// NewBadRequestError cria erro de requisição inválida (400)
func NewBadRequestError(msg string) *AppError {
	return &AppError{
		Code:    400,
		Message: msg,
	}
}

func NewUnauthorizedError(msg string) *AppError {
	return &AppError{
		Code:    401,
		Message: msg,
	}
}

// IsAppError verifica se um erro é do tipo AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetHTTPStatus extrai o status HTTP de um erro, ou retorna 500 se não for AppError
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return 500 // Internal Server Error padrão
}
