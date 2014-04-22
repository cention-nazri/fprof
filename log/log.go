package log

import (
	"io"
	"log"
)

var logger *log.Logger

func Init(w io.Writer, prefix string) {
	logger = log.New(w, prefix, log.LstdFlags)
}

func Print(v interface{}) {
	logger.Print(v)
}

func Printf(fmt string, v ...interface{}) {
	logger.Printf(fmt, v...)
}

func Println(v ...interface{}) {
	logger.Println(v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}
