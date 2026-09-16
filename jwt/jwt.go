package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ============================================================
// CONSTANTES
// ============================================================

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token expirado")
	ErrInvalidUser  = errors.New("usuário inválido")
)

// ============================================================
// TYPES
// ============================================================

// TokenType define o tipo de token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTClaims representa as claims do token JWT
type JWTClaims struct {
	UserID    int                    `json:"user_id"`
	Login     string                 `json:"login,omitempty"`
	GrupoID   int                    `json:"grupo_id,omitempty"`
	EmpresaID int                    `json:"empresa_id,omitempty"`
	Type      string                 `json:"type"` // "access" ou "refresh"
	Extra     map[string]interface{} `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// ============================================================
// FUNÇÕES
// ============================================================

// GenerateToken gera um novo token JWT
func GenerateToken(
	userID int,
	login string,
	grupoID int,
	empresaID int,
	tokenType TokenType,
	secret string,
	expiresIn time.Duration,
) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		Login:     login,
		GrupoID:   grupoID,
		EmpresaID: empresaID,
		Type:      string(tokenType),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "openerp",
			Subject:   "user_auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateTokenPair gera um par de tokens (access e refresh)
func GenerateTokenPair(
	userID int,
	login string,
	extra map[string]interface{},
	secret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) (accessToken string, refreshToken string, expiresAt time.Time, err error) {
	now := time.Now()
	accessExp := now.Add(accessTTL)

	accessClaims := JWTClaims{
		UserID: userID,
		Login:  login,
		Type:   string(AccessToken),
		Extra:  extra,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gokit",
			Subject:   "user_auth",
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshClaims := JWTClaims{
		UserID: userID,
		Login:  login,
		Type:   string(RefreshToken),
		Extra:  extra,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gokit",
			Subject:   "user_auth",
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, accessExp, nil
}

// ValidateToken valida e extrai as claims do token
func ValidateToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verificar o método de assinatura
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de assinatura inválido")
			}
			return []byte(secret), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken valida apenas tokens de acesso
func ValidateAccessToken(tokenString string, secret string) (*JWTClaims, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return nil, err
	}

	if claims.Type != string(AccessToken) {
		return nil, errors.New("token não é um token de acesso")
	}

	return claims, nil
}

// ValidateRefreshToken valida apenas tokens de refresh
func ValidateRefreshToken(tokenString string, secret string) (*JWTClaims, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return nil, err
	}

	if claims.Type != string(RefreshToken) {
		return nil, errors.New("token não é um token de refresh")
	}

	return claims, nil
}

// GetUserIDFromToken extrai o ID do usuário do token
func GetUserIDFromToken(tokenString string, secret string) (int, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetClaim extrai um valor do Extra do token por chave
func GetClaim(tokenString string, secret string, key string) (interface{}, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	if claims.Extra == nil {
		return nil, nil
	}
	return claims.Extra[key], nil
}

// GetStringClaim extrai um valor string do Extra do token por chave
func GetStringClaim(tokenString string, secret string, key string) (string, error) {
	v, err := GetClaim(tokenString, secret, key)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	s, _ := v.(string)
	return s, nil
}

// GetUintClaim extrai um valor uint do Extra do token por chave
func GetUintClaim(tokenString string, secret string, key string) (uint, error) {
	v, err := GetClaim(tokenString, secret, key)
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	switch n := v.(type) {
	case float64:
		return uint(n), nil
	case int:
		return uint(n), nil
	case uint:
		return n, nil
	case int64:
		return uint(n), nil
	}
	return 0, nil
}
