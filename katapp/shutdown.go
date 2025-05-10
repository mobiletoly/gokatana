package katapp

import (
	"context"
	"os"
	"os/signal"
	"time"
)

// WaitForInterruptSignal waits for interrupt signal to gracefully shut down the server with a timeout.
// Use a buffered channel to avoid missing signals as recommended for signal.Notify
func WaitForInterruptSignal(ctx context.Context, timeout time.Duration, shutdown func() error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	Logger(ctx).InfoContext(ctx, "Interrupt signal has been received")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := shutdown(); err != nil {
		Logger(ctx).WarnContext(ctx, "Shutdown has failed with error: %s", err.Error())
	}
	Logger(ctx).InfoContext(ctx, "App is shutdown")
}
