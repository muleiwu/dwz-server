package helper

import (
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func resetJWTHelper() {
	jwtHelper = nil
	jwtHelperOnce = sync.Once{}
}

func TestGenerateAndValidateToken(t *testing.T) {
	resetJWTHelper()
	InitJWTHelper("test-secret-key-for-jwt-testing", 24)

	token, expiresAt, err := GetJWTHelper().GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("expiresAt should be in the future")
	}

	claims, err := GetJWTHelper().ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != 123 {
		t.Errorf("expected UserID 123, got %d", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("expected Username testuser, got %s", claims.Username)
	}
	if claims.Issuer != "dwz-server" {
		t.Errorf("expected Issuer dwz-server, got %s", claims.Issuer)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	resetJWTHelper()
	InitJWTHelper("secret-a", 24)
	token, _, err := GetJWTHelper().GenerateToken(1, "user1")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Reset and init with different secret
	resetJWTHelper()
	InitJWTHelper("secret-b", 24)

	_, err = GetJWTHelper().ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	resetJWTHelper()
	InitJWTHelper("test-secret", 0) // 0 hours = expires immediately

	// Manually create an already-expired token
	claims := LoginClaims{
		UserID:   1,
		Username: "user1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "dwz-server",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	_, err := GetJWTHelper().ValidateToken(tokenString)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	resetJWTHelper()
	InitJWTHelper("test-secret", 24)

	_, err := GetJWTHelper().ValidateToken("not-a-jwt-token")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestValidateToken_TamperedPayload(t *testing.T) {
	resetJWTHelper()
	InitJWTHelper("test-secret", 24)

	token, _, _ := GetJWTHelper().GenerateToken(1, "user1")

	// Tamper with the token by modifying a character in the payload
	tampered := token[:len(token)-5] + "XXXXX"
	_, err := GetJWTHelper().ValidateToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}
