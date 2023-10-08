package cabriui

import (
	"context"
	"errors"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
)

type S3ToolsOptions struct {
	BaseOptions
	S3Session bool
	S3List    bool
	S3Clone   bool
}

type S3ToolsVars struct {
	baseVars
}

func S3ToolsStartup(cr *joule.CLIRunner[S3ToolsOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[S3ToolsOptions, *S3ToolsVars](ctx)).vars = &S3ToolsVars{baseVars: baseVars{uow: work}}
			return nil, s3ToolsRun(ctx)
		})
	return nil
}

func S3ToolsShutdown(cr *joule.CLIRunner[S3ToolsOptions]) error {
	return cr.GetUow("command").GetError()
}

func s3ToolsCtx(ctx context.Context) *uiContext[S3ToolsOptions, *S3ToolsVars] {
	return uiCtxFrom[S3ToolsOptions, *S3ToolsVars](ctx)
}

func s3ToolsVars(ctx context.Context) *S3ToolsVars { return (*s3ToolsCtx(ctx)).vars }

func s3ToolsOpts(ctx context.Context) S3ToolsOptions { return (*s3ToolsCtx(ctx)).opts }

func s3ToolsUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[S3ToolsOptions, *S3ToolsVars](ctx)
}

func s3ToolsOut(ctx context.Context, s string) { s3ToolsUow(ctx).UiStrOut(s) }

func s3ToolsErr(ctx context.Context, s string) { s3ToolsUow(ctx).UiStrErr(s) }

func s3ToolsSession(ctx context.Context) error {
	opts := s3ToolsOpts(ctx)
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

func s3ToolsList(ctx context.Context) error {
	opts := s3ToolsOpts(ctx)
	args := s3ToolsCtx(ctx).args
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
	s3ToolsOut(ctx, fmt.Sprintf("%s\n", ls))
	return nil
}

func s3ToolsClone(ctx context.Context) error {
	var (
		is3o, is3t cabridss.IS3Session
	)
	opts := s3ToolsOpts(ctx)
	for i := 0; i < 2; i++ {
		oc := GetS3Config(opts.BaseOptions, i)
		is3 := cabridss.NewS3Session(oc, nil)
		if err := is3.Initialize(); err != nil {
			return err
		}
		if i == 0 {
			is3o = is3
		} else {
			is3t = is3
		}
	}
	for _, pfx := range []string{"meta-", "content-", ""} {
		rs, err := is3t.List(pfx)
		if err != nil {
			return err
		}
		if len(rs) > 0 {
			return fmt.Errorf("target object storage system must be empty (%s...)", rs[0])
		}
	}
	_ = is3o
	return nil
}

func s3ToolsRun(ctx context.Context) error {
	opts := s3ToolsOpts(ctx)
	err := fmt.Errorf("at least one operation option must be given with the s3Tools command")
	if opts.S3Session {
		err = s3ToolsSession(ctx)
	}
	if opts.S3List {
		err = s3ToolsList(ctx)
	}
	if opts.S3Clone {
		err = s3ToolsClone(ctx)
	}
	if err != nil {
		if errors.Is(err, cabridss.ErrPasswordRequired) {
			return cabridss.ErrPasswordRequired
		}
		return err
	}
	return nil
}
