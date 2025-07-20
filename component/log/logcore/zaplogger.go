package logcore

import (
	"context"
	"fmt"
	"github.com/archine/gin-plus/v4/component/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type conf struct {
	// Level syslog level, default info.
	// 	Supports: error、info、warn、panic、fatal、debug
	Level string `mapstructure:"level"`

	// Formatter syslog format, default console (supports: json、console)
	Format string `mapstructure:"format"`

	// ConsoleSeparator console separator.
	// Note: when the formatter is console, the separator between the fields, default is "\t".
	ConsoleSeparator string `mapstructure:"console-sep"`

	// CtxKeys When using WithContext for syslog output.
	// the value of the specified key is obtained from the context and added to the syslog.
	CtxKeys []string `mapstructure:"ctx-keys"`
}

// defaultLogger is the default implementation of the syslog interface.
type zaplog struct {
	core    *zap.Logger
	ctxKeys []string
}

func NewZapLogger(cp config.Provider) Logger {
	var cf conf
	if err := cp.Unmarshal("gin-plus.log", &cf); err != nil {
		panic("initialization of logging system failed, unable to read configuration: " + err.Error())
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
		panic(fmt.Sprintf("initialization of logging system failed, invalid log level [%s]: %s", cf.Level, err.Error()))
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

	if cf.Format == "console" {
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

	return zl
}

func (d *zaplog) Info(text string, fields ...Field) {
	d.core.Info(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Debug(text string, fields ...Field) {
	d.core.Debug(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Warn(text string, fields ...Field) {
	d.core.Warn(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Error(text string, fields ...Field) {
	d.core.Error(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) Fatal(text string, fields ...Field) {
	d.core.Fatal(text, d.buildFields(nil, fields)...)
}

func (d *zaplog) InfoWithCtx(ctx context.Context, text string, fields ...Field) {
	d.core.Info(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) DebugWithCtx(ctx context.Context, text string, fields ...Field) {
	d.core.Debug(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) WarnWithCtx(ctx context.Context, text string, fields ...Field) {
	d.core.Warn(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) ErrorWithCtx(ctx context.Context, text string, fields ...Field) {
	d.core.Error(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) FatalWithCtx(ctx context.Context, text string, fields ...Field) {
	d.core.Fatal(text, d.buildFields(ctx, fields)...)
}

func (d *zaplog) buildFields(ctx context.Context, gpFields []Field) []zap.Field {
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
