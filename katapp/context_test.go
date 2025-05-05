package katapp

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestContextWithRunInTest(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithRunInTest(ctx, true)

	value := RunningInTest(ctx)
	if !value {
		t.Errorf("Expected RunningInTest to return true, got false")
	}

	ctx = ContextWithRunInTest(ctx, false)
	value = RunningInTest(ctx)
	if value {
		t.Errorf("Expected RunningInTest to return false, got true")
	}
}

func TestRunInTest_DefaultValue(t *testing.T) {
	ctx := context.Background()
	value := RunningInTest(ctx)

	if value {
		t.Errorf("Expected default RunningInTest to return false, got true")
	}
}

func TestStartContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := StartContext(logger, "test")
	if !RunningInTest(ctx) {
		t.Errorf("Expected RunningInTest to return true for 'test' deployment, got false")
	}

	// Test with non-"test" deployment
	ctx = StartContext(logger, "production")
	if RunningInTest(ctx) {
		t.Errorf("Expected RunningInTest to return false for 'production' deployment, got true")
	}
}
