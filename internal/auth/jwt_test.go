package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTService_Generate(t *testing.T) {
	svc := NewJWTService("test-secret")

	token, err := svc.Generate("user-123", "testuser")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestJWTService_Validate_ValidToken(t *testing.T) {
	svc := NewJWTService("test-secret")

	token, _ := svc.Generate("user-123", "testuser")

	claims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected user_id 'user-123', got %s", claims.UserID)
	}

	if claims.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", claims.Username)
	}
}

func TestJWTService_Validate_InvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret")

	_, err := svc.Validate("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestJWTService_Validate_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-1")
	svc2 := NewJWTService("secret-2")

	token, _ := svc1.Generate("user-123", "testuser")

	_, err := svc2.Validate(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestJWTService_Validate_ExpiredToken(t *testing.T) {
	svc := &JWTService{
		secret:     []byte("test-secret"),
		expiration: -1 * time.Hour, // Already expired
	}

	token, _ := svc.Generate("user-123", "testuser")

	_, err := svc.Validate(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTService_Validate_TamperedToken(t *testing.T) {
	svc := NewJWTService("test-secret")

	token, _ := svc.Generate("user-123", "testuser")

	// Tamper with the token by changing a character
	tamperedToken := token[:len(token)-5] + "xxxxx"

	_, err := svc.Validate(tamperedToken)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestJWTService_Claims_ExpirationSet(t *testing.T) {
	svc := NewJWTService("test-secret")

	token, _ := svc.Generate("user-123", "testuser")
	claims, _ := svc.Validate(token)

	if claims.ExpiresAt == nil {
		t.Fatal("expected expiration to be set")
	}

	// Should expire roughly 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)

	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("expiration not within expected range, diff: %v", diff)
	}
}

func TestJWTService_Validate_WrongSigningMethod(t *testing.T) {
	// Create a token with a different signing method (none)
	claims := &Claims{
		UserID:   "user-123",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	svc := NewJWTService("test-secret")

	_, err := svc.Validate(tokenString)
	if err == nil {
		t.Fatal("expected error for token with 'none' signing method")
	}
}
