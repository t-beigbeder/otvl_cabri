package cabriui

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"strings"
	"sync"
)

type S3ToolsOptions struct {
	BaseOptions
	S3Session bool
	S3List    bool
	S3Purge   bool
	S3Clone   bool
	S3AsOlf   bool
	S3Put     bool
	S3Get     bool
	S3Rename  bool
	S3Delete  bool
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

func s3ToolsPurge(ctx context.Context) error {
	opts := s3ToolsOpts(ctx)
	oc, err := GetObsConfig(opts.BaseOptions, 0, "", "")
	if err != nil {
		return err
	}
	is3 := cabridss.NewS3Session(oc, nil)
	if err = is3.Initialize(); err != nil {
		return err
	}
	return is3.DeleteAll("")
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
	var red plumber.Reducer = nil
	if opts.RedLimit != 0 {
		red = plumber.NewReducer(opts.RedLimit, 0)
	}
	es, err := is3o.List("")
	if err != nil {
		return err
	}
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	var errs []error
	recErr := func(iErr error) {
		mx.Lock()
		defer mx.Unlock()
		errs = append(errs, iErr)
	}
	cloneEntry := func(pe string) error {
		rc, err := is3o.Download(pe)
		if err != nil {
			return err
		}
		defer rc.Close()
		if err = is3t.Upload(pe, rc); err != nil {
			return err
		}
		return nil
	}
	for _, ent := range es {
		wg.Add(1)
		go func(pe string) {
			iErr := func() error {
				defer wg.Done()
				if red == nil {
					return cloneEntry(pe)
				} else {
					return red.Launch(fmt.Sprintf("cloneEntry-%s", pe), func() error {
						return cloneEntry(pe)
					})
				}
			}()
			if iErr != nil {
				recErr(iErr)
			}
		}(ent)
	}
	wg.Wait()
	if len(errs) > 0 {
		for _, err := range errs {
			s3ToolsErr(ctx, err.Error()+"\n")
		}
		return fmt.Errorf("some errors occured")
	}
	return nil
}

func s3ToPath(pe string, size string) string {
	pes := strings.Split(pe, "-")
	s := pes[1]
	if size == "s" {
		s = fmt.Sprintf("%s/%s", s[0:2], s[2:])
	} else if size == "m" {
		s = fmt.Sprintf("%s/%s", s[0:3], s[3:])
	} else {
		s = fmt.Sprintf("%s/%s/%s", s[0:3], s[3:6], s[6:])
	}
	return pes[0] + "/" + s
}

func s3ToOlf(ctx context.Context, red plumber.Reducer, is3 cabridss.IS3Session, olfPath string, size string) error {
	var (
		names []string
		cns   []string
		err   error
	)
	names, err = is3.List("meta-")
	if err != nil {
		return err
	}
	cns, err = is3.List("content-")
	if err != nil {
		return err
	}
	names = append(names, cns...)
	appFs := afero.NewOsFs()
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	var errs []error
	recErr := func(iErr error) {
		mx.Lock()
		defer mx.Unlock()
		errs = append(errs, iErr)
	}
	entry2olf := func(pe string) error {
		pep := olfPath + "/" + s3ToPath(pe, size)
		_, err := appFs.Stat(pep)
		if err == nil {
			return nil
		}
		rc, err := is3.Download(pe)
		if err != nil {
			return err
		}
		defer rc.Close()
		if err = appFs.MkdirAll(ufpath.Dir(pep), 0o777); err != nil {
			return err
		}
		oe, err := appFs.Create(pep)
		if err != nil {
			return err
		}
		_, err = io.Copy(oe, rc)
		if err != nil {
			oe.Close()
			return err
		}
		return oe.Close()
	}
	for _, ent := range names {
		wg.Add(1)
		go func(pe string) {
			iErr := func() error {
				defer wg.Done()
				if red == nil {
					return entry2olf(pe)
				} else {
					return red.Launch(fmt.Sprintf("entry2olf-%s", pe), func() error {
						return entry2olf(pe)
					})
				}
			}()
			if iErr != nil {
				recErr(iErr)
			}
		}(ent)
	}
	wg.Wait()
	if len(errs) > 0 {
		for _, err := range errs {
			s3ToolsErr(ctx, err.Error()+"\n")
		}
		return fmt.Errorf("some errors occured")
	}
	return nil
}

func s3ToolsAsOlf(ctx context.Context) error {
	opts := s3ToolsOpts(ctx)
	args := s3ToolsCtx(ctx).args
	if len(args) != 1 {
		return fmt.Errorf("an olf DSS local path must be provided")
	}
	var (
		err error
		oc  cabridss.ObsConfig
		ure UiRunEnv
	)

	dssType, root, err := CheckDssSpec(args[0])
	if err != nil {
		return err
	}
	if dssType != "olf" && dssType != "xolf" {
		return fmt.Errorf("only olf and xolf DSS are supported")
	}
	if ure, err = GetUiRunEnv[S3ToolsOptions, *S3ToolsVars](ctx, dssType[0] == 'x', false); err != nil {
		return err
	}
	_, _ = root, ure
	bc := cabridss.DssBaseConfig{LocalPath: root, ConfigPassword: ure.MasterPassword}
	olfc := cabridss.OlfConfig{}
	err = cabridss.LoadDssConfig(bc, &olfc)
	if err != nil {
		return err
	}

	oc, err = GetObsConfig(opts.BaseOptions, 0, "", "")
	if err != nil {
		return err
	}
	is3 := cabridss.NewS3Session(oc, nil)
	if err = is3.Initialize(); err != nil {
		return err
	}
	var red plumber.Reducer = nil
	if opts.RedLimit != 0 {
		red = plumber.NewReducer(opts.RedLimit, 0)
	}
	return s3ToOlf(ctx, red, is3, root, olfc.Size)
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
	if opts.S3Purge {
		err = s3ToolsPurge(ctx)
	}
	if opts.S3Clone {
		err = s3ToolsClone(ctx)
	}
	if opts.S3AsOlf {
		err = s3ToolsAsOlf(ctx)
	}
	if err != nil {
		if errors.Is(err, cabridss.ErrPasswordRequired) {
			return cabridss.ErrPasswordRequired
		}
		return err
	}
	return nil
}
