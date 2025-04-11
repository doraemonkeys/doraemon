package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	_ SymmetricCipher = &AESCBC{}
	_ SymmetricCipher = &AESCBC2{}
	_ SymmetricCipher = &AESGCM{}
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
	// dst and src must overlap entirely or not at all.
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

type AESGCM struct {
	cipher.AEAD
	key []byte
}

func NewAESGCM(key []byte) (AESGCM, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return AESGCM{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return AESGCM{}, err
	}
	return AESGCM{AEAD: gcm, key: key}, nil
}

func NewAESGCMFromHex(hexKey string) (AESGCM, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return AESGCM{}, err
	}
	return NewAESGCM(key)
}

func (c AESGCM) Encrypt(data []byte) ([]byte, error) {
	return c.encryptAuth(data)
}

func (c AESGCM) EncryptAuth(data []byte, additionalData ...[]byte) ([]byte, error) {
	return c.encryptAuth(data, additionalData...)
}

func (c AESGCM) encryptAuth(data []byte, additionalData ...[]byte) ([]byte, error) {
	nonce := make([]byte, c.AEAD.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	additional := bytes.Join(additionalData, nil)
	ciphertext := c.AEAD.Seal(nonce, nonce, data, additional)
	return ciphertext, nil
}

// EncryptWithNonce encrypts data with a nonce, the nonce is the first part of the plaintext(12 bytes).
//
// Data storage will not be reused, so the data will not be modified.
// At least in go1.24, this is the case.
func (c AESGCM) EncryptWithNonce(data []byte, additionalData ...[]byte) ([]byte, error) {
	if len(data) < c.AEAD.NonceSize() {
		return nil, errors.New("data is too short")
	}
	nonce := data[:c.AEAD.NonceSize()]
	ciphertext := c.AEAD.Seal(nonce, nonce, data[c.AEAD.NonceSize():], bytes.Join(additionalData, nil))
	return ciphertext, nil
}

// Decrypt decrypts data with a nonce, the nonce is the first part of the plaintext(12 bytes).
//
// Data storage will be reused, so the data will be modified.
func (c AESGCM) Decrypt(data []byte) ([]byte, error) {
	return c.decryptAuth(data)
}

// DecryptAuth decrypts data with a nonce, the nonce is the first part of the plaintext(12 bytes).
//
// Data storage will be reused, so the data will be modified.
func (c AESGCM) DecryptAuth(data []byte, additionalData ...[]byte) ([]byte, error) {
	return c.decryptAuth(data, additionalData...)
}

func (c AESGCM) decryptAuth(data []byte, additionalData ...[]byte) ([]byte, error) {
	nonceSize := c.AEAD.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	additional := bytes.Join(additionalData, nil)

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := c.AEAD.Open(ciphertext[:0], nonce, ciphertext, additional)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
