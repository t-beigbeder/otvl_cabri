package plumber

import (
	"context"
	"testing"
	"time"
)

func TestContextWithConfig(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithConfig(ctx,
		Config{PlizerEnabled: true, RglatorsByName: map[string]uint{"a": 0, "b": 1, "c": 2}})
	cfg, ok := ConfigFromContext(ctx)
	if !ok {
		t.Fatal("ok")
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		cfg.rglatorChans["a"] <- struct{}{}
	}()
	<-cfg.rglatorChans["a"]
	go func() {
		time.Sleep(100 * time.Millisecond)
		cfg.rglatorChans["b"] <- struct{}{}
	}()
	<-cfg.rglatorChans["b"]
	<-cfg.rglatorChans["b"]
	go func() {
		time.Sleep(100 * time.Millisecond)
		cfg.rglatorChans["c"] <- struct{}{}
	}()
	<-cfg.rglatorChans["c"]
	<-cfg.rglatorChans["c"]
	<-cfg.rglatorChans["c"]
}
