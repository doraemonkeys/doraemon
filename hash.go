package doraemon

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func ComputeSHA1(content io.Reader) ([]byte, error) {
	hash := sha1.New()
	_, err := io.Copy(hash, content)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func ComputeSHA1Hex(content io.Reader) (string, error) {
	sha1, err := ComputeSHA1(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha1), nil
}

func ComputeFileSha1(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ComputeSHA1(file)
}

func ComputeMD5(content io.Reader) ([]byte, error) {
	hash := md5.New()
	_, err := io.Copy(hash, content)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func ComputeMD5Hex(content io.Reader) (string, error) {
	md5, err := ComputeMD5(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5), nil
}

func ComputeFileMd5(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ComputeMD5(file)
}

func ComputeSHA256(content io.Reader) ([]byte, error) {
	hash := sha256.New()
	_, err := io.Copy(hash, content)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func ComputeSHA256Hex(content io.Reader) (string, error) {
	sha256, err := ComputeSHA256(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256), nil
}
