package plumber

import (
	"context"
	"io"
	"os"
)

type testUiContext struct {
	Ctx     context.Context
	OutW    io.Writer
	ErrW    io.Writer
	Err     error
	Options interface{}
	Vars    interface{}
	Payload interface{}
}

type key int

const uiCtxKey key = 0

func newContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, uiCtxKey, &testUiContext{Ctx: ctx, OutW: os.Stdout, ErrW: os.Stderr})
	return ctx, cancel
}

func fromContext(ctx context.Context) *testUiContext {
	return ctx.Value(uiCtxKey).(*testUiContext)
}
