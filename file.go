package doraemon

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ListFilesRecursively recursively retrieves all files under the specified path, including files in subdirectories.
// If the path is a file, it returns the path itself.
// The returned file paths are either absolute or relative based on the input path.
func ListFilesRecursively(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	files := make([]string, 0)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// ListDirsRecursively recursively retrieves all directories under the specified path, including subdirectories.
// If the path is a directory, it returns the path itself.
// The returned directory paths are either absolute or relative based on the input path.
func ListDirsRecursively(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	dirs := make([]string, 0)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	return dirs, err
}

// ListAllRecursively recursively retrieves all paths under the specified path, including files and directories.
// If the path is a file, it returns the path itself.
// The returned paths are either absolute or relative based on the input path.
func ListAllRecursively(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	paths := make([]string, 0)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	return paths, err
}

// 获取path下所有文件名称(含后缀,不含路径)
func GetFileNmaesInPath(path string) ([]string, error) {
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

// 获取path路径下的文件夹名称(不含路径)
func GetFolderNamesInPath(path string) ([]string, error) {
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
func GetAllNamesInPath(path string) ([]string, error) {
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

// 获取path路径下的文件(含后缀)和文件夹名称，以及是否为文件夹
func GetAllNamesInPath2(path string) ([]Pair[string, bool], error) {
	DirEntry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var all []Pair[string, bool]
	for _, v := range DirEntry {
		all = append(all, Pair[string, bool]{v.Name(), v.IsDir()})
	}
	return all, nil
}

// FileOrDirIsExist Check if a file or directory exists
func FileOrDirIsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// FileIsExist Check if a file exists, return True, False, or Unknown
func FileIsExist(path string) Ternary {
	f, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return Unknown
		}
		return False
	}
	if f.IsDir() {
		return False
	}
	return True
}

// DirIsExist Check if a directory exists, return True, False, or Unknown
func DirIsExist(path string) Ternary {
	f, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return Unknown
		}
		return False
	}
	if f.IsDir() {
		return True
	}
	return False
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
	if path == "" {
		return false, false, fmt.Errorf("path is empty")
	}
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

// windows下读取某些非正常快捷方式文件时会报错 read xxx : Incorrect function.
// 这种快捷方式使用os.Stat()查询会报告为文件夹(IsDir()会返回true)，
// 但是使用os.ReadDir读取父文件夹来查询这个子快捷方式时，IsDir() 会返回false。

// When reading certain abnormal shortcut files on Windows, an error "read xxx: Incorrect function" may occur.
// These shortcuts are reported as folders when queried using os.Stat() (IsDir() returns true).
// However, when using os.ReadDir to read the parent folder and query this child shortcut, IsDir() returns false.
const WindowsReadLnkFileErrorKeyWords = "Incorrect function"

// 复制文件到指定目录
//
// overwrite为true时，如果目标文件存在则覆盖，
// overwrite为false时，如果目标文件存在则返回错误。
// dst必须是一个存在的文件夹，否则返回错误。
// scr为文件的绝对或相对路径(包含文件名)。
func CopyFile(src, dst string, overwrite bool) error {
	dst = strings.TrimSuffix(dst, `\`)
	dst = strings.TrimSuffix(dst, `/`)
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
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("copy file error: %v", err)
	}
	dstFile.Close()
	srcInfo, _ := srcFile.Stat()
	os.Chmod(target, srcInfo.Mode())
	os.Chtimes(target, time.Now(), srcInfo.ModTime())
	return nil
}

// 移动文件或文件夹到指定目录
//
// overwrite为true时，如果目标文件存在则覆盖(dst中的目标文件或文件夹会被直接删除)，
// overwrite为false时，如果目标文件存在则返回错误。
// dst必须是一个存在的文件夹，否则返回错误。
// scr为的绝对或相对路径。
func MoveFileOrDir(src, dst string, overwrite bool) error {
	dst = strings.TrimSuffix(dst, `\`)
	dst = strings.TrimSuffix(dst, `/`)
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

// 复制文件或文件夹
//
// overwrite为true时，如果目标文件存在则覆盖(dst中的目标文件或文件夹会被直接删除)，
// overwrite为false时，如果目标文件存在则返回错误。
// scr,dst 为绝对或相对路径,dst必须是一个文件夹(可以不存在)。
func CopyFileOrDir(src, dst string, overwrite bool) error {
	dstIsFile, _, _ := IsFile(dst)
	if dstIsFile {
		return fmt.Errorf("%s is not a folder", dst)
	}
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}
	srcIsDir, _, _ := IsDir(src)
	if srcIsDir {
		return CopyDir(src, dst, overwrite)
	}
	return CopyFile(src, dst, overwrite)
}

// 复制文件夹到指定目录
//
// overwrite为true时，如果目标文件夹存在名字相同的文件则覆盖，
// overwrite为false时，如果目标文件存在则返回错误。
// dst,scr都必须是一个存在的文件夹，否则返回错误。
func CopyDir(src, dst string, overwrite bool) error {
	dst = strings.TrimSuffix(dst, `\`)
	dst = strings.TrimSuffix(dst, `/`)
	src = strings.TrimSuffix(src, `\`)
	src = strings.TrimSuffix(src, `/`)
	// dst加上原文件夹名字
	dst = filepath.Join(dst, filepath.Base(src)) //dst更新为目标文件夹
	if !FileOrDirIsExist(dst) {
		err := os.MkdirAll(dst, 0666)
		if err != nil {
			return err
		}
	}
	// 排除dst，防止死循环(如果dst是src的子文件夹)
	if IsChildDir(src, dst) {
		return fmt.Errorf("\"%s\" is a child folder of \"%s\"", dst, src)
	}
	//获取src下所有文件
	srcFiles, err := GetFileNmaesInPath(src)
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
	srcDirNames, err := GetFolderNamesInPath(src)
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

// 判断child是否是parent的子文件夹(不存在的文件夹会返回false)
func IsChildDir(parent, child string) bool {
	// abs会统一路径分隔符为系统默认的分隔符
	parentAbs, err := filepath.Abs(parent)
	if err != nil {
		return false
	}
	childAbs, err := filepath.Abs(child)
	if err != nil {
		return false
	}
	return strings.HasPrefix(childAbs, parentAbs)
}

// 判断child是否是parent的子文件夹(为了性能只是简单的判断前缀，需要保证路径分隔符一致)
func IsChildDir2(parent, child string) bool {
	parent = strings.ToUpper(parent)
	child = strings.ToUpper(child)
	return strings.HasPrefix(child, parent)
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

// InitJsonConfig initializes a JSON configuration object of type T from a file.
// If the file doesn't exist, it creates a default file using the createDefault function (or a default create function if not provided).
// It uses JSON unmarshalling to parse the file content into the config object.
func InitJsonConfig[T any](configFilePath string, createDefault func(path string) error) (*T, error) {
	var config T

	if createDefault == nil {
		createDefault = func(path string) error {
			c, err := json.MarshalIndent(DeepCreateEmptyInstance(reflect.TypeOf(config)), "", "    ")
			if err != nil {
				return err
			}
			return os.WriteFile(path, c, 0666)
		}
	}
	return InitConfig[T](configFilePath, createDefault, json.Unmarshal)
}

const DevLocalConfigFileKeyWord = ".local.dev"

// InitConfig initializes a configuration object of type T from a specified file.
// If a development-specific configuration file exists (with a ".local.dev" suffix),
// it will be loaded instead. If the specified configuration file does not exist,
// it will be created using the provided createDefault function.
// The unmarshal function is used to parse the configuration file content into the object.
func InitConfig[T any](
	configFilePath string,
	createDefault func(path string) error,
	unmarshal func(data []byte, v any) error,
) (*T, error) {
	devConfigFilePath := generateDevConfigPath(configFilePath)
	if FileIsExist(devConfigFilePath).IsTrue() {
		return loadConfigFromFile[T](devConfigFilePath, unmarshal)
	}
	if FileIsExist(configFilePath).IsFalse() {
		err := createDefault(configFilePath)
		if err != nil {
			return nil, err
		}
	}
	return loadConfigFromFile[T](configFilePath, unmarshal)
}

func generateDevConfigPath(configFile string) string {
	var baseDir = filepath.Dir(configFile)
	var fileName = filepath.Base(configFile)
	var ext = filepath.Ext(fileName)
	var devConfigName = strings.TrimSuffix(fileName, ext) + DevLocalConfigFileKeyWord + ext
	var devConfigFilePath = filepath.Join(baseDir, devConfigName)
	return devConfigFilePath
}

func loadConfigFromFile[T any](
	configFile string,
	unmarshal func(data []byte, v any) error,
) (*T, error) {
	var config T
	configFileContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = unmarshal(configFileContent, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type LazyFileWriter struct {
	// Path to the file
	filePath string
	// File handle
	file *os.File
	// Ensures file is opened only once
	once *sync.Once
}

// Ensure LazyFileWriter implements io.WriteCloser
var _ io.WriteCloser = (*LazyFileWriter)(nil)

// NewLazyFileWriter creates a new LazyFileWriter with the given file path
func NewLazyFileWriter(filePath string) *LazyFileWriter {
	return &LazyFileWriter{filePath: filePath, once: &sync.Once{}}
}

// Write writes the given bytes to the file, creating the file if it doesn't exist.
func (w *LazyFileWriter) Write(p []byte) (n int, err error) {
	w.once.Do(func() {
		w.file, err = os.OpenFile(w.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	})
	if err != nil {
		return 0, err
	}
	f := w.file
	if f != nil {
		return f.Write(p)
	}
	return 0, os.ErrClosed
}

// Close closes the file. Close will return an error if it has already been called.
func (w *LazyFileWriter) Close() error {
	f := w.file
	if f != nil {
		w.file = nil
		return f.Close()
	}
	return os.ErrClosed
}

// Sync flushes file's in-memory state to disk
func (w *LazyFileWriter) Sync() error {
	f := w.file
	if f != nil {
		return f.Sync()
	}
	return os.ErrClosed
}

// Name returns the base name of the file
func (w *LazyFileWriter) Name() string {
	return filepath.Base(w.filePath)
}

func (w *LazyFileWriter) Path() string {
	return w.filePath
}

// IsCreated checks if the file has been created/opened
func (w *LazyFileWriter) IsCreated() bool {
	return w.file != nil
}

// File returns the file handle
func (w *LazyFileWriter) File() *os.File {
	return w.file
}

// GenerateUniqueFilepath generates a unique filepath by appending a number to the original filepath
// if the original filepath already exists.
func GenerateUniqueFilepath(filePath string) string {
	if !FileOrDirIsExist(filePath) {
		return filePath
	}
	dir := filepath.Dir(filePath)
	name := filepath.Base(filePath)
	fileExt := filepath.Ext(name)
	name = name[:len(name)-len(fileExt)]
	for i := 1; ; i++ {
		if fileExt != "" {
			filePath = filepath.Join(dir, fmt.Sprintf("%s(%d)%s", name, i, fileExt))
		} else {
			filePath = filepath.Join(dir, fmt.Sprintf("%s(%d)", name, i))
		}
		if !FileOrDirIsExist(filePath) {
			return filePath
		}
	}
}

func WriteFile(filePath string, perm fs.FileMode, datas ...[]byte) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	for _, data := range datas {
		_, err = f.Write(data)
		if err != nil {
			return err
		}
	}
	return f.Close()
}

func WriteFile2(filePath string, data io.Reader, perm fs.FileMode) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, data)
	if err != nil {
		return err
	}
	return f.Close()
}

// WriteFilePreservePerms writes data to a file named name, preserving existing permissions if the file exists.
// If the file does not exist, it is created with permissions 0644 (rw-r--r--).
func WriteFilePreservePerms(name string, data []byte) error {
	perm := fs.FileMode(0644)
	f, err := os.Stat(name)
	if err == nil {
		perm = f.Mode()
	}
	return os.WriteFile(name, data, perm)
}
