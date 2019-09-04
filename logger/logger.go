package logger

import (
	"fmt"
	"io"
	"log"
	"time"
)

type Logger struct {
	logger *log.Logger
	Prefix string
	Icon   string
}

func NewLogger(out io.Writer, name string, icon string, flag int) *Logger {
	return &Logger{
		logger: log.New(out, "", flag),
		Prefix: fmt.Sprintf("[GDJ][%s]", name),
		Icon:   icon,
	}
}

func (l *Logger) SetPrefix() {
	prefix := fmt.Sprintf("%s[%s] ", l.Prefix, time.Now().Format("15:04:05"))
	if l.Icon != "" {
		prefix += l.Icon + " "
	}
	l.logger.SetPrefix(prefix)
}

func (l *Logger) Print(text string) {
	l.SetPrefix()
	l.logger.Print(text)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.SetPrefix()
	l.logger.Printf(format, v...)
}

func (l *Logger) Println(v ...interface{}) {
	l.logger.Println(v...)
}

var debugLogger, infoLogger, warnLogger, errorLogger *Logger

func Init(dw, iw, ww, ew io.Writer) {
	debugLogger = NewLogger(dw, "DEBUG", "", 0)
	infoLogger = NewLogger(iw, "INFO", "", 0)
	warnLogger = NewLogger(ww, "WARN", "⚠️", 0)
	errorLogger = NewLogger(ew, "ERROR", "☒", log.Lshortfile)
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
