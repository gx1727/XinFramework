package logger

import (
	"log"
	"os"
)

var Logger *log.Logger

func Init() {
	Logger = log.New(os.Stdout, "[xin] ", log.LstdFlags)
}

func Info(v ...interface{}) {
	Logger.Println(v...)
}

func Error(v ...interface{}) {
	Logger.Println(v...)
}
