// Package logger provides a super simple way for logging on different levels
package logger

import (
	"log"
)

// LogLevel is a wrapping type around int to set the loglevel of the package
type LogLevel int

const (
	// ErrorLevel log level
	ErrorLevel LogLevel = iota
	// InfoLevel log level
	InfoLevel
	// DebugLevel log level
	DebugLevel
)

var level = ErrorLevel

// SetLogLevel allows consumers to set the logging level
func SetLogLevel(lvl LogLevel) {
	level = lvl
}

// Error writes errors
func Error(msg string) {
	if level >= ErrorLevel {
		log.Println(msg)
	}
}

// Info writes info logs
func Info(msg string) {
	if level >= InfoLevel {
		log.Println(msg)
	}
}

// Debug writes info logs
func Debug(msg string) {
	if level >= DebugLevel {
		log.Println(msg)
	}
}
