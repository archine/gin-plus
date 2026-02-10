package zapper

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type conf struct {
	// Level syslog level, default info.
	// 	Supports: error、info、warn、panic、fatal、debug
	Level string `mapstructure:"level"`

	// Formatter syslog format, default console (supports: json、console)
	Format string `mapstructure:"format"`

	// ConsoleSeparator console separator.
	// Note: when the formatter is console, the separator between the fields, default is "\t".
	ConsoleSeparator string `mapstructure:"console-separator"`

	// Keys When using WithContext for syslog output.
	// the value of the specified key is obtained from the context and added to the syslog.
	Keys []string `mapstructure:"keys"`

	// File output configuration.
	File *fileConfig `mapstructure:"file"`
}

// fileConfig represents file output configuration with log rotation support.
type fileConfig struct {
	// Enable whether to enable file output, default false.
	Enable bool `mapstructure:"enable"`

	// Filename is the file to write logs to, default "logs/app.log".
	Filename string `mapstructure:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	// Default is 100 MB.
	MaxSize int `mapstructure:"max-size"`

	// MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
	// Default is not to remove old log files based on age.
	MaxAge int `mapstructure:"max-age"`

	// MaxBackups is the maximum number of old log files to retain.
	// Default is to retain all old log files.
	MaxBackups int `mapstructure:"max-backups"`

	// Compress determines if the rotated log files should be compressed using gzip.
	// Default is false.
	Compress bool `mapstructure:"compress"`
}

func (f *fileConfig) transform(level zapcore.Level, ec zapcore.EncoderConfig) zapcore.Core {
	if f.Filename == "" {
		f.Filename = "logs/app.log"
	}
	if f.MaxSize == 0 {
		f.MaxSize = 100 // 100MB
	}

	lumberLogger := &lumberjack.Logger{
		Filename:   f.Filename,
		MaxSize:    f.MaxSize,
		MaxAge:     f.MaxAge,
		MaxBackups: f.MaxBackups,
		Compress:   f.Compress,
		LocalTime:  true,
	}
	fileEncoder := zapcore.NewJSONEncoder(ec)

	return zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(lumberLogger),
		level,
	)
}

// Option is a function that configures a zapLogger.
type Option func(*zapLogger)

// WithCores adds additional zapper cores to the logger.
// This allows you to output logs to multiple destinations (e.g., file, remote service).
func WithCores(cores ...zapcore.Core) Option {
	return func(zl *zapLogger) {
		zl.additionalCores = append(zl.additionalCores, cores...)
	}
}

// WithOptions adds additional zapper options to the logger.
func WithOptions(opts ...zap.Option) Option {
	return func(zl *zapLogger) {
		zl.zapOptions = append(zl.zapOptions, opts...)
	}
}

// zapLogger is the zapper-based implementation of the gplog.Logger interface.
type zapLogger struct {
	format          string
	core            *zap.Logger
	keys            []string
	additionalCores []zapcore.Core
	zapOptions      []zap.Option
}

// NewLogger creates a new zapper-based logger.
func NewLogger(cp config.Provider, opts ...Option) gplog.Logger {
	var cf conf
	if err := cp.Unmarshal("gin-plus.log", &cf); err != nil {
		panic("Failed to initialize logging system: unable to read configuration. Error: " + err.Error())
	}

	if cf.Level == "" {
		cf.Level = "info"
	}
	if cf.Format == "" || (cf.Format != gplog.JSONFormat && cf.Format != gplog.ConsoleFormat) {
		cf.Format = gplog.ConsoleFormat
	}
	if cf.ConsoleSeparator == "" {
		cf.ConsoleSeparator = "\t"
	}

	zapLevel, err := zapcore.ParseLevel(cf.Level)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logging system: invalid log level [%s]. Error: %s", cf.Level, err.Error()))
	}

	ec := zapcore.EncoderConfig{
		TimeKey:          "timestamp",
		LevelKey:         "level",
		CallerKey:        "caller",
		MessageKey:       "msg",
		FunctionKey:      zapcore.OmitKey,
		EncodeTime:       zapcore.TimeEncoderOfLayout(time.DateTime),
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		LineEnding:       zapcore.DefaultLineEnding,
		StacktraceKey:    "stack",
		ConsoleSeparator: cf.ConsoleSeparator,
	}

	var encoder zapcore.Encoder

	if cf.Format == gplog.ConsoleFormat {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(ec)
	} else {
		encoder = zapcore.NewJSONEncoder(ec)
	}

	zapCore := zapcore.NewCore(encoder, os.Stderr, zapLevel)

	zl := &zapLogger{
		format: cf.Format,
		keys:   cf.Keys,
	}

	for _, opt := range opts {
		opt(zl)
	}

	if cf.File != nil && cf.File.Enable {
		zl.additionalCores = append(zl.additionalCores, cf.File.transform(zapLevel, ec))
	}

	// Combine default core with additional cores if any
	var finalCore zapcore.Core
	if len(zl.additionalCores) > 0 {
		allCores := append([]zapcore.Core{zapCore}, zl.additionalCores...)
		finalCore = zapcore.NewTee(allCores...)
	} else {
		finalCore = zapCore
	}

	zl.core = zap.New(finalCore, zl.zapOptions...)

	return zl
}

func (d *zapLogger) GetFormat() string {
	return d.format
}

func (d *zapLogger) Info(text string, fields ...gplog.Field) {
	d.core.Info(text, d.buildFields(context.TODO(), fields)...)
}

func (d *zapLogger) Debug(text string, fields ...gplog.Field) {
	d.core.Debug(text, d.buildFields(context.TODO(), fields)...)
}

func (d *zapLogger) Warn(text string, fields ...gplog.Field) {
	d.core.Warn(text, d.buildFields(context.TODO(), fields)...)
}

func (d *zapLogger) Error(text string, fields ...gplog.Field) {
	d.core.Error(text, d.buildFields(context.TODO(), fields)...)
}

func (d *zapLogger) Fatal(text string, fields ...gplog.Field) {
	d.core.Fatal(text, d.buildFields(context.TODO(), fields)...)
}

func (d *zapLogger) InfoWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Info(text, d.buildFields(ctx, fields)...)
}

func (d *zapLogger) DebugWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Debug(text, d.buildFields(ctx, fields)...)
}

func (d *zapLogger) WarnWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Warn(text, d.buildFields(ctx, fields)...)
}

func (d *zapLogger) ErrorWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Error(text, d.buildFields(ctx, fields)...)
}

func (d *zapLogger) FatalWithCtx(ctx context.Context, text string, fields ...gplog.Field) {
	d.core.Fatal(text, d.buildFields(ctx, fields)...)
}

func (d *zapLogger) buildFields(ctx context.Context, gpFields []gplog.Field) []zap.Field {
	if len(gpFields) == 0 && (ctx == nil || len(d.keys) == 0) {
		return []zap.Field{}
	}

	capacity := len(gpFields)
	if len(d.keys) > 0 {
		capacity += len(d.keys)
	}

	zapFields := make([]zap.Field, 0, capacity)

	for _, f := range gpFields {
		for key, value := range f {
			zapFields = append(zapFields, zap.Any(key, value))
		}
	}

	if ctx != nil {
		for _, key := range d.keys {
			if value := ctx.Value(key); value != nil {
				zapFields = append(zapFields, zap.Any(key, value))
			}
		}
	}

	return zapFields
}
