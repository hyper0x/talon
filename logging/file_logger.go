package logging

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type FileLogger struct {
	sync.RWMutex
	position  Position
	logDir    string      //日志存放目录
	logName   string      //日志名称
	timestamp time.Time   //日志创建时的时间戳
	logFile   *os.File    //当前日志文件实例
	maxSize   int64       //单个文件上限
	pipe      chan string //日志传输通道
}

// 开始从管道读取日志数据
func (f *FileLogger) startPipe() {

	var content string

	for {
		content = <-f.pipe
		f.logFile.WriteString(content + "\r\n")

	}
}

//初始化FileLogger
//fileDir 日志文件目录
//logName 日志名称
//cacheSize 缓冲区大小
//maxSize 单个文件大小上限 1M = 1 * 1024 * 1024
func NewFileLogger(fileDir string, logName string, cacheSize int, maxSize int64) Logger {

	f := &FileLogger{}
	//目录修正
	dir := fileDir
	if fileDir[len(fileDir)-1] == '\\' || fileDir[len(fileDir)-1] == '/' {
		dir = fileDir[:len(fileDir)-1]
	}

	f.logDir = dir           //日志目录
	f.logName = logName      //日志名称
	f.timestamp = time.Now() //初始时间戳

	//512 * 1024 * 1024
	f.maxSize = maxSize * 1024 * 1024
	f.pipe = make(chan string, cacheSize)
	return f

}

// 启动logger
func (f *FileLogger) Initialize() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("err", err)
		}
	}()

	f.newlogfile() //创建初始文件

	go f.fileMonitor() //监控文件变化
	go f.startPipe()   //开始读取日志
}

//开始监控日志文件
func (f *FileLogger) fileMonitor() {

	timer := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-timer.C:
			f.checkFile()
		}
	}
}

//检查文件

func (f *FileLogger) checkFile() {

	defer func() {
		if err := recover(); err != nil {
			log.Println("err", err)
		}
	}()

	if !isFileExist(f.GetCurrentLogPath()) { //如果文件不存在,创建日志文件
		f.newlogfile()
		return
	}

	flag := false //标志位,是否需要创建新文件

	//检查是否跨天
	if f.checkFileDate() {
		flag = true
	}

	//检查文件大小
	if f.checkFileSize() {
		flag = true
	}

	if flag { //创建新文件
		f.Lock()
		defer f.Unlock()

		f.timestamp = time.Now() //更新时间戳
		f.logFile.Close()
		f.newlogfile()
	}

}

//构建日志目录及日志文件
//返回日志文件路径
func (f *FileLogger) newlogfile() {

	//按天创建目录
	dir := fmt.Sprintf("%s/%04d-%02d-%02d/", f.getLogPath(), f.timestamp.Year(), f.timestamp.Month(), f.timestamp.Day())
	os.MkdirAll(dir, os.ModePerm)
	//按时间创建文件
	filename := fmt.Sprintf("%02d_%02d_%02d", f.timestamp.Hour(), f.timestamp.Minute(), f.timestamp.Second())

	fn := filename + ".log"

	fileFullPath := path.Join(dir, fn)

	file, err := os.OpenFile(fileFullPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)

	if err != nil {
		panic(err)
	}

	f.logFile = file

}

// 文件是否存在
func isFileExist(file string) bool {
	finfo, err := os.Stat(file)
	if err != nil {
		return false
	}
	if finfo.IsDir() {
		return false
	} else {
		return true
	}
}

//获取当前日志目录 logDir/logName
func (f *FileLogger) getLogPath() string {
	return path.Join(f.logDir, f.logName)
}

//获取当前日志文件全路径

func (f *FileLogger) GetCurrentLogPath() string {
	path, _ := filepath.Abs(f.logFile.Name())
	return path
}

//检查当前日志文件大小是够超限
func (f *FileLogger) checkFileSize() bool {

	fileInfo, err := f.logFile.Stat()
	if err != nil {
		return false
	}

	if fileInfo.Size() >= f.maxSize {
		return true
	}

	return false
}

// 检查日期是否已经跨天
func (f *FileLogger) checkFileDate() bool {
	if time.Now().YearDay() != f.timestamp.YearDay() {
		return true
	}

	return false
}

func (logger *FileLogger) GetPosition() Position {
	return logger.position
}

func (logger *FileLogger) SetPosition(pos Position) {
	logger.position = pos
}

func (logger *FileLogger) Error(v ...interface{}) string {
	content := generateLogContent(getErrorLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content

	return content
}

func (logger *FileLogger) Errorf(format string, v ...interface{}) string {
	content := generateLogContent(getErrorLogTag(), logger.GetPosition(), format, v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Errorln(v ...interface{}) string {
	content := generateLogContent(getErrorLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Fatal(v ...interface{}) string {
	content := generateLogContent(getFatalLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Fatalf(format string, v ...interface{}) string {
	content := generateLogContent(getFatalLogTag(), logger.GetPosition(), format, v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Fatalln(v ...interface{}) string {
	content := generateLogContent(getFatalLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Info(v ...interface{}) string {
	content := generateLogContent(getInfoLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Infof(format string, v ...interface{}) string {
	content := generateLogContent(getInfoLogTag(), logger.GetPosition(), format, v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Infoln(v ...interface{}) string {
	content := generateLogContent(getInfoLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Panic(v ...interface{}) string {
	content := generateLogContent(getPanicLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Panicf(format string, v ...interface{}) string {
	content := generateLogContent(getPanicLogTag(), logger.GetPosition(), format, v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Panicln(v ...interface{}) string {
	content := generateLogContent(getPanicLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Warn(v ...interface{}) string {
	content := generateLogContent(getWarnLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Warnf(format string, v ...interface{}) string {
	content := generateLogContent(getWarnLogTag(), logger.GetPosition(), format, v...)
	logger.pipe <- content
	return content
}

func (logger *FileLogger) Warnln(v ...interface{}) string {
	content := generateLogContent(getWarnLogTag(), logger.GetPosition(), "", v...)
	logger.pipe <- content
	return content
}
