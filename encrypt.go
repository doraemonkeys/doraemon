package doraemon

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/wumansgy/goEncrypt/aes"
)

// Setup a bare-bones TLS config for the server
//
//	例如:添加quic协议支持
//	config.NextProtos = append(config.NextProtos, "quic")
func GenerateTLSConfig() (*tls.Config, error) {
	tlsCert, err := GenCertificate()
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}, nil
}

func LoadX509KeyPairTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert}, // 证书
	}
	return config, nil
}

func GenCertificate() (cert tls.Certificate, err error) {
	rawCert, rawKey, err := GenerateKeyPair()
	if err != nil {
		return
	}
	return tls.X509KeyPair(rawCert, rawKey)
}

func GenerateKeyPair() (rawCert, rawKey []byte, err error) {
	// Create private key and self-signed certificate
	// Adapted from https://golang.org/src/crypto/tls/generate_cert.go

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	validFor := time.Hour * 24 * 365 * 10 // ten years
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"doraemon"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return
	}

	rawCert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	rawKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return
}

// 生成RSA私钥和公钥，pem格式
func GenerateRSAKey(bits int) (Private []byte, Public []byte, err error) {
	//GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	//Reader是一个全局、共享的密码用强随机数生成器
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	//保存私钥
	//通过x509标准将得到的ras私钥序列化为ASN.1 的 DER编码字符串
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	//使用pem格式对x509输出的内容进行编码
	//创建文件保存私钥
	// privateFile, err := os.Create("private.pem")
	// if err != nil {
	// 	panic(err)
	// }
	// defer privateFile.Close()
	//构建一个pem.Block结构体对象
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	//将数据保存到文件
	//pem.Encode(privateFile, &privateBlock)

	//保存私钥
	Private = pem.EncodeToMemory(&privateBlock)

	//保存公钥
	//获取公钥的数据
	publicKey := privateKey.PublicKey
	//X509对公钥编码
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return nil, nil, err
	}
	//pem格式编码
	//创建用于保存公钥的文件
	// publicFile, err := os.Create("public.pem")
	// if err != nil {
	// 	panic(err)
	// }
	//defer publicFile.Close()
	//创建一个pem.Block结构体对象
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	//保存到文件
	//pem.Encode(publicFile, &publicBlock)

	//保存公钥
	Public = pem.EncodeToMemory(&publicBlock)
	return Private, Public, nil
}

// RSA加密,填充方式为PKCS1v15,publicKey为公钥的pem格式。
// 对结果进行base64编码可以提高可读性。
//
// The message must be no longer than the length of the public modulus minus 11 bytes.
//
// WARNING: use of this function to encrypt plaintexts other than session keys is dangerous. Use RSA OAEP in new protocols.
//
//	在RSA攻击中，存在着“小明文攻击“的方式；
//	在明文够小时，密文也够小，直接开e次方即可；
//	在明文有点小时，如果e也较小，可用pow(m,e)=n*k+c穷举k尝试爆破，所以，比如说，在选择明文攻击中，单纯的RSA非常容易被破解。
//	于是，我们就像将密文进行一下填充，最好让密文都等长。
//	但是填充方式也是很讲究的；不好的填充规则往往仅仅有限的增加了攻击的难度，或者难以实现等长密文。
//	于是我们就查到了OAEP——最优非对称加密填充。
func RSA_EncryptPKCS1v15(plainText []byte, publicKey []byte) ([]byte, error) {
	//pem解码
	block, _ := pem.Decode(publicKey)
	//x509解码
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//类型断言
	rsaPublicKey := publicKeyInterface.(*rsa.PublicKey)

	//对明文进行加密
	msgLen := len(plainText)
	step := rsaPublicKey.Size() - 11
	var cipherText []byte
	for i := 0; i < msgLen; i += step {
		end := i + step
		if end > msgLen {
			end = msgLen
		}
		tmp, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, plainText[i:end])
		if err != nil {
			return nil, err
		}
		cipherText = append(cipherText, tmp...)
	}
	//返回密文
	return cipherText, nil
}

// RSA加密,填充方式为OAEP，publicKey为公钥的pem格式。
// 对结果进行base64编码可以提高可读性。
//
// The random parameter is used as a source of entropy to ensure that
// encrypting the same message twice doesn't result in the same ciphertext.
//
// The label parameter may contain arbitrary data that will not be encrypted,
// but which gives important context to the message. For example, if a given
// public key is used to encrypt two types of messages then distinct label
// values could be used to ensure that a ciphertext for one purpose cannot be
// used for another by an attacker. If not required it can be empty.
func RSA_EncryptOAEP(plainText []byte, publicKey []byte, label []byte) ([]byte, error) {
	//pem解码
	block, _ := pem.Decode(publicKey)
	//x509解码
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//类型断言
	rsaPublicKey := publicKeyInterface.(*rsa.PublicKey)

	//对明文进行加密
	msgLen := len(plainText)
	hash := sha256.New()
	step := rsaPublicKey.Size() - 2*hash.Size() - 2
	// fmt.Println("step:", step)
	var cipherText []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, rand.Reader, rsaPublicKey, plainText[start:finish], label)
		if err != nil {
			return nil, err
		}
		cipherText = append(cipherText, encryptedBlockBytes...)
	}
	//返回密文
	return cipherText, nil
}

// RSA解密,填充模式PKCS1v15,private为私钥的pem格式
func RSA_DecryptPKCS1v15(cipherText []byte, privateKey []byte) ([]byte, error) {
	//pem解码
	block, _ := pem.Decode(privateKey)
	//X509解码
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//对密文进行解密
	msgLen := len(cipherText)
	step := rsaPrivateKey.Size()
	// fmt.Println("step:", step)
	var plainText []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		decryptedBlockBytes, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, cipherText[start:finish])
		if err != nil {
			return nil, err
		}
		plainText = append(plainText, decryptedBlockBytes...)
	}
	//返回明文
	return plainText, nil
}

// RSA解密,填充模式OAEP,private为私钥的pem格式
func RSA_DecryptOAEP(cipherText []byte, privateKey []byte, label []byte) ([]byte, error) {
	//pem解码
	block, _ := pem.Decode(privateKey)
	//X509解码
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//对密文进行解密
	msgLen := len(cipherText)
	hash := sha256.New()
	step := rsaPrivateKey.Size()
	// fmt.Println("step:", step)
	var plainText []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, rand.Reader, rsaPrivateKey, cipherText[start:finish], label)
		if err != nil {
			return nil, err
		}
		plainText = append(plainText, decryptedBlockBytes...)
	}
	//返回明文
	return plainText, nil
}

// rsa数字签名，private为私钥的pem格式
func RsaSignPKCS1v15(src []byte, privateKey []byte) ([]byte, error) {
	// todo 获取私钥
	//pem解码
	block, _ := pem.Decode(privateKey)
	//X509解码
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 签名
	shaNew := sha256.New()
	_, err = shaNew.Write(src)
	if err != nil {
		return nil, err
	}
	shaByte := shaNew.Sum(nil)
	v15, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, shaByte)
	if err != nil {
		return nil, err
	}
	return v15, nil
}

// rsa数字验签，publicKey为公钥的pem格式
func RsaVerifyPKCS1v15(src []byte, sign []byte, publicKey []byte) error {
	// todo 获取公钥
	//pem解码
	block, _ := pem.Decode(publicKey)
	//X509解码
	rsaPublicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	//类型断言
	rsaPublicKey := rsaPublicKeyInterface.(*rsa.PublicKey)
	// 验签
	shaNew := sha256.New()
	_, err = shaNew.Write(src)
	if err != nil {
		return err
	}
	shaByte := shaNew.Sum(nil)
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, shaByte, sign)
	if err != nil {
		return err
	}
	return nil
}

// GenerateRandomSeed 生成随机种子
func GenerateRandomSeed() int64 {
	// Generate a random number between 0 and 2^63 - 1
	// max := new(big.Int).Exp(big.NewInt(2), big.NewInt(63), nil)
	// max.Sub(max, big.NewInt(1))
	n, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return n.Int64()
}

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

// NewAESCryptFromHex 创建AES加密器。
// 第三方库使用 CBC模式，PKCS5填充。
func NewAESCrypt(SecretKey []byte) (*CbcAESCrypt, error) {
	// 128, 192, or 256 bits
	if len(SecretKey) != 16 && len(SecretKey) != 24 && len(SecretKey) != 32 {
		return nil, errors.New("SecretKey length must be 16, 24 or 32")
	}
	return &CbcAESCrypt{secretKey: SecretKey}, nil
}

// Encrypt 加密后返回 密文+16字节IV。
func (a *CbcAESCrypt) Encrypt(plainText []byte) ([]byte, error) {
	if len(plainText) == 0 {
		return nil, errors.New("plainText is empty")
	}
	IV := a.rand16Byte()
	rawCipherTextHex, err := aes.AesCbcEncrypt(plainText, a.secretKey, IV)
	if err != nil {
		return nil, err
	}
	rawCipherTextHex = append(rawCipherTextHex, IV...)
	return rawCipherTextHex, nil
}

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
	return randNByte(16)
}

// randNByte returns a slice of n random bytes.
func randNByte(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
func PKCS5Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - (len(plainText) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	newText := append(plainText, padText...)
	return newText
}

func PKCS5UnPadding(plainText []byte, blockSize int) ([]byte, error) {
	length := len(plainText)
	number := int(plainText[length-1])
	if number >= length || number > blockSize {
		return nil, errors.New("invalid plaintext")
	}
	return plainText[:length-number], nil
}
