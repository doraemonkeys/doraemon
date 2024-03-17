package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptHash 对传入字符串进行bcrypt哈希
func BcryptHash(str string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// BcryptMatch 对传入字符串和哈希字符串进行比对,str为明文
func BcryptMatch(hash string, str string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(str))
	return err == nil
}
