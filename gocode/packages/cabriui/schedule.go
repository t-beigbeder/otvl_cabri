package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"gopkg.in/yaml.v3"
	_ "gopkg.in/yaml.v3"
	"os"
	"time"
)

type ScheduleOptions struct {
	BaseOptions
	HasLog   bool
	SpecFile string
	HasHttp  bool
	Address  string
}

type ScheduleVars struct {
	baseVars
}

func ScheduleStartup(cr *joule.CLIRunner[ScheduleOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[ScheduleOptions, *ScheduleVars](ctx)).vars =
				&ScheduleVars{baseVars: baseVars{uow: work}}
			return nil, schedule(ctx, cr)
		})
	return nil
}

func ScheduleShutdown(cr *joule.CLIRunner[ScheduleOptions]) error {
	return cr.GetUow("command").GetError()
}

func scheduleCtx(ctx context.Context) *uiContext[ScheduleOptions, *ScheduleVars] {
	return uiCtxFrom[ScheduleOptions, *ScheduleVars](ctx)
}

func scheduleVars(ctx context.Context) *ScheduleVars { return (*scheduleCtx(ctx)).vars }

func scheduleOpts(ctx context.Context) ScheduleOptions { return (*scheduleCtx(ctx)).opts }

func scheduleUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[ScheduleOptions, *ScheduleVars](ctx)
}

func scheduleOut(ctx context.Context, s string) { scheduleUow(ctx).UiStrOut(s) }

func scheduleErr(ctx context.Context, s string) { scheduleUow(ctx).UiStrErr(s) }

func schedule(ctx context.Context, cr *joule.CLIRunner[ScheduleOptions]) error {
	opts := scheduleOpts(ctx)
	vars := scheduleVars(ctx)
	_, _ = opts, vars
	ure, err := GetUiRunEnv[ScheduleOptions, *ScheduleVars](ctx, false, false)
	if err != nil {
		return err
	}
	_ = ure
	bs, err := os.ReadFile(opts.SpecFile)
	if err != nil {
		return err
	}
	var spec CabriScheduleSpec
	if err = yaml.Unmarshal(bs, &spec); err != nil {
		return err
	}
	t, err := yaml.Marshal(spec)
	_ = t
	scheduleErr(ctx, fmt.Sprintf("Running %v\n", os.Args))
	sc := ScheduleConfig{ctx: ctx, cancel: cr.CancelFunc(), Spec: spec, run: map[string]*ScheduleRunStatus{}}
	for k, _ := range spec {
		sc.run[k] = &ScheduleRunStatus{label: k, LastTime: time.Now().UnixNano()}
	}
	cr.SetWorkDelay(time.Second)
	var ws cabridss.WebServer
	if opts.HasHttp {
		ws = cabridss.NewEServer(opts.Address, opts.HasLog, nil)
		ws.ConfigureApi("", &sc, nil, SchedServerConfigurator)
		if err := ws.Serve(); err != nil {
			return err
		}
	}

	next := time.Second
SCHED:
	for {
		select {
		case <-ctx.Done():
			break SCHED
		case <-time.After(next):
			next, _ = Schedule(&sc)
		}
	}

	if ws != nil {
		if err := ws.Shutdown(); err != nil {
			webApiErr(ctx, fmt.Sprintf("server at %s shutdown failed with error %v\n", opts.Address, err))
		}
	}
	return nil
}
