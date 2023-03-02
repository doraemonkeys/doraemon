package doraemon

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// 日志配置,可以为空
type LogConfig struct {
	//日志路径(可以为空)
	LogPath string
	//日志文件名后缀(不包含日期和.log，可以为空)
	LogFileNameSuffix string
	//是否分离错误日志(Error级别以上)
	ErrSeparate bool
	//如果分离错误日志，普通日志文件是否仍然包含错误日志
	ErrInNormal bool
	//按日期分割日志(不能和按大小分割同时使用)
	DateSplit bool
	//取消日志输出到文件
	NoFile bool
	//取消日志输出到控制台
	NoConsole bool
	//取消时间戳Timestamp
	NoTimestamp bool
	//在控制台输出shortfile
	ShowShortFileInConsole bool
	//在控制台输出func
	ShowFuncInConsole bool
	//按大小分割日志,单位byte。(不能和按日期分割同时使用)
	MaxLogSize int64
	//日志扩展名(默认.log)
	LogExt string
	//panic,fatal,error,warn,info,debug,trace
	LogLevel string
	//时区
	TimeLocation *time.Location
	//在每条log末尾添加key-value
	key string
	//在每条log末尾添加key-value
	value interface{}
}

// 在每条log末尾添加key-value
func (c *LogConfig) SetKeyValue(key string, value interface{}) {
	c.key = key
	c.value = value
}

type logHook struct {
	ErrWriter   *os.File
	OtherWriter *os.File
	//修改Writer时加锁
	WriterLock *sync.RWMutex
	LogConfig  LogConfig
	// 2006_01_02
	FileDate string
	// byte,仅在SizeSplit>0时有效
	LogSize int64
	// 2006_01_02
	dateFmt string
}

func (hook *logHook) Fire(entry *logrus.Entry) error {
	if hook.LogConfig.key != "" {
		entry.Data[hook.LogConfig.key] = hook.LogConfig.value
	}
	file := entry.Caller.File
	if strings.Contains(file, "/") {
		file = file[strings.LastIndex(file, "/")+1:] + ":" + fmt.Sprint(entry.Caller.Line)
	}
	entry.Data["FILE"] = file
	entry.Data["FUNC"] = entry.Caller.Function[strings.LastIndex(entry.Caller.Function, ".")+1:]

	if !hook.LogConfig.ShowShortFileInConsole {
		defer delete(entry.Data, "FILE")
	}
	if !hook.LogConfig.ShowFuncInConsole {
		defer delete(entry.Data, "FUNC")
	}

	//取消日志输出到文件
	if hook.LogConfig.NoFile {
		return nil
	}

	//msg前添加固定前缀 DORAEMON
	//entry.Message = "DORAEMON " + entry.Message

	line, err := entry.Bytes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}
	line = eliminateColor(line)

	hook.checkSplit()

	hook.WriterLock.RLock()
	defer hook.WriterLock.RUnlock()
	if hook.ErrWriter != nil && entry.Level <= logrus.ErrorLevel {
		hook.LogSize += int64(len(line))
		hook.ErrWriter.Write(line)

		if !hook.LogConfig.ErrInNormal {
			return nil
		}
	}

	if hook.OtherWriter != nil {
		hook.LogSize += int64(len(line))
		hook.OtherWriter.Write(line)
	}

	return nil
}

func (hook *logHook) Levels() []logrus.Level {
	//return []logrus.Level{logrus.ErrorLevel}

	//hook全部级别
	return logrus.AllLevels
}

// 去除颜色
func eliminateColor(line []byte) []byte {
	//"\033[31m 红色 \033[0m"
	if bytes.Contains(line, []byte("\x1b[0m")) {
		line = bytes.ReplaceAll(line, []byte("\x1b[0m"), []byte(""))

		index := bytes.Index(line, []byte("\x1b[")) //找到\x1b[的位置
		for index >= 0 && index+5 < len(line) {
			line = bytes.ReplaceAll(line, line[index:index+5], []byte("")) //删除\x1b[31m
			index = bytes.Index(line, []byte("\x1b["))
		}
	}
	return line
}

// 检查是否需要分割日志
func (hook *logHook) checkSplit() {
	if hook.LogConfig.DateSplit {
		//按日期分割
		now := time.Now().In(hook.LogConfig.TimeLocation).Format(hook.dateFmt)
		if hook.FileDate != now {
			hook.WriterLock.Lock()
			if hook.FileDate == now {
				//已经分割过了
				hook.WriterLock.Unlock()
				return
			}
			hook.FileDate = now
			hook.split_date()
			hook.WriterLock.Unlock()
		}
		return
	}

	if hook.LogConfig.MaxLogSize > 0 {
		//按大小分割
		if hook.LogSize >= hook.LogConfig.MaxLogSize {
			//fmt.Println("日志大小超过限制，开始分割日志", hook.LogSize, hook.LogConfig.MaxLogSize)
			hook.WriterLock.Lock()
			if hook.LogSize < hook.LogConfig.MaxLogSize {
				//已经分割过了
				hook.WriterLock.Unlock()
				return
			}
			hook.LogSize = 0
			hook.split_size()
			hook.WriterLock.Unlock()
		}
		return
	}
}

// 按大小分割日志
func (hook *logHook) split_size() {
	if hook.ErrWriter != nil {
		hook.ErrWriter.Close()
	}
	if hook.OtherWriter != nil {
		hook.OtherWriter.Close()
	}
	err := hook.updateNewLogPathAndFile()
	if err != nil {
		panic(fmt.Sprintf("分割日志失败: %v", err))
	}
}

// 按日期分割日志
func (hook *logHook) split_date() {
	if hook.ErrWriter != nil {
		hook.ErrWriter.Close()
	}
	if hook.OtherWriter != nil {
		hook.OtherWriter.Close()
	}
	err := hook.updateNewLogPathAndFile()
	if err != nil {
		panic(fmt.Sprintf("分割日志失败: %v", err))
	}
}

func (hook *logHook) updateNewLogPathAndFile() error {
	if hook.LogConfig.NoFile {
		return nil
	}

	// 检查日志目录是否存在
	if hook.LogConfig.LogPath != "" {
		if _, err := os.Stat(hook.LogConfig.LogPath); os.IsNotExist(err) {
			err = os.MkdirAll(hook.LogConfig.LogPath, 0755)
			if err != nil {
				return err
			}
		}
	}

	//更新日期(不多余，split_size也会用到)
	hook.FileDate = time.Now().In(hook.LogConfig.TimeLocation).Format(hook.dateFmt)
	//按大小分割时，文件名可能会重复
	var tempFileDate string = hook.FileDate
	if hook.LogConfig.MaxLogSize > 0 {
		//纳秒后4位
		tempFileDate = fmt.Sprintf("%s_%d", hook.FileDate, (time.Now().UnixNano()%1000000)/100)
	}

	var newFileName string
	if !hook.LogConfig.ErrSeparate {
		if hook.LogConfig.LogFileNameSuffix == "" {
			newFileName = tempFileDate + hook.LogConfig.LogExt
		} else {
			newFileName = tempFileDate + "_" + hook.LogConfig.LogFileNameSuffix + hook.LogConfig.LogExt
		}
		newFileName = makeFileNameLegal(newFileName)
		newFileName = filepath.Join(hook.LogConfig.LogPath, newFileName)

		file, err := os.OpenFile(newFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		hook.OtherWriter = file
		//更新日志大小
		hook.LogSize, _ = file.Seek(0, io.SeekEnd)
		return nil
	}

	var commonFileName string
	if hook.LogConfig.LogFileNameSuffix == "" {
		newFileName = tempFileDate + "_" + "ERROR" + hook.LogConfig.LogExt
		commonFileName = tempFileDate + hook.LogConfig.LogExt
	} else {
		newFileName = tempFileDate + "_" + "ERROR" + "_" + hook.LogConfig.LogFileNameSuffix + hook.LogConfig.LogExt
		commonFileName = tempFileDate + "_" + hook.LogConfig.LogFileNameSuffix + hook.LogConfig.LogExt
	}
	newFileName = makeFileNameLegal(newFileName)
	commonFileName = makeFileNameLegal(commonFileName)

	newPath := filepath.Join(hook.LogConfig.LogPath, hook.FileDate)
	newFileName = filepath.Join(newPath, newFileName)
	commonFileName = filepath.Join(newPath, commonFileName)
	err := os.MkdirAll(newPath, 0777)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(newFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	hook.ErrWriter = file
	hook.LogSize, _ = file.Seek(0, io.SeekEnd)

	file2, err := os.OpenFile(commonFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	hook.OtherWriter = file2
	tempSize, _ := file2.Seek(0, io.SeekEnd)
	hook.LogSize += tempSize

	return nil
}

func InitGlobalLog(config LogConfig) error {
	return initlLog(logrus.StandardLogger(), config)
}

func NewLog(config LogConfig) (*logrus.Logger, error) {
	logger := logrus.New()
	err := initlLog(logger, config)
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func initlLog(logger *logrus.Logger, config LogConfig) error {

	var levelStr = flag.String("level", "", "log level(panic,fatal,error,warn,info,debug,trace)")
	flag.Parse()
	if *levelStr == "" {
		*levelStr = config.LogLevel
	}

	var level logrus.Level = praseLevel(*levelStr)
	//fmt.Println("level:", level)

	logger.SetReportCaller(true) //开启调用者信息
	logger.SetLevel(level)       //设置最低的Level
	formatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05", //时间格式
		FullTimestamp:   true,                  //开启时间戳
		ForceColors:     true,                  //开启颜色
		// CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		// 	//返回shortfile,funcname,linenum
		// 	//main.go:main:12
		// 	shortFile := f.File
		// 	if strings.Contains(f.File, "/") {
		// 		shortFile = f.File[strings.LastIndex(f.File, "/")+1:]
		// 	}
		// 	return "", fmt.Sprintf("%s:%s():%d:", shortFile, f.Function, f.Line)
		// },
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", ""
		},
	}

	if config.NoTimestamp {
		formatter.DisableTimestamp = true
	}
	logrus.SetFormatter(formatter)

	if config.NoConsole {
		logrus.SetOutput(io.Discard)
	}

	if config.LogExt == "" {
		config.LogExt = ".log"
	}
	if config.LogExt[0] != '.' {
		config.LogExt = "." + config.LogExt
	}
	if config.TimeLocation == nil {
		config.TimeLocation = time.Local
	}

	hook := &logHook{}
	hook.dateFmt = "2006_01_02"
	hook.FileDate = time.Now().In(config.TimeLocation).Format(hook.dateFmt)
	hook.LogSize = 0
	hook.WriterLock = &sync.RWMutex{}
	hook.LogConfig = config

	//添加hook
	logger.AddHook(hook)

	err := hook.updateNewLogPathAndFile()
	if err != nil {
		return fmt.Errorf("updateNewLogPathAndFile err:%v", err)
	}
	return nil
}

// panic,fatal,error,warn,info,debug,trace
func praseLevel(level string) logrus.Level {
	level = strings.ToLower(level)
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// 替换文件名中的非法字符为下划线
func makeFileNameLegal(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	return s
}
