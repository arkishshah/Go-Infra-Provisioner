package logger

import (
	"log"
	"os"
)

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLog:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.infoLog.Println(v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.errorLog.Println(v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.errorLog.Fatal(v...)
}
