package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

type AESCBC struct {
	cipher.Block
	key []byte
}

func NewAESCBC(key []byte) (AESCBC, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return AESCBC{}, err
	}
	return AESCBC{Block: block, key: key}, nil
}

func NewAESCBCFromHex(hexKey string) (AESCBC, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return AESCBC{}, err
	}
	return NewAESCBC(key)
}

func (c AESCBC) Encrypt(data []byte) ([]byte, error) {
	plaintext := PKCS7Padding(data, c.BlockSize())
	ciphertext := make([]byte, len(plaintext)+c.BlockSize())
	iv := ciphertext[:c.BlockSize()]
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(c.Block, iv)
	mode.CryptBlocks(ciphertext[c.BlockSize():], plaintext)
	return ciphertext, nil
}

func (c AESCBC) Decrypt(data []byte) (plaintext []byte, err error) {
	if len(data) < c.BlockSize() {
		return nil, errors.New("data is too short")
	}
	if len(data)%c.BlockSize() != 0 {
		return nil, errors.New("data is not a multiple of the block size")
	}
	defer func() {
		if e := recover(); e != nil {
			plaintext = nil
			err = fmt.Errorf("decrypt failed: %v", e)
		}
	}()
	iv := data[:c.BlockSize()]
	ciphertext := data[c.BlockSize():]
	mode := cipher.NewCBCDecrypter(c.Block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	return PKCS7UnPadding(ciphertext, c.BlockSize())
}

// AESCBC2 puts iv at the end of the ciphertext
type AESCBC2 struct {
	cipher.Block
	key []byte
}

func NewAESCBC2(key []byte) (AESCBC2, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return AESCBC2{}, err
	}
	return AESCBC2{Block: block, key: key}, nil
}

func NewAESCBCFromHex2(hexKey string) (AESCBC2, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return AESCBC2{}, err
	}
	return NewAESCBC2(key)
}

func (c AESCBC2) Encrypt(data []byte) ([]byte, error) {
	plaintext := PKCS7Padding(data, c.BlockSize())
	ciphertext := make([]byte, len(plaintext)+c.BlockSize())
	iv := ciphertext[len(ciphertext)-c.BlockSize():]
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(c.Block, iv)
	mode.CryptBlocks(ciphertext[:len(ciphertext)-c.BlockSize()], plaintext)
	return ciphertext, nil
}

func (c AESCBC2) Decrypt(data []byte) (plaintext []byte, err error) {
	if len(data) < c.BlockSize() {
		return nil, errors.New("data is too short")
	}
	if len(data)%c.BlockSize() != 0 {
		return nil, errors.New("data is not a multiple of the block size")
	}
	defer func() {
		if e := recover(); e != nil {
			plaintext = nil
			err = fmt.Errorf("decrypt failed: %v", e)
		}
	}()
	iv := data[len(data)-c.BlockSize():]
	ciphertext := data[:len(data)-c.BlockSize()]
	mode := cipher.NewCBCDecrypter(c.Block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)
	return PKCS7UnPadding(ciphertext, c.BlockSize())
}
