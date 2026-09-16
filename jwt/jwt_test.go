package jwt

import (
	"testing"
	"time"
)

const testSecret = "test-secret-123"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := GenerateToken(42, "user@test.com", 1, 2, AccessToken, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("erro ao gerar: %v", err)
	}
	if token == "" {
		t.Fatal("token vazio")
	}

	claims, err := ValidateAccessToken(token, testSecret)
	if err != nil {
		t.Fatalf("erro ao validar: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID: want 42, got %d", claims.UserID)
	}
	if claims.Login != "user@test.com" {
		t.Errorf("Login: want user@test.com, got %s", claims.Login)
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	token, _ := GenerateToken(1, "a", 0, 0, AccessToken, testSecret, time.Hour)
	if _, err := ValidateToken(token, "outro"); err == nil {
		t.Fatal("esperava erro com secret errado")
	}
}

func TestExpiredToken(t *testing.T) {
	token, _ := GenerateToken(1, "a", 0, 0, AccessToken, testSecret, -time.Hour)
	_, err := ValidateAccessToken(token, testSecret)
	if err == nil {
		t.Fatal("esperava erro com token expirado")
	}
	if err != ErrExpiredToken {
		t.Errorf("want ErrExpiredToken, got %v", err)
	}
}

func TestValidateAccessRejectsRefresh(t *testing.T) {
	token, _ := GenerateToken(1, "a", 0, 0, RefreshToken, testSecret, time.Hour)
	if _, err := ValidateAccessToken(token, testSecret); err == nil {
		t.Fatal("access validator aceitou refresh")
	}
}

func TestValidateRefreshRejectsAccess(t *testing.T) {
	token, _ := GenerateToken(1, "a", 0, 0, AccessToken, testSecret, time.Hour)
	if _, err := ValidateRefreshToken(token, testSecret); err == nil {
		t.Fatal("refresh validator aceitou access")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	extra := map[string]interface{}{
		"tenant_id": uint(7),
		"role":      "admin",
		"email":     "a@b.com",
	}
	access, refresh, exp, err := GenerateTokenPair(42, "a@b.com", extra, testSecret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("tokens vazios")
	}
	if access == refresh {
		t.Fatal("access e refresh devem ser diferentes")
	}
	if exp.Before(time.Now()) {
		t.Fatal("exp no passado")
	}

	claims, err := ValidateAccessToken(access, testSecret)
	if err != nil {
		t.Fatalf("access invalido: %v", err)
	}
	if claims.Extra["role"] != "admin" {
		t.Errorf("role: want admin, got %v", claims.Extra["role"])
	}
	if _, err := ValidateRefreshToken(refresh, testSecret); err != nil {
		t.Fatalf("refresh invalido: %v", err)
	}
}

func TestGetStringClaim(t *testing.T) {
	extra := map[string]interface{}{"role": "admin", "email": "x@y.com"}
	access, _, _, _ := GenerateTokenPair(1, "x@y.com", extra, testSecret, time.Hour, time.Hour)

	role, err := GetStringClaim(access, testSecret, "role")
	if err != nil || role != "admin" {
		t.Errorf("role: want admin, got %q (err=%v)", role, err)
	}

	vazio, err := GetStringClaim(access, testSecret, "naoexiste")
	if err != nil || vazio != "" {
		t.Errorf("inexistente: want vazio, got %q (err=%v)", vazio, err)
	}
}

func TestGetUintClaim(t *testing.T) {
	extra := map[string]interface{}{"tenant_id": uint(7)}
	access, _, _, _ := GenerateTokenPair(1, "x", extra, testSecret, time.Hour, time.Hour)

	tenant, err := GetUintClaim(access, testSecret, "tenant_id")
	if err != nil || tenant != 7 {
		t.Errorf("tenant: want 7, got %d (err=%v)", tenant, err)
	}
}

func TestGetUintClaimFromFloat(t *testing.T) {
	// depois do round-trip JSON, num vem como float64
	extra := map[string]interface{}{"tenant_id": 7.0}
	access, _, _, _ := GenerateTokenPair(1, "x", extra, testSecret, time.Hour, time.Hour)

	tenant, err := GetUintClaim(access, testSecret, "tenant_id")
	if err != nil || tenant != 7 {
		t.Errorf("tenant float: want 7, got %d (err=%v)", tenant, err)
	}
}

func TestGetUserIDFromToken(t *testing.T) {
	token, _ := GenerateToken(99, "a", 0, 0, AccessToken, testSecret, time.Hour)
	id, err := GetUserIDFromToken(token, testSecret)
	if err != nil || id != 99 {
		t.Errorf("id: want 99, got %d (err=%v)", id, err)
	}
}

func TestGetClaimInexistente(t *testing.T) {
	access, _, _, _ := GenerateTokenPair(1, "x", nil, testSecret, time.Hour, time.Hour)
	v, err := GetClaim(access, testSecret, "qualquer")
	if err != nil {
		t.Errorf("err inesperado: %v", err)
	}
	if v != nil {
		t.Errorf("want nil, got %v", v)
	}
}
