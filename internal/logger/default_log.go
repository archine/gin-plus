package logger

import (
	"log"
)

// DefaultLog use golang log as default
type DefaultLog struct{}

func (d *DefaultLog) Info(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Warn(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Debug(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Error(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Fatal(format string, v ...any) {
	log.Fatalf(format, v...)
}
