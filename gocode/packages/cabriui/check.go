package cabriui

import (
	"context"
	"errors"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
)

type CheckOptions struct {
	BaseOptions
	S3Session bool
	S3List    bool
}

type CheckVars struct {
	baseVars
}

func CheckStartup(cr *joule.CLIRunner[CheckOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[CheckOptions, *CheckVars](ctx)).vars = &CheckVars{baseVars: baseVars{uow: work}}
			return nil, checkRun(ctx)
		})
	return nil
}

func CheckShutdown(cr *joule.CLIRunner[CheckOptions]) error {
	return cr.GetUow("command").GetError()
}

func checkCtx(ctx context.Context) *uiContext[CheckOptions, *CheckVars] {
	return uiCtxFrom[CheckOptions, *CheckVars](ctx)
}

func checkVars(ctx context.Context) *CheckVars { return (*checkCtx(ctx)).vars }

func checkOpts(ctx context.Context) CheckOptions { return (*checkCtx(ctx)).opts }

func checkUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[CheckOptions, *CheckVars](ctx)
}

func checkOut(ctx context.Context, s string) { checkUow(ctx).UiStrOut(s) }

func checkErr(ctx context.Context, s string) { checkUow(ctx).UiStrErr(s) }

func checkS3Session(ctx context.Context) error {
	opts := checkOpts(ctx)
	oc, err := GetObsConfig(opts.BaseOptions, 0, "", "")
	if err != nil {
		return err
	}
	is3 := cabridss.NewS3Session(oc, nil)
	if err = is3.Initialize(); err != nil {
		return err
	}
	if err = is3.Check(); err != nil {
		return err
	}
	return nil
}

func checkS3List(ctx context.Context) error {
	opts := checkOpts(ctx)
	args := checkCtx(ctx).args
	prefix := ""
	if len(args) != 0 {
		prefix = args[0]
	}
	oc, err := GetObsConfig(opts.BaseOptions, 0, "", "")
	if err != nil {
		return err
	}
	is3 := cabridss.NewS3Session(oc, nil)
	if err = is3.Initialize(); err != nil {
		return err
	}
	var ls internal.StringSliceEOL
	if ls, err = is3.List(prefix); err != nil {
		return err
	}
	checkOut(ctx, fmt.Sprintf("%s\n", ls))
	return nil
}

func checkRun(ctx context.Context) error {
	opts := checkOpts(ctx)
	err := fmt.Errorf("at least one operation option must be given with the check command")
	if opts.S3Session {
		err = checkS3Session(ctx)
	}
	if opts.S3List {
		err = checkS3List(ctx)
	}
	if err != nil {
		if errors.Is(err, cabridss.ErrPasswordRequired) {
			return cabridss.ErrPasswordRequired
		}
		return err
	}
	return nil
}
