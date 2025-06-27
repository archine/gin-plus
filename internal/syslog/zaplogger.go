package syslog

import (
	"context"
	"fmt"
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/exception"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type conf struct {
	// Level syslog level, default info.
	// 	Supports: error、info、trace、warn、panic、fatal、debug
	Level string `json:"level" yaml:"level"`

	// LevelColor whether to enable color output for syslog levels, default true.
	// Note: this option only works when the formatter is console.
	EnableColor bool `json:"enable_color" yaml:"enable_color"`

	// Formatter syslog format, default console (supports: json、console)
	// 	json: output syslog in json format.
	Format string `json:"format" yaml:"format"`

	// ConsoleSeparator console separator.
	// Note: when the formatter is console, the separator between the fields, default is "\t".
	ConsoleSeparator string `json:"console_separator" yaml:"console_separator"`

	// CtxKeys When using WithContext for syslog output.
	// the value of the specified key is obtained from the context and added to the syslog.
	CtxKeys []string `json:"ctx_keys" yaml:"ctx_keys"`
}

// defaultLogger is the default implementation of the syslog interface.
type zaplog struct {
	z       *zap.Logger
	ctxKeys []string
}

func (d *zaplog) convertFields(fields []gplog.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

func (d *zaplog) Info(text string, fields ...gplog.Field) {
	d.z.Info(text, d.convertFields(fields)...)
}

func (d *zaplog) Debug(text string, fields ...gplog.Field) {
	d.z.Debug(text, d.convertFields(fields)...)
}

func (d *zaplog) Warn(text string, fields ...gplog.Field) {
	d.z.Warn(text, d.convertFields(fields)...)
}

func (d *zaplog) Error(text string, fields ...gplog.Field) {
	d.z.Error(text, d.convertFields(fields)...)
}

func (d *zaplog) Fatal(text string, fields ...gplog.Field) {
	d.z.Fatal(text, d.convertFields(fields)...)
}

func (d *zaplog) WithContext(ctx context.Context) gplog.Logger {
	if d.ctxKeys == nil {
		return d
	}

	var fields []zap.Field
	for _, key := range d.ctxKeys {
		if value := ctx.Value(key); value != nil {
			fields = append(fields, zap.Any(key, value))
		}
	}
	return &zaplog{
		z:       d.z.With(fields...),
		ctxKeys: d.ctxKeys,
	}
}

type SysLoggerInitListener struct{}

func (s *SysLoggerInitListener) OnConfigAfterLoad(configure config.Configure) {
	var cf conf
	if err := configure.Unmarshal("syslog", &cf); err != nil {
		panic(exception.NewStackErr("Init syslog config failed: " + err.Error()))
	}
	if cf.Level == "" {
		cf.Level = "info"
	}
	if cf.Format == "" {
		cf.Format = "console"
	}
	if cf.ConsoleSeparator == "" {
		cf.ConsoleSeparator = "\t"
	}

	zapLevel, err := zapcore.ParseLevel(cf.Level)
	if err != nil {
		panic(fmt.Sprintf("Invalid syslog level: %s, error: %v", cf.Level, err))
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
		ConsoleSeparator: cf.ConsoleSeparator,
	}

	if cf.EnableColor && cf.Format == "console" {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	var encoder zapcore.Encoder
	if cf.Format == "json" {
		encoder = zapcore.NewJSONEncoder(ec)
	} else {
		encoder = zapcore.NewConsoleEncoder(ec)
	}

	zapCore := zapcore.NewCore(encoder, os.Stderr, zapLevel)

	_ = &zaplog{
		z:       zap.New(zapCore),
		ctxKeys: cf.CtxKeys,
	}
}

func (s *SysLoggerInitListener) Order() int {
	return 0
}
