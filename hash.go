package doraemon

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

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

// 获取文件的SHA1值(字母小写)
func GetFileSha1(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetSha1 获取[]byte的SHA1值(字母小写)
func GetSha1(data []byte) ([]byte, error) {
	hash := sha1.New()
	if _, err := hash.Write(data); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// 获取文件md5
func GetFileMd5(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	//将[]byte转成16进制的字符串表示
	//var hex string = "48656c6c6f"//(hello)
	//其中每两个字符对应于其ASCII值的十六进制表示,例如:
	//0x48 0x65 0x6c 0x6c 0x6f = "Hello"
	//fmt.Printf("%x\n", hash.Sum(nil))
	return hash.Sum(nil), nil
}

// 计算md5
func GatMd5(content []byte) ([]byte, error) {
	hash := md5.New()
	_, err := hash.Write(content)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func GatMd5Hex(content []byte) (string, error) {
	md5, err := GatMd5(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5), nil
}

// 计算sha256
func GetSha256(content []byte) ([]byte, error) {
	hash := sha256.New()
	_, err := hash.Write(content)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
