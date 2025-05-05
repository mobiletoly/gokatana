package katapp

import (
	"context"
	"log/slog"
)

type runInTestContextKey struct{}

// ContextWithRunInTest returns a new context with the given value associated with the key.
func ContextWithRunInTest(ctx context.Context, runInTest bool) context.Context {
	return context.WithValue(ctx, runInTestContextKey{}, runInTest)
}

// RunningInTest returns the value associated with the key from the context.
func RunningInTest(ctx context.Context) bool {
	runInTest, ok := ctx.Value(runInTestContextKey{}).(bool)
	if !ok {
		return false
	}
	return runInTest
}

func StartContext(logger *slog.Logger, deployment string) context.Context {
	ctx := ContextWithAppLogger(logger)
	if deployment == "test" {
		ctx = ContextWithRunInTest(ctx, true)
	}
	slog.SetDefault(logger)
	return ctx
}
