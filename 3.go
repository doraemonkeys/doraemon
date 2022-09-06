package doraemon

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//获取当前文件夹下所有文件
func GetFiles(path string) []string {
	files := make([]string, 0)
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files
}

//获取当前文件夹下所有文件夹
func GetDirs(path string) []string {
	dirs := make([]string, 0)
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	return dirs
}

//获取当前文件夹下所有文件和文件夹
func GetAll(path string) ([]string, error) {
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

//获取当前程序的执行路径(包含可执行文件名称)
//C:\Users\*\AppData\Local\Temp\*\exe\main.exe
//(读取命令行的方式，可能得不到想要的路径)
func GetCurrentPath() (string, error) {
	//LookPath 在 PATH 环境变量命名的目录中搜索可执行文件。如果文件包含斜杠，则直接尝试，不参考 PATH。
	s, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

//获取当前程序源代码的详细路径
//D:/Go/workspace/port/network_learn/server/server.go
func CurrentFilePath() (string, error) {
	//报告当前go程序调用栈所执行的函数的文件和行号信息
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("can not get current file info")
	}
	return file, nil
}

//获取文件md5
func GetFileMd5(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	//将[]byte转成16进制的字符串表示
	//var hex string = "48656c6c6f"//(hello)
	//其中每两个字符对应于其ASCII值的十六进制表示,例如:
	//0x48 0x65 0x6c 0x6c 0x6f = "Hello"
	//fmt.Printf("%x\n", hash.Sum(nil))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

//从文件末尾按行读取文件。
//name:文件路径 lineNum:读取行数(超过文件行数则读取全文)。
//最后一行为空也算读取了一行,会返回此行为空串,若全是空格也会原样返回。
//返回的每一行都不包含换行符号。
func ReverseRead(name string, lineNum uint) ([]string, error) {
	//打开文件
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	//获取文件大小
	fs, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fs.Size()

	var offset int64 = -1   //偏移量，初始化为-1，若为0则会读到EOF
	char := make([]byte, 1) //用于读取单个字节
	lineStr := ""           //存放一行的数据
	buff := make([]string, 0, 100)
	for (-offset) <= fileSize {
		//通过Seek函数从末尾移动游标然后每次读取一个字节，offset为偏移量
		file.Seek(offset, io.SeekEnd)
		_, err := file.Read(char)
		if err != nil {
			return buff, err
		}
		if char[0] == '\n' {
			//判断文件类型为unix(LF)还是windows(CRLF)
			file.Seek(-2, io.SeekCurrent) //io.SeekCurrent表示游标放置于当前位置，逆向偏移2个字节
			//读完一个字节后游标会自动正向偏移一个字节
			file.Read(char)
			if char[0] == '\r' {
				offset-- //windows跳过'\r'
			}
			lineNum-- //到此读取完一行
			buff = append(buff, lineStr)
			lineStr = ""
			if lineNum == 0 {
				return buff, nil
			}
		} else {
			lineStr = string(char) + lineStr
		}
		offset--
	}
	buff = append(buff, lineStr)
	return buff, nil
}

//读取倒数第n行(n从1开始),
//若n大于文件行数则返回错误io.EOF。
func ReadStartWithLastLine(filename string, n int) (string, error) {
	//打开文件
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	//获取文件大小
	fs, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := fs.Size()

	var offset int64 = -1   //偏移量，初始化为-1，若为0则会读到EOF
	char := make([]byte, 1) //用于读取单个字节
	lineStr := ""           //存放一行的数据
	lineCount := 0          //行数
	for (-offset) <= fileSize {
		//通过Seek函数从末尾移动游标然后每次读取一个字节，offset为偏移量
		file.Seek(offset, io.SeekEnd)
		_, err := file.Read(char)
		if err != nil {
			return "", err
		}
		if char[0] == '\n' {
			lineCount++
			if lineCount == n {
				return lineStr, nil
			}
			//判断文件类型为unix(LF)还是windows(CRLF)
			file.Seek(-2, io.SeekCurrent) //io.SeekCurrent表示游标放置于当前位置，逆向偏移2个字节
			//读完一个字节后游标会自动正向偏移一个字节
			file.Read(char)
			if char[0] == '\r' {
				offset-- //windows跳过'\r'
			}
			offset--
			continue
		}
		if lineCount == n-1 {
			lineStr = string(char) + lineStr
		}
		offset--
	}
	if lineStr == "" {
		return "", io.EOF
	}
	return lineStr, nil
}
