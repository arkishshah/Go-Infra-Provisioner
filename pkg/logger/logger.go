package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func NewLogger() *Logger {
	flags := log.Ldate | log.Ltime | log.Lshortfile

	return &Logger{
		infoLog:  log.New(os.Stdout, "INFO: ", flags),
		errorLog: log.New(os.Stderr, "ERROR: ", flags),
	}
}

func (l *Logger) Info(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.infoLog.Output(2, fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), msg))
}

func (l *Logger) Error(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.errorLog.Output(2, fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), msg))
}

func (l *Logger) Fatal(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.errorLog.Output(2, fmt.Sprintf("[%s] FATAL: %s", time.Now().Format(time.RFC3339), msg))
	os.Exit(1)
}
