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

// callerSkip makes the reported caller the code calling the gplog package-level functions:
// user code -> gplog.Info -> zapLogger.Info -> zapLogger.log -> zap.Logger.Check.
// Calling the Logger interface directly (gplog.GetLogger().Info) reports one frame higher.
const callerSkip = 3

type conf struct {
	// Level syslog level, default info.
	// 	Supports: error、info、warn、panic、fatal、debug
	Level string `mapstructure:"level"`

	// Formatter syslog format, default console (supports: json、console)
	Format string `mapstructure:"format"`

	// ConsoleSeparator console separator.
	// Note: when the formatter is console, the separator between the fields, default is "\t".
	ConsoleSeparator string `mapstructure:"console-separator"`

	// Caller whether to add the caller (file:line) to each log entry, default false.
	Caller bool `mapstructure:"caller"`

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

// newCore builds a JSON core writing to a rotating file.
// ec must not use a color level encoder, ANSI codes would end up in the file.
func (f *fileConfig) newCore(level zapcore.Level, ec zapcore.EncoderConfig) zapcore.Core {
	filename := f.Filename
	if filename == "" {
		filename = "logs/app.log"
	}
	maxSize := f.MaxSize
	if maxSize == 0 {
		maxSize = 100 // 100MB
	}

	writer := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     f.MaxAge,
		MaxBackups: f.MaxBackups,
		Compress:   f.Compress,
		LocalTime:  true,
	}
	return zapcore.NewCore(zapcore.NewJSONEncoder(ec), zapcore.AddSync(writer), level)
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
// It panics on invalid configuration, since it runs before any logger is available.
func NewLogger(cp config.Provider, opts ...Option) gplog.Logger {
	var cf conf
	if err := cp.Unmarshal("gin-plus.log", &cf); err != nil {
		panic("Failed to initialize logging system: unable to read configuration. Error: " + err.Error())
	}

	if cf.Level == "" {
		cf.Level = "info"
	}
	switch cf.Format {
	case "":
		cf.Format = gplog.ConsoleFormat
	case gplog.ConsoleFormat, gplog.JSONFormat:
	default:
		panic(fmt.Sprintf("Failed to initialize logging system: invalid log format [%s], must be one of %s, %s",
			cf.Format, gplog.ConsoleFormat, gplog.JSONFormat))
	}
	if cf.ConsoleSeparator == "" {
		cf.ConsoleSeparator = "\t"
	}

	zapLevel, err := zapcore.ParseLevel(cf.Level)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logging system: invalid log level [%s]. Error: %s", cf.Level, err.Error()))
	}

	// Plain encoder config, shared by JSON stdout and file output.
	ec := zapcore.EncoderConfig{
		TimeKey:          "time",
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

	var stdoutEncoder zapcore.Encoder
	if cf.Format == gplog.ConsoleFormat {
		// EncoderConfig is a value type, the colored copy never leaks into the file core.
		colored := ec
		colored.EncodeLevel = zapcore.CapitalColorLevelEncoder
		stdoutEncoder = zapcore.NewConsoleEncoder(colored)
	} else {
		stdoutEncoder = zapcore.NewJSONEncoder(ec)
	}

	zl := &zapLogger{
		format: cf.Format,
		keys:   cf.Keys,
	}
	for _, opt := range opts {
		opt(zl)
	}

	cores := []zapcore.Core{zapcore.NewCore(stdoutEncoder, zapcore.Lock(os.Stdout), zapLevel)}
	if cf.File != nil && cf.File.Enable {
		cores = append(cores, cf.File.newCore(zapLevel, ec))
	}
	cores = append(cores, zl.additionalCores...)

	zapOpts := zl.zapOptions
	if cf.Caller {
		// Prepend so user options (e.g. a custom AddCallerSkip) are applied on top.
		zapOpts = append([]zap.Option{zap.AddCaller(), zap.AddCallerSkip(callerSkip)}, zapOpts...)
	}

	// NewTee returns the core itself when there is only one.
	zl.core = zap.New(zapcore.NewTee(cores...), zapOpts...)
	return zl
}

func (d *zapLogger) GetFormat() string {
	return d.format
}

// Non-ctx variants pass a nil context, so buildFields skips the context key lookups.

func (d *zapLogger) Info(text string, keyvals ...any) {
	d.log(nil, zapcore.InfoLevel, text, keyvals)
}

func (d *zapLogger) Debug(text string, keyvals ...any) {
	d.log(nil, zapcore.DebugLevel, text, keyvals)
}

func (d *zapLogger) Warn(text string, keyvals ...any) {
	d.log(nil, zapcore.WarnLevel, text, keyvals)
}

func (d *zapLogger) Error(text string, keyvals ...any) {
	d.log(nil, zapcore.ErrorLevel, text, keyvals)
}

func (d *zapLogger) Fatal(text string, keyvals ...any) {
	d.log(nil, zapcore.FatalLevel, text, keyvals)
}

func (d *zapLogger) InfoWithCtx(ctx context.Context, text string, keyvals ...any) {
	d.log(ctx, zapcore.InfoLevel, text, keyvals)
}

func (d *zapLogger) DebugWithCtx(ctx context.Context, text string, keyvals ...any) {
	d.log(ctx, zapcore.DebugLevel, text, keyvals)
}

func (d *zapLogger) WarnWithCtx(ctx context.Context, text string, keyvals ...any) {
	d.log(ctx, zapcore.WarnLevel, text, keyvals)
}

func (d *zapLogger) ErrorWithCtx(ctx context.Context, text string, keyvals ...any) {
	d.log(ctx, zapcore.ErrorLevel, text, keyvals)
}

func (d *zapLogger) FatalWithCtx(ctx context.Context, text string, keyvals ...any) {
	d.log(ctx, zapcore.FatalLevel, text, keyvals)
}

// log checks the level before building fields, so disabled levels cost no allocation.
// For FatalLevel, Check always returns an entry whose Write exits the process.
func (d *zapLogger) log(ctx context.Context, level zapcore.Level, text string, keyvals []any) {
	if ce := d.core.Check(level, text); ce != nil {
		ce.Write(d.buildFields(ctx, keyvals)...)
	}
}

func (d *zapLogger) buildFields(ctx context.Context, keyvals []any) []zap.Field {
	// Round up, an odd trailing key still produces a field.
	capacity := (len(keyvals) + 1) / 2
	if ctx != nil {
		capacity += len(d.keys)
	}
	if capacity == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, capacity)
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprint(keyvals[i])
		}
		var val any
		if i+1 < len(keyvals) {
			val = keyvals[i+1]
		}
		fields = append(fields, zap.Any(key, val))
	}

	if ctx != nil {
		for _, key := range d.keys {
			if value := ctx.Value(key); value != nil {
				fields = append(fields, zap.Any(key, value))
			}
		}
	}
	return fields
}
