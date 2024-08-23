package internal

import (
	"context"
	"github.com/archine/gin-plus/v3/module/logger"
	"log"
)

var Log logger.Logger = &defaultLogImplement{}

// defaultLogImplement provides the default implementation for the Logger interface.
// The ErrorWithCtx method only logs the message and does not utilize the context.
type defaultLogImplement struct{}

func (d *defaultLogImplement) Info(msg string) {
	log.Println(msg)
}

func (d *defaultLogImplement) Warn(msg string) {
	log.Println(msg)
}

func (d *defaultLogImplement) Debug(msg string) {
	log.Println(msg)
}

func (d *defaultLogImplement) Error(msg string) {
	log.Println(msg)
}

func (d *defaultLogImplement) ErrorWithCtx(ctx context.Context, msg string) {
	log.Println(msg) // Context is not used here
}
