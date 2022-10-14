package joule

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

type tOpts struct {
	oVal int
}

func TestControlCTwice(t *testing.T) {
	optionalKeep(t)
	cr := NewCLIRunner[tOpts](
		tOpts{oVal: 42}, []string{"31"}, os.Stdin, os.Stdout, os.Stderr,
		//startup
		func(cr *CLIRunner[tOpts]) error {
			_ = cr.AddUow(
				"id1",
				func(ctx context.Context, uow UnitOfWork, input interface{}) (interface{}, error) {
					<-ctx.Done()
					time.Sleep(800 * time.Millisecond)
					return fmt.Sprintf("uow#1 processed %v", input), nil
				})
			go func() {
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
				time.Sleep(100 * time.Millisecond)
				_ = syscall.Kill(os.Getpid(), syscall.SIGINT)
			}()

			cr.GetUow("id1").SetInput(fmt.Sprintf("set to opts %v args %v", cr.Opts, cr.Args))
			return nil
		},

		//shutdown
		func(cr *CLIRunner[tOpts]) error {
			fmt.Fprintf(os.Stderr, "Shutdown uow output %v\n", cr.GetUow("id1").GetOutput())
			return nil
		},
	)
	cr.Run()
}

func TestCLIWG(t *testing.T) {
	cr := NewCLIRunner[tOpts](
		tOpts{oVal: 42}, []string{"31"}, os.Stdin, os.Stdout, os.Stderr,
		func(cr *CLIRunner[tOpts]) error {
			for i := 0; i < 10; i++ {
				i := i
				cr.AddUow(
					fmt.Sprintf("uow%d", i),
					func(ctx context.Context, uow UnitOfWork, input interface{}) (interface{}, error) {
						return fmt.Sprintf("uow#%d processed %v", i, input), nil
					})
			}
			for i := 0; i < 10; i++ {
				cr.GetUow(fmt.Sprintf("uow%d", i)).SetInput(fmt.Sprintf("#%d set to opts %v args %v", i, cr.Opts, cr.Args))
			}
			return nil
		},
		func(cr *CLIRunner[tOpts]) error {
			for i := 0; i < 10; i++ {
				fmt.Fprintf(os.Stderr, "Shutdown uow output %d %v\n", i,
					cr.GetUow(fmt.Sprintf("uow%d", i)).GetOutput())
			}
			return nil
		},
	)
	cr.Run()
}

func TestCLIUIOut(t *testing.T) {
	cr := NewCLIRunner[string](
		"o", []string{"a"}, os.Stdin, os.Stdout, os.Stderr,
		func(cr *CLIRunner[string]) error {
			_ = cr.AddUow("id",
				func(ctx context.Context, uow UnitOfWork, input interface{}) (interface{}, error) {
					uow.UiStrOut("A message from uow on stdout\n")
					uow.UiStrErr("A message from uow on stderr\n")
					fmt.Fprintf(uow.UiOutWriter(), "A message from uow on stdout with Fprintf\n")
					fmt.Fprintf(uow.UiErrWriter(), "A message from uow on stderr with Fprintf\n")
					s := uow.UiSecret("prompt")
					uow.UiStrOut(fmt.Sprintf("A secret from uow: %s\n", s))
					return fmt.Sprintf("uow processed %v", input), nil
				})
			cr.GetUow("id").SetInput(fmt.Sprintf("set to opts %v args %v", cr.Opts, cr.Args))
			return nil
		},
		func(cr *CLIRunner[string]) error {
			fmt.Fprintf(os.Stderr, "Shutdown uow output %v\n", cr.GetUow("id").GetOutput())
			return nil
		},
	)

	cr.Run()
}
