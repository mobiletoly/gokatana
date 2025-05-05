package katapp

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type loggerContextKey struct{}

const RequestIdKey = "requestId"
const ScopeIdKey = "scopeId"

var stdLog = log.New(os.Stderr, "", log.LstdFlags)

type KatLogger struct {
	*slog.Logger
	ctx context.Context
}

func ContextWithAddedGroup(ctx context.Context, group string) context.Context {
	logger := Logger(ctx).WithGroup(group)
	return context.WithValue(ctx, loggerContextKey{}, logger.Logger)
}

func ContextWithRequestLogger(ctx context.Context, logger *slog.Logger, reqID string) context.Context {
	logger = logger.With(RequestIdKey, reqID)
	ctx = context.WithValue(ctx, RequestIdKey, reqID)
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func ContextWithAppLogger(logger *slog.Logger) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIdKey, "_app_")
	logger = logger.With(RequestIdKey, "_app_")
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func Logger(ctx context.Context) *KatLogger {
	logger := ctx.Value(loggerContextKey{}).(*slog.Logger)
	if logger == nil {
		panic("this context does not contain the required slog.Logger")
	}
	return &KatLogger{
		Logger: logger,
		ctx:    ctx,
	}
}

func (l *KatLogger) WithGroup(group string) *KatLogger {
	return &KatLogger{
		Logger: l.Logger.WithGroup(group),
		ctx:    l.ctx,
	}
}

func (l *KatLogger) With(args ...any) *KatLogger {
	return &KatLogger{
		Logger: l.Logger.With(args...),
		ctx:    l.ctx,
	}
}

func (l *KatLogger) Debugf(format string, a ...any) {
	l.log(slog.LevelDebug, fmt.Sprintf(format, a...))
}

func (l *KatLogger) Infof(format string, a ...any) {
	l.log(slog.LevelInfo, fmt.Sprintf(format, a...))
}

func (l *KatLogger) Warnf(format string, a ...any) {
	l.log(slog.LevelWarn, fmt.Sprintf(format, a...))
}

func (l *KatLogger) Errorf(format string, a ...any) {
	l.log(slog.LevelError, fmt.Sprintf(format, a...))
}

func (l *KatLogger) LogNewError(format string, a ...any) error {
	e := fmt.Errorf(format, a...)
	l.log(slog.LevelError, e.Error())
	return e
}

func (l *KatLogger) Fatal(msg string, fields ...any) {
	l.log(slog.LevelError, "(FATAL) "+msg, fields...)
	os.Exit(1)
}

func (l *KatLogger) Fatalf(format string, a ...any) {
	l.log(slog.LevelError, "(FATAL) "+fmt.Sprintf(format, a...))
	os.Exit(1)
}

func (l *KatLogger) log(level slog.Level, msg string, fields ...interface{}) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(fields...)
	_ = l.Handler().Handle(l.ctx, r)
}
