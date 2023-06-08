package doraemon

import (
	"crypto/rand"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
)

// https://github.com/tjfoc/gmsm/blob/master/API%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E.md

// SM1 为对称加密，该算法的算法实现原理没有公开，他的加密强度和 AES 相当，
// 需要调用加密芯片的接口进行使用。SM1 高达 128 bit 的密钥长度以及算法本身的强度和不公开性保证了通信的安全性。

// 国密 SM2 为非对称加密，也称为公钥密码；国密算法的本质是椭圆曲线加密。
// SM2 椭圆曲线公钥密码 ( ECC ) 算法是我国公钥密码算法标准。

// 在所有的公钥密码中，使用得比较广泛的有ECC 和 RSA；
// 而在相同安全强度下 ECC 比 RSA 的私钥位长及系统参数小得多，
// 这意味着应用 ECC 所需的存储空间要小得多， 传输所的带宽要求更低，
// 硬件实现 ECC 所需逻辑电路的逻辑门数要较 RSA 少得多， 功耗更低。
// 这使得 ECC 比 RSA 更适合实现到资源严重受限制的设备中。

// SM3密码杂凑算法是中国国家密码管理局2010年公布的中国商用密码杂凑算法标准。
// SM3算法是在SHA-256基础上改进的一种算法，消息分组的长度为512位，生成的摘要长度为256位，与SHA256安全性相当。

// SM4 是一种 Feistel 结构的分组密码算法，其分组长度和密钥长度均为128bit。
// 加解密算法与密钥扩张算法都采用 32 轮非线性迭代结构。
// 解密算法与加密算法的结构相同，只是轮密钥的使用顺序相反，即解密算法使用的轮密钥是加密算法使用的轮密钥的逆序。
// SM4 密码算法是中国第一次由专业密码机构公布并设计的商用密码算法，
// 到目前为止，尚未发现有任何攻击方法对SM4 算法的安全性产生威胁。

func GenerateSM2Key() (pubKey, priKey []byte, err error) {
	priv, err := sm2.GenerateKey(rand.Reader) // 生成密钥对
	if err != nil {
		return nil, nil, err
	}
	pub := &priv.PublicKey
	pubKey = sm2.Compress(pub) // 生成压缩公钥
	priKey = priv.D.Bytes()    // 生成私钥
	return pubKey, priKey, nil
}

// 返回asn.1编码格式的密文内容
func SM2EncryptAsn1(pubKey, plainText []byte) (cipherText []byte, err error) {
	pub := sm2.Decompress(pubKey)
	cipherText, err = pub.EncryptAsn1(plainText, rand.Reader)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

func SM2DecryptAsn1(priKey, cipherText []byte) (plainText []byte, err error) {
	priv := new(sm2.PrivateKey)
	priv.Curve = sm2.P256Sm2()
	priv.D = new(big.Int).SetBytes(priKey)
	plainText, err = priv.DecryptAsn1(cipherText)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}
