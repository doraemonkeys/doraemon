package jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/exp/slices"
)

type JWT struct {
	SecretKey   any
	SigningAlgo jwt.SigningMethod
}

type CustomClaims struct {
	SignInfo []byte
	jwt.StandardClaims
}

var _ jwt.Claims = &CustomClaims{}

var (
	ErrInvalidSecretKey = errors.New("invalid secret key")
)

func NewJWT(secretKey any, signingAlgo jwt.SigningMethod) (*JWT, error) {
	switch signingAlgo.(type) {
	case *jwt.SigningMethodHMAC:
		_, ok := secretKey.([]byte)
		if !ok {
			return nil, ErrInvalidSecretKey
		}
	case *jwt.SigningMethodECDSA:
		_, ok := secretKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, ErrInvalidSecretKey
		}
	case *jwt.SigningMethodRSA:
		_, ok := secretKey.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrInvalidSecretKey
		}
	default:
	}

	return &JWT{
		SecretKey:   secretKey,
		SigningAlgo: signingAlgo,
	}, nil
}

func NewHS256JWT(secretKey []byte) *JWT {
	ret, _ := NewJWT(secretKey, jwt.SigningMethodHS256)
	return ret
}

func NewES256JWT(secretKey *ecdsa.PrivateKey) *JWT {
	ret, _ := NewJWT(secretKey, jwt.SigningMethodES256)
	return ret
}

func (j *JWT) CreateToken(signInfo []byte, claims jwt.StandardClaims) (string, error) {
	customClaims := &CustomClaims{
		SignInfo:       signInfo,
		StandardClaims: claims,
	}
	token := jwt.NewWithClaims(j.SigningAlgo, customClaims)
	return token.SignedString(j.SecretKey)
}

func (j *JWT) CreateDefaultToken(signInfo []byte, expiresAt int64) (string, error) {
	claims := jwt.StandardClaims{
		ExpiresAt: expiresAt,
	}
	return j.CreateToken(signInfo, claims)
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	return j.ParseTokenWithKeyFunc(tokenString, func(token *jwt.Token) (any, error) {
		switch j.SigningAlgo.(type) {
		case *jwt.SigningMethodHMAC:
			return j.SecretKey.([]byte), nil
		case *jwt.SigningMethodECDSA:
			return j.SecretKey.(*ecdsa.PrivateKey).Public(), nil
		case *jwt.SigningMethodRSA:
			return j.SecretKey.(*rsa.PrivateKey).Public(), nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", j.SigningAlgo)
		}
	})
}

func (j *JWT) ParseTokenWithKeyFunc(tokenString string, keyFunc func(token *jwt.Token) (any, error)) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		return keyFunc(t)
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*CustomClaims)
	if ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (j *JWT) VerifyTokenOnlySignInfo(tokenString string, signInfo []byte) error {
	claimsParsed, err := j.ParseToken(tokenString)
	if err != nil {
		return err
	}
	if !slices.Equal(claimsParsed.SignInfo, signInfo) {
		return fmt.Errorf("invalid signInfo")
	}
	return nil
}

func (j *JWT) VerifyToken(tokenString string, signInfo []byte, claims jwt.StandardClaims) error {
	claimsParsed, err := j.ParseToken(tokenString)
	if err != nil {
		return err
	}
	if !slices.Equal(claimsParsed.SignInfo, signInfo) {
		return fmt.Errorf("invalid signInfo")
	}
	if claimsParsed.Audience != claims.Audience {
		return fmt.Errorf("invalid audience")
	}
	if claimsParsed.Issuer != claims.Issuer {
		return fmt.Errorf("invalid issuer")
	}
	if claimsParsed.Subject != claims.Subject {
		return fmt.Errorf("invalid subject")
	}
	return nil
}
