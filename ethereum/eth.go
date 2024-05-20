package ethereum

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenRandomETHKey() (privateKey *ecdsa.PrivateKey, err error) {
	return crypto.GenerateKey()
}
