package logger

import (
	"github.com/gin-gonic/gin"
	"log"
)

// DefaultLog use golang log as default
type DefaultLog struct{}

func (d *DefaultLog) GetLogger() any {
	return log.Default()
}

func (d *DefaultLog) Init() {
	// do nothing
}

func (d *DefaultLog) Infof(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Warnf(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Debugf(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Errorf(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (d *DefaultLog) Info(v ...any) {
	log.Println(v)
}

func (d *DefaultLog) Warn(v ...any) {
	log.Println(v)
}

func (d *DefaultLog) Debug(v ...any) {
	log.Println(v)
}

func (d *DefaultLog) Error(v ...any) {
	log.Println(v)
}

func (d *DefaultLog) Println(v ...any) {
	log.Println(v...)
}

func (d *DefaultLog) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

func (d *DefaultLog) Fatal(v ...any) {
	log.Fatal(v...)
}

func (d *DefaultLog) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}

func (d *DefaultLog) GinMiddleware() gin.HandlerFunc {
	return nil
}
