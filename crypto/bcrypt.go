package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

func BcryptHash2(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
}

func BcryptMatch2(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
