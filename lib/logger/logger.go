package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Settings 存储 logger 的配置信息
type Settings struct {
	Path       string `yaml:"path"`
	Name       string `yaml:"name"`
	Ext        string `yaml:"ext"`
	TimeFormat string `yaml:"time-format"`
}

var (
	logFile            *os.File
	defaultPrefix      = ""
	defaultCallerDepth = 2
	logger             *log.Logger
	mu                 sync.Mutex
	logPrefix          = ""
	levelFlags         = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

type logLevel int

// 日志等级
const (
	DEBUG logLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

const flags = log.LstdFlags

func init() {
	logger = log.New(os.Stdout, defaultPrefix, flags)
}

// Setup 初始化 logger
func Setup(settings *Settings) {
	var err error
	dir := settings.Path
	fileName := fmt.Sprintf("%s-%s.%s",
		settings.Name,
		time.Now().Format(settings.TimeFormat),
		settings.Ext)

	logFile, err := mustOpen(fileName, dir)
	if err != nil {
		log.Fatalf("logging.Setup err: %s", err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(mw, defaultPrefix, flags)
}

func setPrefix(level logLevel) {
	_, file, line, ok := runtime.Caller(defaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d] ", levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s] ", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}

func Debug(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(DEBUG)
	logger.Println(v...)
}

func Info(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(INFO)
	logger.Println(v...)
}

func Warn(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(WARNING)
	logger.Println(v...)
}

func Error(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(ERROR)
	logger.Println(v...)
}

func Fatal(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(FATAL)
	logger.Fatalln(v...)
}
