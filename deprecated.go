package doraemon

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// Deprecated: use ListAllRecursively instead
//
// 递归获取path下所有文件和文件夹
// path决定返回的文件路径是绝对路径还是相对路径。
func GetAllNamesRecursive(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	var files []string
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// Deprecated: use ListDirsRecursively instead
//
// 递归获取path下所有文件夹(包含子文件夹)
// path决定返回的文件路径是绝对路径还是相对路径。
func GetFolderNamesRecursive(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	dirs := make([]string, 0)
	err := filepath.Walk(path, func(childPath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		if childPath == path {
			return nil
		}
		dirs = append(dirs, childPath)
		return nil
	})
	return dirs, err
}

// Deprecated: 使用 ListFilesRecursively 代替
//
// 递归获取path下所有文件(包含子文件夹中的文件)。
// path决定返回的文件路径是绝对路径还是相对路径。
func GetFileNamesRecursive(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	files := make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// Deprecated: use SHA1 instead
func GetSha1(data []byte) ([]byte, error) {
	hash := sha1.New()
	if _, err := hash.Write(data); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// Deprecated: use MD5 instead
func GetMd5(content []byte) ([]byte, error) {
	hash := md5.New()
	_, err := hash.Write(content)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// Deprecated: use MD5Hex instead
func GetMd5Hex(content []byte) (string, error) {
	md5, err := GetMd5(content)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5), nil
}

// Deprecated: use SHA256 instead
func GetSha256(content []byte) ([]byte, error) {
	hash := sha256.New()
	_, err := hash.Write(content)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
