package kattest

import (
	"context"
	"github.com/mobiletoly/gokatana/katapp"
	"log/slog"
	"os"
)

func AppTestContext() context.Context {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := katapp.ContextWithAppLogger(logger)
	ctx = katapp.ContextWithRunInTest(ctx, true)
	return ctx
}
