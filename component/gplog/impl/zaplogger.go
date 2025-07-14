package impl

import (
	"context"
	"fmt"
	"github.com/archine/gin-plus/v4/component/gplog"
	"os"
	"strings"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/exception"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type conf struct {
	// Level syslog level, default info.
	// 	Supports: error、info、trace、warn、panic、fatal、debug
	Level string `yaml:"level"`

	// LevelColor whether to enable color output for syslog levels, default true.
	// Note: this option only works when the formatter is console.
	EnableColor bool `yaml:"enable-color"`

	// Formatter syslog format, default console (supports: json、console)
	// 	json: output syslog in json format.
	Format string `json:"format" yaml:"format"`

	// ConsoleSeparator console separator.
	// Note: when the formatter is console, the separator between the fields, default is "\t".
	ConsoleSeparator string `yaml:"console-separator"`

	// CtxKeys When using WithContext for syslog output.
	// the value of the specified key is obtained from the gpctx and added to the syslog.
	CtxKeys []string `yaml:"ctx-keys"`
}

// defaultLogger is the default implementation of the syslog interface.
type zaplog struct {
	core    *zap.Logger
	ctxKeys []string
}

func NewZapLogger(configure config.Configure) gplog.Logger {
	var cf conf
	if err := configure.Unmarshal("gin-plus.log", &cf); err != nil {
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
		StacktraceKey:    "stack",
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

	zl := &zaplog{
		core:    zap.New(zapCore),
		ctxKeys: cf.CtxKeys,
	}

	zl.Info(fmt.Sprintf("Logging system initialization completed: default level set to [%s]", strings.ToUpper(cf.Level)))
	return zl
}

func (d *zaplog) Info(text string, fields ...gplog.Field) {
	d.core.Info(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Debug(text string, fields ...gplog.Field) {
	d.core.Debug(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Warn(text string, fields ...gplog.Field) {
	d.core.Warn(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Error(text string, fields ...gplog.Field) {
	d.core.Error(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Fatal(text string, fields ...gplog.Field) {
	d.core.Fatal(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) InfoWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Info(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) DebugWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Debug(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) WarnWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Warn(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) ErrorWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Error(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) FatalWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Fatal(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) buildFields(ctx context.Context, gpFields []gplog.Field) []zap.Field {
	totalCapacity := len(gpFields) + len(d.ctxKeys)
	if totalCapacity == 0 {
		return nil
	}

	zapFields := make([]zap.Field, 0, totalCapacity)
	for _, f := range gpFields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	if ctx != nil {
		for _, key := range d.ctxKeys {
			if value := ctx.Value(key); value != nil {
				zapFields = append(zapFields, zap.Any(key, value))
			}
		}
	}

	return zapFields
}
