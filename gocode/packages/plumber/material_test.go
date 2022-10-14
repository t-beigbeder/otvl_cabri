package plumber

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRegulated(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithConfig(ctx,
		Config{PlizerEnabled: true, RglatorsByName: map[string]uint{"a": 0, "b": 1, "c": 2}})
	job := func(ctx context.Context, i string) (o string) {
		time.Sleep(100 * time.Millisecond)
		o = fmt.Sprintf("job(%s)", i)
		return
	}
	var wg sync.WaitGroup
	var o1, o2, o3 string

	wg.Add(2)
	go func() {
		o1 = Regulated[string, string](ctx, "b", job, "i1")
		wg.Done()
	}()
	go func() {
		o2 = Regulated[string, string](ctx, "b", job, "i2")
		wg.Done()
	}()
	wg.Wait()
	_ = o1
	_ = o2

	wg.Add(3)
	go func() {
		o1 = Regulated[string, string](ctx, "c", job, "i1")
		wg.Done()
	}()
	go func() {
		o2 = Regulated[string, string](ctx, "c", job, "i2")
		wg.Done()
	}()
	go func() {
		o3 = Regulated[string, string](ctx, "c", job, "i3")
		wg.Done()
	}()
	wg.Wait()
	_ = o1
	_ = o2
	_ = o3
}

func TestLaunchAndWait(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithConfig(ctx,
		Config{PlizerEnabled: true, RglatorsByName: map[string]uint{"b": 1, "c": 2}})
	job := func(ctx context.Context, i interface{}) (o interface{}) {
		time.Sleep(100 * time.Millisecond)
		o = fmt.Sprintf("%s - job(%v)", time.Now().Format("2006-01-02T15:04:05.000"), i)
		return
	}
	fmt.Printf("%s\n", job(ctx, "test"))
	outs := LaunchAndWait(ctx,
		[]string{"b", "c", "b", "c", "b", "c"},
		[]Launchable{job, job, job, job, job, job},
		[]interface{}{"i1", "i2", "i3", "i4", "i5", "i6"},
	)
	for _, out := range outs {
		fmt.Printf("out %v\n", out)
	}
}
