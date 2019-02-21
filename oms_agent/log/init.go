package log

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	EnableDebug = false
	LogPath     = "/tmp"
	outFilePath = filepath.Join(LogPath, "go_agent.log")
	errFilePath = filepath.Join(LogPath, "go_agent_err.log")
	outFile, _  = os.OpenFile(outFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	errFile, _ = os.OpenFile(errFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//log.Fatalln("Failed to open error log file:", err)
	//}
	Info = log.New(io.MultiWriter(outFile, os.Stdout),
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(io.MultiWriter(outFile, os.Stdout),
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(io.MultiWriter(outFile, os.Stdout),
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(errFile, os.Stderr),
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
)

//func Info(msg ...interface{}) {
//	info.Println(msg...)
//}
//
//func Infof(s string, msg ...interface{}) {
//	info.Printf(s, msg...)
//}
//
//func Warning(msg ...interface{}) {
//	warning.Println(msg...)
//}
//
//func Debug(msg ...interface{}) {
//	if EnableDebug {
//		debug.Println(msg)
//	}
//}
//
//func Debugf(s string, msg ...interface{}) {
//	if EnableDebug {
//		debug.Printf(s, msg)
//	}
//}
//
//func Error(msg ...interface{}) {
//	error.Println(msg...)
//}
//
//func Errorf(s string, msg ...interface{}) {
//	error.Printf(s,msg...)
//}
