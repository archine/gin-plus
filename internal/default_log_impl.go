package internal

import (
	"context"
	"fmt"
	"github.com/archine/gin-plus/v3/module/logger"
	"log"
)

var Log logger.Logger = &defaultLogImplement{}

// defaultLogImplement provides the default implementation for the Logger interface.
// The ErrorWithCtx method only logs the message and does not utilize the context.
type defaultLogImplement struct{}

func (d *defaultLogImplement) Info(msg string, args ...any) {
	d.process(msg, args...)
}

func (d *defaultLogImplement) Warn(msg string, args ...any) {
	d.process(msg, args...)
}

func (d *defaultLogImplement) Debug(msg string, args ...any) {
	d.process(msg, args...)
}

func (d *defaultLogImplement) Error(msg string, args ...any) {
	d.process(msg, args...)
}

func (d *defaultLogImplement) ErrorWithCtx(ctx context.Context, msg string) {
	d.process(msg) // Context is not used here
}

// process formats and prints log messages with a timestamp and log level.
func (d *defaultLogImplement) process(msg string, args ...any) {
	if len(args) > 0 {
		log.Printf("%s %s", msg, fmt.Sprint(args...))
	} else {
		log.Println(msg)
	}
}
