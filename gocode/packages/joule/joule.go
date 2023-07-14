package joule

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type UnitOfWork interface {
	GetId() string
	SetInput(interface{})
	GetOutput() interface{}
	GetError() error
	UiSecret(string) string
	UiOut([]byte)
	UiErr([]byte)
	UiStrOut(string)
	UiStrErr(string)
	UiOutWriter() io.Writer
	UiErrWriter() io.Writer
}

type usp struct { // worker request secret input
	id     string
	prompt string
}

type sUOW struct {
	id             string
	input          interface{}
	output         interface{}
	err            error
	uiSecretPrompt chan usp         // worker request secret input
	uiSecret       chan string      // runner provides secret input
	uiInReq        chan string      // worker request read from os.stdin
	uiIn           chan []byte      // runner provides os.stdin until EOF
	uiOut          chan []byte      // worker writes to os.stdout
	uiErr          chan []byte      // worker writes to os.stderr
	cliIn          chan interface{} // response to uiSecret/uiIn request
	work           func(context.Context, UnitOfWork, interface{}) (interface{}, error)
}

func (uow *sUOW) GetId() string { return uow.id }

func (uow *sUOW) SetInput(input interface{}) { uow.input = input }

func (uow *sUOW) GetOutput() interface{} { return uow.output }

func (uow *sUOW) GetError() error { return uow.err }

func (uow *sUOW) UiSecret(prompt string) string {
	uow.uiSecretPrompt <- usp{id: uow.id, prompt: prompt}
	return <-uow.uiSecret
}

func (uow *sUOW) UiOut(bs []byte) { uow.uiOut <- bs }

func (uow *sUOW) UiErr(bs []byte) { uow.uiErr <- bs }

func (uow *sUOW) UiStrOut(s string) { uow.uiOut <- []byte(s) }

func (uow *sUOW) UiStrErr(s string) { uow.uiErr <- []byte(s) }

func (uow *sUOW) UiOutWriter() io.Writer { return newC2w(uow.uiOut) }

func (uow *sUOW) UiErrWriter() io.Writer { return newC2w(uow.uiErr) }

type CLIRunner[OT any] struct {
	Ctx            *context.Context
	mux            sync.Mutex
	isRunning      bool
	isStopping     bool
	workDelay      time.Duration
	cancel         context.CancelFunc
	workersWg      sync.WaitGroup
	finalizer      func()
	uiSecretPrompt chan usp    // worker request secret input
	uiSecret       chan string // runner provides secret input
	uiInReq        chan string // worker request read from os.stdin
	uiIn           chan []byte // runner provides os.stdin until EOF
	uiOut          chan []byte // workers write to os.stdout
	uiErr          chan []byte // workers write to os.stderr
	Opts           OT
	Args           []string
	stdout         io.Writer
	stderr         io.Writer
	startup        func(cr *CLIRunner[OT]) error
	shutdown       func(cr *CLIRunner[OT]) error
	uows           []*sUOW
	uowReg         map[string]*sUOW
}

func NewCLIRunner[OT any](
	opts OT, args []string,
	stdin io.Reader, stdout io.Writer, stderr io.Writer,
	startup func(cr *CLIRunner[OT]) error,
	shutdown func(cr *CLIRunner[OT]) error,
) *CLIRunner[OT] {
	cr := &CLIRunner[OT]{
		Opts: opts, Args: args,
		stdout: stdout, stderr: stderr,
		startup: startup, shutdown: shutdown}
	cr.finalizer = cr.initAndGetFinalizer()
	return cr
}

func (cr *CLIRunner[OT]) SetWorkDelay(workDelay time.Duration) { cr.workDelay = workDelay }

func (cr *CLIRunner[OT]) AddUow(
	id string,
	work func(context.Context, UnitOfWork, interface{}) (interface{}, error),
) UnitOfWork {
	if cr.isStopping {
		return nil
	}
	cr.mux.Lock()
	defer cr.mux.Unlock()
	if id == "" {
		id = uuid.New().String()
	}
	uow := sUOW{
		id: id, work: work,
		uiSecretPrompt: cr.uiSecretPrompt, uiSecret: cr.uiSecret,
		uiInReq: cr.uiInReq, uiIn: cr.uiIn,
		uiOut: cr.uiOut, uiErr: cr.uiErr,
	}
	cr.uows = append(cr.uows, &uow)
	if cr.uowReg == nil {
		cr.uowReg = map[string]*sUOW{}
	}
	cr.uowReg[id] = &uow
	cr.workersWg.Add(1)
	if cr.isRunning {
		cr.controlWork(&uow)
	}
	return &uow
}

func (cr *CLIRunner[OT]) GetUow(id string) UnitOfWork { return cr.uowReg[id] }

func (cr *CLIRunner[OT]) initAndGetFinalizer() func() {
	var ctx context.Context
	ctx, cr.cancel = context.WithCancel(context.Background())
	cr.Ctx = &ctx
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		count := 0
		for sig := range c {
			count += 1
			if count == 1 {
				fmt.Fprintf(os.Stderr, "signal %s received, preparing to exit\n", sig)
				cr.cancel()
				continue
			}
			if count == 2 {
				fmt.Fprintf(os.Stderr, "signal %s received twice, send it again to force exit\n", sig)
				continue
			}
			fmt.Fprintf(os.Stderr, "signal %s received 3 times, exiting now\n", sig)
			os.Exit(1)
		}
	}()
	cr.uiSecretPrompt = make(chan usp)
	cr.uiSecret = make(chan string)
	cr.uiOut = make(chan []byte)
	cr.uiErr = make(chan []byte)
	finalize := func() {
		close(cr.uiSecretPrompt)
		close(cr.uiOut)
		close(cr.uiErr)
	}
	return finalize
}

func (cr *CLIRunner[OT]) controlWork(uow *sUOW) {
	go func() {
		if cr.workDelay != 0 {
			time.Sleep(cr.workDelay)
		}
		uow.output, uow.err = uow.work(*cr.Ctx, uow, uow.input)
		cr.workersWg.Done()
	}()
}

func (cr *CLIRunner[OT]) handleSecretPrompt(uow UnitOfWork, prompt string) {
	_, _ = os.Stderr.WriteString(prompt)
	secret, err := terminal.ReadPassword(int(syscall.Stdin))
	_ = err
	fmt.Println()
	cr.uiSecret <- string(secret)
}

func (cr *CLIRunner[OT]) handleInReq(uow UnitOfWork) {
	for {
		data := make([]byte, 8192)
		n, _ := os.Stdin.Read(data)
		if n > 0 {
			cr.uiIn <- data[0:n]
		}
	}
}

func (cr *CLIRunner[OT]) controlUI(done chan interface{}) (completed chan interface{}) {
	completed = make(chan interface{})
	go func() {
		defer close(completed)
		for {
			select {
			case uiUsp := <-cr.uiSecretPrompt:
				cr.handleSecretPrompt(cr.GetUow(uiUsp.id), uiUsp.prompt)
			case id := <-cr.uiInReq:
				cr.handleInReq(cr.GetUow(id))
			case uiOut := <-cr.uiOut:
				cr.stdout.Write([]byte(uiOut))
			case uiErr := <-cr.uiErr:
				cr.stderr.Write([]byte(uiErr))
			case <-done:
				return
			}
		}
	}()
	return
}

func (cr *CLIRunner[OT]) Run() error {
	defer cr.finalizer()
	if cr.startup != nil {
		if err := cr.startup(cr); err != nil {
			return err
		}
	}
	cr.mux.Lock()
	for _, uow := range cr.uows {
		cr.controlWork(uow)
	}
	cr.mux.Unlock()

	cr.isRunning = true
	stopUi := make(chan interface{})
	uiStopped := cr.controlUI(stopUi)
	cr.workersWg.Wait()
	cr.isStopping = true
	stopUi <- nil
	<-uiStopped

	if cr.shutdown != nil {
		if err := cr.shutdown(cr); err != nil {
			return err
		}
	}
	return nil
}

func (cr *CLIRunner[OT]) CancelFunc() context.CancelFunc {
	return cr.cancel
}
