package logger

import (
	"fmt"
	"io"
	"log"
)

type Logger struct {
	logger *log.Logger
}

func (l *Logger) Init(name string, w io.Writer) {
	l.logger = log.New(w, fmt.Sprintf("[GDJ][%s]", name), log.LstdFlags)
}

func (l *Logger) Print(text string) {
	l.logger.Print(text)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

var debugLogger, infoLogger, warnLogger, errorLogger *log.Logger

func Init(dw, iw, ww, ew io.Writer) {
	debugLogger = log.New(dw, "[GDJ][DEBUG] ", log.LstdFlags)
	infoLogger = log.New(iw, "[GDJ][INFO] ", log.LstdFlags)
	warnLogger = log.New(ww, "[GDJ][WARN] ", log.LstdFlags)
	errorLogger = log.New(ew, "[GDJ][ERROR] ", log.LstdFlags)
}

func Debug(info string) {
	debugLogger.Println(info)
}

func Debugf(format string, v ...interface{}) {
	debugLogger.Printf(format, v...)
}

func Info(info string) {
	infoLogger.Println(info)
}

func Infof(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

func Warn(warn string) {
	warnLogger.Println(warn)
}

func Warnf(format string, v ...interface{}) {
	warnLogger.Printf(format, v...)
}

func Error(error string) {
	errorLogger.Println(error)
}

func Errorf(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}
