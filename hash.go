package doraemon

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func ComputeSHA1(content io.Reader) Result[[]byte] {
	hash := sha1.New()
	_, err := io.Copy(hash, content)
	if err != nil {
		return Err[[]byte](err)
	}
	return Ok(hash.Sum(nil))
}

func ComputeSHA1Hex(content io.Reader) Result[string] {
	sha1 := ComputeSHA1(content)
	if sha1.IsErr() {
		return Err[string](sha1.Err)
	}
	return Ok(hex.EncodeToString(sha1.Value))
}

func ComputeFileSha1(filename string) Result[[]byte] {
	file, err := os.Open(filename)
	if err != nil {
		return Err[[]byte](err)
	}
	defer file.Close()
	return ComputeSHA1(file)
}

func ComputeMD5(content io.Reader) Result[[]byte] {
	hash := md5.New()
	_, err := io.Copy(hash, content)
	if err != nil {
		return Err[[]byte](err)
	}
	return Ok(hash.Sum(nil))
}

func ComputeMD5Hex(content io.Reader) Result[string] {
	md5 := ComputeMD5(content)
	if md5.IsErr() {
		return Err[string](md5.Err)
	}
	return Ok(hex.EncodeToString(md5.Value))
}

func ComputeFileMd5(filename string) Result[[]byte] {
	file, err := os.Open(filename)
	if err != nil {
		return Err[[]byte](err)
	}
	defer file.Close()
	return ComputeMD5(file)
}

func ComputeSHA256(content io.Reader) Result[[]byte] {
	hash := sha256.New()
	_, err := io.Copy(hash, content)
	if err != nil {
		return Err[[]byte](err)
	}
	return Ok(hash.Sum(nil))
}

func ComputeSHA256Hex(content io.Reader) Result[string] {
	sha256 := ComputeSHA256(content)
	if sha256.IsErr() {
		return Err[string](sha256.Err)
	}
	return Ok(hex.EncodeToString(sha256.Value))
}
