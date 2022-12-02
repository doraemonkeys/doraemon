package doraemon

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// 递归获取path下所有文件(包含子文件夹中的文件)。
// path决定返回的文件路径是绝对路径还是相对路径。
func GetFiles(path string) ([]string, error) {
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

// 递归获取path下所有文件夹(包含子文件夹)
// path决定返回的文件路径是绝对路径还是相对路径。
func GetDirs(path string) ([]string, error) {
	dirs := make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	return dirs, err
}

// 递归获取path下所有文件和文件夹
// path决定返回的文件路径是绝对路径还是相对路径。
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

// 获取path下所有文件名称(含后缀)
func GetFileNmaes(path string) ([]string, error) {
	DirEntry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, v := range DirEntry {
		if !v.IsDir() {
			files = append(files, v.Name())
		}
	}
	return files, nil
}

// 获取path路径下的文件夹名称
func GetDirNames(path string) ([]string, error) {
	DirEntry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, v := range DirEntry {
		if v.IsDir() {
			dirs = append(dirs, v.Name())
		}
	}
	return dirs, nil
}

// 获取path路径下的文件(含后缀)和文件夹名称
func GetAllNames(path string) ([]string, error) {
	DirEntry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var all []string
	for _, v := range DirEntry {
		all = append(all, v.Name())
	}
	return all, nil
}

// 文件或文件夹是否存在
func FileOrDirIsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 是否为文件夹
func IsDir(path string) (is bool, exist bool, err error) {
	Info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, true, err
	}
	return Info.IsDir(), true, nil
}

// 是否为文件
func IsFile(path string) (is bool, exist bool, err error) {
	if path[len(path)-1] == '/' || path[len(path)-1] == '\\' {
		path = path[:len(path)-1] //去除最后一个路径分隔符
	}
	Info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, true, err
	}
	return !Info.IsDir(), true, nil
}

// 复制文件到指定目录
//
// overwrite为true时，如果目标文件存在则覆盖，
// overwrite为false时，如果目标文件存在则返回错误。
// dst必须是一个存在的文件夹，否则返回错误。
// scr为文件的绝对或相对路径(包含文件名)。
func CopyFile(src, dst string, overwrite bool) error {
	if dst[len(dst)-1:] != `\` && dst[len(dst)-1:] != `/` {
		dst += `\` // 保证dst是文件夹
	}
	target := filepath.Join(dst, filepath.Base(src))
	if !overwrite && FileOrDirIsExist(target) {
		return fmt.Errorf("%s is exist", target)
	}
	dstFile, err := os.Create(target)
	if err != nil {
		is, _, _ := IsFile(dst)
		if is {
			return fmt.Errorf("%s is not a folder", dst[:len(dst)-1])
		}
		return err
	}
	defer dstFile.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

// 移动文件或文件夹到指定目录
//
// overwrite为true时，如果目标文件存在则覆盖(dst中的目标文件或文件夹会被直接删除)，
// overwrite为false时，如果目标文件存在则返回错误。
// dst必须是一个存在的文件夹，否则返回错误。
// scr为的绝对或相对路径。
func MoveFileOrDir(src, dst string, overwrite bool) error {
	if dst[len(dst)-1:] != `\` && dst[len(dst)-1:] != `/` {
		dst += `\`
	}
	//判断src是否存在
	if !FileOrDirIsExist(src) {
		return fmt.Errorf("%s is not exist", src)
	}
	target := filepath.Join(dst, filepath.Base(src))
	if FileOrDirIsExist(target) {
		if !overwrite {
			return fmt.Errorf("%s is exist", target)
		}
		err := os.RemoveAll(target)
		if err != nil {
			return err
		}
	}
	err := os.Rename(src, target)
	if err != nil {
		return err
	}
	return nil
}

// 复制文件夹到指定目录
//
// overwrite为true时，如果目标文件夹存在名字相同的文件则覆盖，
// overwrite为false时，如果目标文件存在则返回错误。
// dst,scr都必须是一个存在的文件夹，否则返回错误。
func CopyDir(src, dst string, overwrite bool) error {
	if dst[len(dst)-1:] != `\` && dst[len(dst)-1:] != `/` {
		dst += `\` //添加路径分隔符
	}
	if src[len(src)-1:] != `\` && src[len(src)-1:] != `/` {
		src += `\` //添加路径分隔符
	}
	dst = filepath.Join(dst, filepath.Base(src)) //dst更新为目标文件夹
	if FileOrDirIsExist(dst) {
		if !overwrite {
			return fmt.Errorf("%s is exist", dst)
		}
	} else {
		err := os.Mkdir(dst, 0666)
		if err != nil {
			return err
		}
	}
	//获取src下所有文件
	srcFiles, err := GetFileNmaes(src)
	if err != nil {
		return err
	}
	//复制文件
	for _, v := range srcFiles {
		srcFile := filepath.Join(src, v)
		err := CopyFile(srcFile, dst, overwrite)
		if err != nil {
			return err
		}
	}
	//获取src下所有文件夹
	srcDirNames, err := GetDirNames(src)
	if err != nil {
		return err
	}
	for _, v := range srcDirNames {
		srcDir := filepath.Join(src, v)
		err = CopyDir(srcDir, dst, overwrite)
		if err != nil {
			return err
		}
	}
	return nil
}

// 获取当前程序的执行路径(包含可执行文件名称)
// C:\Users\*\AppData\Local\Temp\*\exe\main.exe 或 .\main.exe
// (读取命令参数的方式)
func GetExecutionPath2() (string, error) {
	//LookPath 在 PATH 环境变量命名的目录中搜索可执行文件。如果文件包含斜杠，则直接尝试(返回相对路径)，不参考 PATH。
	path, err := exec.LookPath(os.Args[0])
	if errors.Is(err, exec.ErrDot) {
		// 说明是当前目录,参数是相对路径且不包含斜杠(./或.\)
		return os.Executable()
	}
	if err != nil {
		return "", err
	}
	//Abs 返回路径的绝对表示形式。
	return filepath.Abs(path)
}

// 获取当前程序的执行环境路径(不包含可执行文件名称)
func GetCurrentPath() (string, error) {
	return os.Getwd()
}

// 获取当前程序所在的绝对路径+文件名
func GetExecutionPath() (string, error) {
	return os.Executable()
}

// 获取当前程序源代码的详细路径
// D:/Go/workspace/port/network_learn/server/server.go
func ExecutionFilePath() (string, error) {
	//报告当前go程序调用栈所执行的函数的文件和行号信息
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("can not get file info")
	}
	return file, nil
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

// 获取文件md5(字母小写)
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

// 计算md5
func Md5(content []byte) string {
	hash := md5.New()
	hash.Write(content)
	return hex.EncodeToString(hash.Sum(nil))
}

// 从文件末尾按行读取文件。
// name:文件路径 lineNum:读取行数(超过文件行数则读取全文)。
// 最后一行为空也算读取了一行,会返回此行为空串,若全是空格也会原样返回。
// 返回的每一行都不包含换行符号。
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
			//防止偏移量-2后越界
			if fileSize-(-offset) >= 1 {
				//判断文件类型为unix(LF)还是windows(CRLF)
				file.Seek(-2, io.SeekCurrent) //io.SeekCurrent表示游标放置于当前位置，逆向偏移2个字节
				//读完一个字节后游标会自动正向偏移一个字节
				file.Read(char)
				if char[0] == '\r' {
					offset-- //windows跳过'\r'
				}
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

// 读取倒数第n行(n从1开始),
// 若n大于文件行数则返回错误io.EOF。
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
	//到此文件已经从尾部读到头部
	if lineCount == n-1 {
		return lineStr, nil
	}
	return "", io.EOF
}

// 给目录或文件创建快捷方式(filename可以为绝对路径也可以为相对路径,dir必须是绝对路径)
func CreateShortcut(filename, dir string) error {
	//获取文件的绝对路径
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	//获取文件的名称,(最后一个'\'后的内容)Base returns the last element of path
	name := filepath.Base(filename)
	//获取文件的扩展名
	ext := filepath.Ext(filename)
	//获取文件的名称(不包含扩展名)
	name = strings.TrimSuffix(name, ext)
	//拼接快捷方式的绝对路径
	shortcut := filepath.Join(dir, name+".lnk")
	//创建快捷方式
	return os.Symlink(absPath, shortcut)
}

// Compress compresses the file to the zip file.
func Compress(file []string, zipFile string) error {
	//创建一个新的zip文件
	fw, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer fw.Close()
	//创建一个新的zip writer
	zw := zip.NewWriter(fw)
	defer zw.Close()
	//遍历所有文件
	for _, f := range file {
		//打开文件
		fr, err := os.Open(f)
		if err != nil {
			return err
		}
		defer fr.Close()
		//创建一个zip文件信息头
		fw, err := zw.Create(f)
		if err != nil {

			return err
		}
		//将文件写入zip文件
		_, err = io.Copy(fw, fr)
		if err != nil {
			return err
		}
	}
	return nil
}

// 解压
func UnCompress(zipFile, dest string) error {
	//若目标文件夹不存在，则创建
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		os.MkdirAll(dest, os.ModePerm)
	}
	//打开zip文件
	fr, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer fr.Close()
	//遍历所有文件
	for _, f := range fr.File {
		//打开文件
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		//创建文件
		fw, err := os.Create(dest + "/" + f.Name)
		if err != nil {
			return err
		}
		defer fw.Close()
		//将文件写入磁盘
		_, err = io.Copy(fw, rc)
		if err != nil {
			return err
		}
	}
	return nil
}
