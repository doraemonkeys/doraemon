package doraemon

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
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
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, plainText)
	if err != nil {
		return nil, err
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
	cipherText, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPublicKey, plainText, label)
	if err != nil {
		return nil, err
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
	plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, cipherText)
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
	//plainText, _ := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPrivateKey, cipherText, label)
}
