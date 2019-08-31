package logger

import (
	"io"
	"log"
)

var infoLogger, warnLogger, errorLogger *log.Logger

func Init(iw, ww, ew io.Writer) {
	infoLogger = log.New(iw, "[GDJ][INFO] ", log.LstdFlags)
	warnLogger = log.New(ww, "[GDJ][WARN] ", log.LstdFlags)
	errorLogger = log.New(ew, "[GDJ][ERROR] ", log.LstdFlags)
}

func Info(info string) {
	infoLogger.Println(info)
}

func Infof(format string, v ...interface{}) {
	infoLogger.Printf(format, v)
}

func Warn(warn string) {
	warnLogger.Println(warn)
}

func Warnf(format string, v ...interface{}) {
	warnLogger.Printf(format, v)
}

func Error(error string) {
	errorLogger.Println(error)
}

func Errorf(format string, v ...interface{}) {
	errorLogger.Printf(format, v)
}
