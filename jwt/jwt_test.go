package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"slices"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewHS256JWT(t *testing.T) {
	secretKey := []byte("secret")
	jwtInstance, _ := NewHS256JWT[string](secretKey)
	assert.NotNil(t, jwtInstance)
	assert.Equal(t, jwt.SigningMethodHS256, jwtInstance.SigningAlgo)
	assert.Equal(t, secretKey, jwtInstance.SecretKey)
}

func TestNewES256JWT(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	jwtInstance, _ := NewES256JWT[string](privateKey)
	assert.NotNil(t, jwtInstance)
	assert.Equal(t, jwt.SigningMethodES256, jwtInstance.SigningAlgo)
	assert.Equal(t, privateKey, jwtInstance.SecretKey)
}

func TestCreateToken(t *testing.T) {
	secretKey := []byte("secret")
	jwtInstance, _ := NewHS256JWT[string](secretKey)
	signInfo := "signInfo"
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test",
	}

	tokenString, err := jwtInstance.CreateToken(signInfo, claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestCreateDefaultToken(t *testing.T) {
	secretKey := []byte("secret")
	jwtInstance, _ := NewHS256JWT[string](secretKey)
	signInfo := "signInfo"
	expiresAt := time.Now().Add(time.Hour)

	tokenString, err := jwtInstance.CreateDefaultToken(signInfo, expiresAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	parsedClaims, err := jwtInstance.ParseToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, signInfo, parsedClaims.SignInfo)

	signInfo2 := "signInfo2"
	tokenString2, err := jwtInstance.CreateDefaultToken(signInfo2, expiresAt)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString2)

	parsedClaims2, err := jwtInstance.ParseToken(tokenString2)
	assert.NoError(t, err)
	assert.Equal(t, signInfo2, parsedClaims2.SignInfo)
}

func TestParseToken(t *testing.T) {
	secretKey := []byte("secret")
	jwtInstance, _ := NewHS256JWT[string](secretKey)
	signInfo := "signInfo"
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test",
	}

	tokenString, err := jwtInstance.CreateToken(signInfo, claims)
	assert.NoError(t, err)

	parsedClaims, err := jwtInstance.ParseToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, signInfo, parsedClaims.SignInfo)
	assert.Equal(t, claims.ExpiresAt, parsedClaims.ExpiresAt)
	assert.Equal(t, claims.Issuer, parsedClaims.Issuer)
}

func TestVerifyToken(t *testing.T) {
	secretKey := []byte("secret")
	jwtInstance, _ := NewHS256JWT[string](secretKey)
	signInfo := "signInfo"
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test",
		Audience:  []string{"audience"},
		Subject:   "subject",
	}

	tokenString, err := jwtInstance.CreateToken(signInfo, claims)
	assert.NoError(t, err)

	err = jwtInstance.VerifyToken(tokenString, signInfo, claims)
	assert.NoError(t, err)
}

func TestECDSAToken(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	jwtInstance, _ := NewES256JWT[string](privateKey)
	signInfo := "signInfo"
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test",
	}

	// Create token
	tokenString, err := jwtInstance.CreateToken(signInfo, claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Parse token
	parsedClaims, err := jwtInstance.ParseToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, signInfo, parsedClaims.SignInfo)
	assert.Equal(t, claims.ExpiresAt, parsedClaims.ExpiresAt)
	assert.Equal(t, claims.Issuer, parsedClaims.Issuer)

	// Verify token
	err = jwtInstance.VerifyToken(tokenString, signInfo, claims)
	assert.NoError(t, err)
}

func TestParseToken2(t *testing.T) {
	// Generate a new ECDSA private key for testing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	jwtService, _ := NewES256JWT[string](privateKey)

	// Create a token with a specific issuer
	claims := jwt.RegisteredClaims{
		Issuer:    "test-issuer",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	signInfo := "test-sign-info"
	tokenString, err := jwtService.CreateToken(signInfo, claims)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Parse the token
	parsedClaims, err := jwtService.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// Verify the issuer matches
	if parsedClaims.Issuer != claims.Issuer {
		t.Errorf("Expected issuer %v, got %v", claims.Issuer, parsedClaims.Issuer)
	}

	// Create another token with a different issuer
	claimsDifferentIssuer := jwt.RegisteredClaims{
		Issuer:    "different-issuer",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tokenStringDifferentIssuer, err := jwtService.CreateToken(signInfo, claimsDifferentIssuer)
	if err != nil {
		t.Fatalf("Failed to create token with different issuer: %v", err)
	}

	// Parse the token with the different issuer
	parsedClaimsDifferentIssuer, err := jwtService.ParseToken(tokenStringDifferentIssuer)
	if err != nil {
		t.Fatalf("Failed to parse token with different issuer: %v", err)
	}

	// Verify the issuer does not match the original issuer
	if parsedClaimsDifferentIssuer.Issuer == claims.Issuer {
		t.Errorf("Expected issuer to be different, got the same issuer %v", parsedClaimsDifferentIssuer.Issuer)
	}
}
func TestParseToken_IssuerMismatch(t *testing.T) {
	// Generate ECDSA key pair for testing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	// Create a new JWT instance with ES256 algorithm
	jwtInstance, _ := NewES256JWT[string](privateKey)

	// Define claims with a specific issuer
	claims := jwt.RegisteredClaims{
		Issuer:    "valid-issuer",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}

	// Create a token with the specified claims
	tokenString, err := jwtInstance.CreateToken("signInfo", claims)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Parse the token and verify the issuer is as expected
	parsedClaims, err := jwtInstance.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// Assert that the issuer matches the expected value
	if parsedClaims.Issuer != "valid-issuer" {
		t.Errorf("Expected issuer 'valid-issuer', got '%s'", parsedClaims.Issuer)
	}

	// Create a token with a different issuer
	claims.Issuer = "invalid-issuer"
	tokenString, err = jwtInstance.CreateToken("signInfo", claims)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Attempt to parse the token with the different issuer
	parsedClaims, err = jwtInstance.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// Assert that the issuer does not match the expected value
	if parsedClaims.Issuer == "valid-issuer" {
		t.Errorf("Expected issuer 'invalid-issuer', but got 'valid-issuer'")
	}
}

func TestVerifyTokenWithDifferentIssuer(t *testing.T) {
	// Generate a new ECDSA private key for ES256
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	// Create a new JWT instance with ES256 signing method
	jwtInstance, _ := NewES256JWT[string](privateKey)

	// Create a token with a specific issuer
	signInfo := "signInfo"
	claims := jwt.RegisteredClaims{
		Issuer:    "issuer1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tokenString, err := jwtInstance.CreateToken(signInfo, claims)
	assert.NoError(t, err)

	// Create a different claims object with a different issuer
	claimsDifferentIssuer := jwt.RegisteredClaims{
		Issuer: "issuer2",
	}

	// Verify the token with the different issuer claims
	err = jwtInstance.VerifyToken(tokenString, signInfo, claimsDifferentIssuer)
	assert.Error(t, err)
	assert.Equal(t, "invalid issuer", err.Error())
}

func TestVerifyTokenWithSameIssuer(t *testing.T) {
	// Generate a new ECDSA private key for ES256
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	// Create a new JWT instance with ES256 signing method
	jwtInstance, _ := NewES256JWT[string](privateKey)

	// Create a token with a specific issuer
	signInfo := "signInfo"
	claims := jwt.RegisteredClaims{
		Issuer:    "issuer1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tokenString, err := jwtInstance.CreateToken(signInfo, claims)
	assert.NoError(t, err)

	// Verify the token with the same issuer claims
	err = jwtInstance.VerifyToken(tokenString, signInfo, claims)
	assert.NoError(t, err)
}

// TestJWT tests the creation and verification of JWT tokens.
func TestJWT(t *testing.T) {
	type SignInfo struct {
		UserID string
		Role   string
	}
	// Generate ECDSA private key for testing
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	// Create a new JWT instance using ES256 algorithm
	jwtInstance, _ := NewES256JWT[SignInfo](ecdsaKey)

	// Define signInfo and claims
	signInfo := SignInfo{
		UserID: "12345",
		Role:   "admin",
	}
	expiresAt := time.Now().Add(1 * time.Hour)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Issuer:    "test_issuer",
		Subject:   "test_subject",
		Audience:  jwt.ClaimStrings{"test_audience"},
	}

	// Create token
	tokenString, err := jwtInstance.CreateToken(signInfo, claims)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Parse and verify token
	parsedClaims, err := jwtInstance.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// Verify signInfo
	if parsedClaims.SignInfo != signInfo {
		t.Fatalf("SignInfo mismatch: expected %v, got %v", signInfo, parsedClaims.SignInfo)
	}

	// Verify claims
	if !parsedClaims.ExpiresAt.Equal(claims.ExpiresAt.Time) {
		t.Fatalf("ExpiresAt mismatch: expected %v, got %v", claims.ExpiresAt, parsedClaims.ExpiresAt)
	}
	if parsedClaims.Issuer != claims.Issuer {
		t.Fatalf("Issuer mismatch: expected %v, got %v", claims.Issuer, parsedClaims.Issuer)
	}
	if parsedClaims.Subject != claims.Subject {
		t.Fatalf("Subject mismatch: expected %v, got %v", claims.Subject, parsedClaims.Subject)
	}
	if !slices.Equal(parsedClaims.Audience, claims.Audience) {
		t.Fatalf("Audience mismatch: expected %v, got %v", claims.Audience, parsedClaims.Audience)
	}
}
