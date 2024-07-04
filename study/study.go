package study

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

// StrLen统计字符串长度
func StrLen(str string) int {
	return len([]rune(str))
}

func GenRandomETHKey() (privateKey *ecdsa.PrivateKey, err error) {
	return crypto.GenerateKey()
}
