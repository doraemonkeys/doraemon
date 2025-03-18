package crypto

import (
	"encoding/hex"
	"errors"

	"github.com/wumansgy/goEncrypt/aes"
)

// 初始化向量（IV，Initialization Vector）是许多任务作模式中用于将加密随机化的一个位块，
// 由此即使同样的明文被多次加密也会产生不同的密文，避免了较慢的重新产生密钥的过程。
// 对于CBC和CFB，重用IV会导致泄露明文首个块的某些信息，亦包括两个不同消息中相同的前缀。
// 对于OFB和CTR而言，重用IV会导致完全失去安全性。另外，在CBC模式中，IV在加密时必须是无法预测的；
//
// AES CBC加密时，用iv和key去加密第一个块，然后用第一个块的加密数据作为下一个块的iv，依次迭代。
// 解密时，用n-1个块的加密数据作为iv和key去解密第n个块（n>1），只有第一个块才会用加密时的iv去解密第一个块。
// 所以如果解密时，使用了错误的iv，出问题的数据只有第一个块。
//
// 分组加密在对明文加密的时候，并不是把整个明文一股脑加密成一整段密文， 而是把明文拆分成一个个独立的明文块，
// AES的每一个明文块长度128bit(16字节)，则记块长度为BlockSize = 16。
// 初始化向量IV是一个长度为BlockSize的随机字节数组，对应一个伪随机数。
type CbcAESCrypt struct {
	// contains filtered or unexported fields
	secretKey []byte
}

// Deprecated
//
// NewAESCryptFromHex 创建AES加密器, HexSecretKey为16进制字符串
// CBC模式，PKCS5填充
func NewAESCryptFromHex(HexSecretKey string) (*CbcAESCrypt, error) {
	// 128, 192, or 256 bits
	if len(HexSecretKey) != 32 && len(HexSecretKey) != 48 && len(HexSecretKey) != 64 {
		return nil, errors.New("HexSecretKey length must be 32, 48 or 64")
	}
	secretKey, err := hex.DecodeString(HexSecretKey)
	return &CbcAESCrypt{secretKey: secretKey}, err
}

// Deprecated
//
// NewAESCryptFromHex 创建AES加密器。
// 第三方库使用 CBC模式，PKCS5填充。
func NewAESCrypt(SecretKey []byte) (*CbcAESCrypt, error) {
	// 128, 192, or 256 bits
	if len(SecretKey) != 16 && len(SecretKey) != 24 && len(SecretKey) != 32 {
		return nil, errors.New("SecretKey length must be 16, 24 or 32")
	}
	return &CbcAESCrypt{secretKey: SecretKey}, nil
}

// Deprecated
//
// Encrypt 加密后返回 密文+16字节IV。
func (a *CbcAESCrypt) Encrypt(plainText []byte) ([]byte, error) {
	if len(plainText) == 0 {
		return nil, errors.New("plainText is empty")
	}
	IV := a.rand16Byte()
	rawCipherText, err := aes.AesCbcEncrypt(plainText, a.secretKey, IV)
	if err != nil {
		return nil, err
	}
	rawCipherText = append(rawCipherText, IV...)
	return rawCipherText, nil
}

// Deprecated
//
// Decrypt 解密，cipherText 为 密文+16字节IV。
func (a *CbcAESCrypt) Decrypt(cipherText []byte) ([]byte, error) {
	if len(cipherText) <= 16 {
		return nil, errors.New("cipherTextHex length must be greater than 16")
	}
	IV := cipherText[len(cipherText)-16:]
	cipherText = cipherText[:len(cipherText)-16]
	return aes.AesCbcDecrypt(cipherText, a.secretKey, IV)
}

func (a *CbcAESCrypt) rand16Byte() []byte {
	return RandNByte(16)
}
