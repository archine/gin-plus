package logger

import (
	"context"
	"fmt"
	"github.com/archine/gin-plus/v3/module/gplog/iface"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var GlobalLogger iface.AbstractAppLogger // Log is the global logger instance.

// config is the configuration for the logger.
type config struct {
	// Level log level, default info.
	// 	Supports: error、info、trace、warn、panic、fatal、debug
	Level string `json:"level" yaml:"level"`

	// LevelColor whether to enable color output for log levels, default true.
	// Note: this option only works when the formatter is console.
	EnableColor bool `json:"enable_color" yaml:"enable_color"`

	// Formatter log format, default console (supports: json、console)
	// 	json: output log in json format.
	Format string `json:"format" yaml:"format"`

	// ConsoleSeparator console separator.
	// Note: when the formatter is console, the separator between the fields, default is "\t".
	ConsoleSeparator string `json:"console_separator" yaml:"console_separator"`

	// CtxKeys When using WithContext for log output.
	// the value of the specified key is obtained from the context and added to the log.
	CtxKeys []string `json:"ctx_keys" yaml:"ctx_keys"`
}

// defaultLogger is the default implementation of the logger interface.
type defaultLogger struct {
	z       *zap.Logger
	ctxKeys []string
}

func (d *defaultLogger) Info(text string) {
	d.z.Info(text)
}

func (d *defaultLogger) Debug(text string) {
	d.z.Debug(text)
}

func (d *defaultLogger) Warn(text string) {
	d.z.Warn(text)
}

func (d *defaultLogger) Error(text string) {
	d.z.Error(text)
}

func (d *defaultLogger) Fatal(text string) {
	d.z.Fatal(text)
}

func (d *defaultLogger) WithContext(ctx context.Context) iface.AbstractAppLogger {
	nd := &defaultLogger{}

	var fields []zap.Field
	for _, key := range d.ctxKeys {
		if value := ctx.Value(key); value != nil {
			fields = append(fields, zap.Any(key, value))
		}
	}
	nd.z = nd.z.With(fields...)

	return nd
}

// DefaultLoggerInitListener is the default logger initialization listener.
type DefaultLoggerInitListener struct {
}

func (d *DefaultLoggerInitListener) OnConfigAfterLoad(v *viper.Viper) {
	if GlobalLogger != nil {
		return
	}

	v.SetDefault("log.level", "info")
	v.SetDefault("log.enable_color", true)
	v.SetDefault("log.format", "console")
	v.SetDefault("log.console_separator", "\t")
	v.SetDefault("log.ctx_keys", []string{})

	var conf config
	if err := v.UnmarshalKey("log", &conf); err != nil {
		panic("Init logger config failed: " + err.Error())
	}

	zapLevel, err := zapcore.ParseLevel(conf.Level)
	if err != nil {
		panic(fmt.Sprintf("Invalid log level: %s, error: %v", conf.Level, err))
	}

	ec := zapcore.EncoderConfig{
		TimeKey:          "timestamp",
		LevelKey:         "level",
		CallerKey:        "caller",
		MessageKey:       "msg",
		FunctionKey:      zapcore.OmitKey,
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		LineEnding:       zapcore.DefaultLineEnding,
		StacktraceKey:    "stacktrace",
		ConsoleSeparator: conf.ConsoleSeparator,
	}

	if conf.EnableColor && conf.Format == "console" {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	var encoder zapcore.Encoder
	if conf.Format == "json" {
		encoder = zapcore.NewJSONEncoder(ec)
	} else {
		encoder = zapcore.NewConsoleEncoder(ec)
	}

	zapCore := zapcore.NewCore(encoder, os.Stderr, zapLevel)

	GlobalLogger = &defaultLogger{
		z:       zap.New(zapCore),
		ctxKeys: conf.CtxKeys,
	}
}

func (d *DefaultLoggerInitListener) Order() int {
	return 99
}
