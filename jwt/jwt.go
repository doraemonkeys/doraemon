package jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"slices"

	"github.com/golang-jwt/jwt/v5"
)

type JWT[T comparable] struct {
	secretKey   any
	signingAlgo jwt.SigningMethod
}

type CustomClaims[T comparable] struct {
	SignInfo T
	jwt.RegisteredClaims
}

var _ jwt.Claims = &CustomClaims[any]{}

var (
	ErrInvalidSecretKey = errors.New("invalid secret key")
)

func NewJWT[T comparable](secretKey any, signingAlgo jwt.SigningMethod) (*JWT[T], error) {
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

	return &JWT[T]{
		secretKey:   secretKey,
		signingAlgo: signingAlgo,
	}, nil
}

func NewHS256JWT[T comparable](secretKey []byte) (*JWT[T], error) {
	return NewJWT[T](secretKey, jwt.SigningMethodHS256)

}

func NewES256JWT[T comparable](secretKey *ecdsa.PrivateKey) (*JWT[T], error) {
	return NewJWT[T](secretKey, jwt.SigningMethodES256)
}

func (j *JWT[T]) CreateToken(signInfo T, claims jwt.RegisteredClaims) (string, error) {
	customClaims := &CustomClaims[T]{
		SignInfo:         signInfo,
		RegisteredClaims: claims,
	}
	token := jwt.NewWithClaims(j.signingAlgo, customClaims)
	return token.SignedString(j.secretKey)
}

func (j *JWT[T]) CreateDefaultToken(signInfo T, expiresAt time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	return j.CreateToken(signInfo, claims)
}

func (j *JWT[T]) ParseToken(tokenString string) (*CustomClaims[T], error) {
	return j.ParseTokenWithKeyFunc(tokenString, func(j *JWT[T], token *jwt.Token) (any, error) {
		switch j.signingAlgo.(type) {
		case *jwt.SigningMethodHMAC:
			return j.secretKey.([]byte), nil
		case *jwt.SigningMethodECDSA:
			return j.secretKey.(*ecdsa.PrivateKey).Public(), nil
		case *jwt.SigningMethodRSA:
			return j.secretKey.(*rsa.PrivateKey).Public(), nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", j.signingAlgo)
		}
	})
}

func (j *JWT[T]) ParseTokenWithKeyFunc(tokenString string, keyFunc func(j *JWT[T], token *jwt.Token) (any, error)) (*CustomClaims[T], error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims[T]{}, func(t *jwt.Token) (any, error) {
		return keyFunc(j, t)
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*CustomClaims[T])
	if ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (j *JWT[T]) VerifyTokenOnlySignInfo(tokenString string, signInfo T) error {
	claimsParsed, err := j.ParseToken(tokenString)
	if err != nil {
		return err
	}
	if claimsParsed.SignInfo != signInfo {
		return fmt.Errorf("invalid signInfo")
	}
	return nil
}

func (j *JWT[T]) VerifyToken(tokenString string, signInfo T, claims jwt.RegisteredClaims) error {
	claimsParsed, err := j.ParseToken(tokenString)
	if err != nil {
		return err
	}
	if claimsParsed.SignInfo != signInfo {
		return fmt.Errorf("invalid signInfo")
	}
	if !slices.Equal(claimsParsed.Audience, claims.Audience) {
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
