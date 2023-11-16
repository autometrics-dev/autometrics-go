package log // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/log"

import (
	"context"
	"fmt"
)

// Logger is an interface to implement to be able to inject a logger to Autometrics.
// The interface follows the interface of slog.Logger
type Logger interface {
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

// NoOpLogger is the default logger for Autometrics. It does nothing.
type NoOpLogger struct{}

var _ Logger = NoOpLogger{}

func (_ NoOpLogger) Debug(msg string, args ...any)                             {}
func (_ NoOpLogger) DebugContext(ctx context.Context, msg string, args ...any) {}
func (_ NoOpLogger) Info(msg string, args ...any)                              {}
func (_ NoOpLogger) InfoContext(ctx context.Context, msg string, args ...any)  {}
func (_ NoOpLogger) Warn(msg string, args ...any)                              {}
func (_ NoOpLogger) WarnContext(ctx context.Context, msg string, args ...any)  {}
func (_ NoOpLogger) Error(msg string, args ...any)                             {}
func (_ NoOpLogger) ErrorContext(ctx context.Context, msg string, args ...any) {}

// PrintLogger is a simple logger implementation that simply prints the events to stdout
type PrintLogger struct{}

var _ Logger = PrintLogger{}

func (_ PrintLogger) Debug(msg string, args ...any) {
	fmt.Printf("Autometrics - Debug: %v", fmt.Sprintf(msg, args...))
}
func (_ PrintLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	fmt.Printf("Autometrics - Debug: %v", fmt.Sprintf(msg, args...))
}
func (_ PrintLogger) Info(msg string, args ...any) {
	fmt.Printf("Autometrics - Info: %v", fmt.Sprintf(msg, args...))
}
func (_ PrintLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	fmt.Printf("Autometrics - Info: %v", fmt.Sprintf(msg, args...))
}
func (_ PrintLogger) Warn(msg string, args ...any) {
	fmt.Printf("Autometrics - Warn: %v", fmt.Sprintf(msg, args...))
}
func (_ PrintLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	fmt.Printf("Autometrics - Warn: %v", fmt.Sprintf(msg, args...))
}
func (_ PrintLogger) Error(msg string, args ...any) {
	fmt.Printf("Autometrics - Error: %v", fmt.Sprintf(msg, args...))
}
func (_ PrintLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	fmt.Printf("Autometrics - Error: %v", fmt.Sprintf(msg, args...))
}
