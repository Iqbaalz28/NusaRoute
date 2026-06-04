package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const TraceIDKey contextKey = "TraceID"

var Log *zap.Logger

func init() {
	Log, _ = zap.NewDevelopment()
}
// InitLogger initializes the global structured logger.
func InitLogger(serviceName string) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	Log = zap.New(core).With(zap.String("service", serviceName))
	zap.ReplaceGlobals(Log)
}

// WithTrace returns a logger instance with the TraceID from the context, if present.
func WithTrace(ctx context.Context) *zap.Logger {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if ok && traceID != "" {
		return Log.With(zap.String("trace_id", traceID))
	}
	return Log
}

// GetTraceID returns the TraceID from the context, or an empty string.
func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if ok {
		return traceID
	}
	return ""
}

// Info logs an info message with context tracing.
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	WithTrace(ctx).Info(msg, fields...)
}

// Error logs an error message with context tracing.
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	WithTrace(ctx).Error(msg, fields...)
}

// Warn logs a warning message with context tracing.
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	WithTrace(ctx).Warn(msg, fields...)
}

// Debug logs a debug message with context tracing.
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	WithTrace(ctx).Debug(msg, fields...)
}
