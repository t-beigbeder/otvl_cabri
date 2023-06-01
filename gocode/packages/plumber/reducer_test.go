package plumber

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestReducerSimple(t *testing.T) {
	if os.Getenv("PLUMBER_KEEP_DEV_TESTS") == "" {
		t.Skip(fmt.Sprintf("Skipping %s because you didn't set PLUMBER_KEEP_DEV_TESTS", t.Name()))
	}
	red := NewReducer(0, 0)
	out := os.Stdout
	red.SetDbgOut(out)
	for it := 0; it < 30; it++ {
		i := it
		go func() {
			st := red.Launch(fmt.Sprintf("l%d", i), func() error {
				fmt.Fprintf(out, "#%d sleeping %d\n", i, time.Duration(100*i))
				time.Sleep(time.Duration(100*i) * time.Millisecond)
				return nil
			})
			if st != nil {
				fmt.Fprintf(out, "#%d error %v\n", i, st)
			}
		}()
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	time.Sleep(time.Duration(10) * time.Second)
	red.Close()
}

func TestReducerClose(t *testing.T) {
	red := NewReducer(0, 0)
	out := os.Stdout
	red.SetDbgOut(out)
	for it := 0; it < 30; it++ {
		i := it
		go func() {
			st := red.Launch(fmt.Sprintf("l%d", i), func() error {
				fmt.Fprintf(out, "#%d sleeping %d\n", i, time.Duration(100*i))
				time.Sleep(time.Duration(100*i) * time.Millisecond)
				return nil
			})
			if st != nil {
				fmt.Fprintf(out, "#%d error %v\n", i, st)
			}
		}()
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	time.Sleep(time.Duration(2) * time.Second)
	red.Close()
}
