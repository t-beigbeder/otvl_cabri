package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"sort"
	"strings"
	"time"
)

type LsnsOptions struct {
	BaseOptions
	Recursive bool
	Sorted    bool
	Time      bool
	Long      bool
	Checksum  bool
	Reverse   bool
	LastTime  string
}

func (los LsnsOptions) getLastTime() (lastTime int64) {
	if los.LastTime != "" {
		lastTime, _ = CheckTimeStamp(los.LastTime)
	}
	return
}

type LsnsVars struct {
	baseVars
	dssType string
	dss     cabridss.Dss
	root    string
	npath   string
}

func LsnsStartup(cr *joule.CLIRunner[LsnsOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[LsnsOptions, *LsnsVars](ctx)).vars = &LsnsVars{baseVars: baseVars{uow: work}}
			return nil, lsns(ctx, cr.Args[0])
		})
	return nil
}

func LsnsShutdown(cr *joule.CLIRunner[LsnsOptions]) error {
	return cr.GetUow("command").GetError()
}

func lsnsCtx(ctx context.Context) *uiContext[LsnsOptions, *LsnsVars] {
	return uiCtxFrom[LsnsOptions, *LsnsVars](ctx)
}

func lsnsVars(ctx context.Context) *LsnsVars { return (*lsnsCtx(ctx)).vars }

func lsnsOpts(ctx context.Context) LsnsOptions { return (*lsnsCtx(ctx)).opts }

func lsnsUow(ctx context.Context) joule.UnitOfWork { return getUnitOfWork[LsnsOptions, *LsnsVars](ctx) }

func lsnsOut(ctx context.Context, s string) { lsnsUow(ctx).UiStrOut(s) }

func lsnsErr(ctx context.Context, s string) { lsnsUow(ctx).UiStrErr(s) }

func lsns(ctx context.Context, dssPath string) error {
	var (
		err error
	)
	vars := lsnsVars(ctx)
	vars.dssType, vars.root, vars.npath, _ = CheckDssPath(dssPath)
	if _, err = GetUiRunEnv[LsnsOptions, *LsnsVars](ctx, vars.dssType[0] == 'x', false); err != nil {
		return err
	}
	if vars.dssType == "fsy" {
		if vars.dss, err = cabridss.NewFsyDss(
			cabridss.FsyConfig{
				DssBaseConfig: cabridss.DssBaseConfig{ReducerLimit: lsnsOpts(ctx).RedLimit},
			},
			vars.root); err != nil {
			return err
		}
	} else if strings.HasPrefix(vars.dssType, "wfsapi+") {
		if vars.dss, err = NewWfsDss[LsnsOptions, *LsnsVars](ctx, nil,
			NewHDssArgs{Lasttime: lsnsOpts(ctx).getLastTime()}); err != nil {
			return err
		}
	} else if vars.dss, err = NewHDss[LsnsOptions, *LsnsVars](ctx, nil,
		NewHDssArgs{Lasttime: lsnsOpts(ctx).getLastTime()}); err != nil {
		return err
	}

	sorted := isLsnsSorted(ctx)
	metas, err := lsnsRecurs(ctx, cabridss.AppendSlashIf(vars.npath))
	if errClose := vars.dss.Close(); errClose != nil {
		if err == nil {
			err = errClose
		}
	}
	if sorted {
		time, reverse := lsnsOpts(ctx).Time, lsnsOpts(ctx).Reverse
		sort.Slice(metas, func(i, j int) bool {
			if reverse {
				i, j = j, i
			}
			if metas[i].IMeta == nil || metas[j].IMeta == nil {
				return i < j
			}
			if time {
				return metas[i].IMeta.GetMtime() < metas[j].IMeta.GetMtime()
			} else {
				return metas[i].IMeta.GetPath() < metas[j].IMeta.GetPath()
			}
		})
	}

	errCount := 0
	for _, meta := range metas {
		if meta.error != nil {
			errCount += 1
		} else if sorted {
			outMeta(ctx, meta.IMeta)
		}
	}
	if err != nil {
		lsnsErr(ctx, fmt.Sprintf("%v\n", err))
	}
	if err == nil && errCount != 0 {
		err = fmt.Errorf("some errors encountered")
	}

	return err
}

type metaOrErr struct {
	cabridss.IMeta
	error
}

func isLsnsSorted(ctx context.Context) bool { return lsnsOpts(ctx).Sorted || lsnsOpts(ctx).Time }

func plizedGetMetas(ctx context.Context, iNpaths interface{}) (iOutput interface{}) {
	sorted := isLsnsSorted(ctx)
	getCh := lsnsOpts(ctx).Checksum
	npaths := plumber.Retype[string](plumber.Untype[string](iNpaths.([]string)))
	var metas []metaOrErr
	for _, meta := range plumber.Parallelize[string, metaOrErr](
		ctx, "",
		func(ctx context.Context, npath string) metaOrErr {
			meta, err := lsnsVars(ctx).dss.GetMeta(npath, getCh)
			if err != nil {
				lsnsErr(ctx, fmt.Sprintf("%v\n", err))
				return metaOrErr{error: err}
			}
			if !sorted {
				outMeta(ctx, meta)
			}
			return metaOrErr{IMeta: meta}
		},
		npaths...) {
		metas = append(metas, meta)
	}
	iOutput = metas
	return
}

func plizedLsnsMetas(ctx context.Context, iNpaths interface{}) (iOutput interface{}) {
	npaths := plumber.Retype[string](plumber.Untype[string](iNpaths.([]string)))
	var metas []metaOrErr
	for _, subMetas := range plumber.Parallelize[string, []metaOrErr](
		ctx, "",
		func(ctx context.Context, npath string) []metaOrErr {
			metas, _ := lsnsRecurs(ctx, npath)
			return metas
		},
		npaths...) {
		for _, meta := range subMetas {
			metas = append(metas, meta)
		}
	}
	iOutput = metas
	return

}

func outMeta(ctx context.Context, meta cabridss.IMeta) {
	t := time.Unix(meta.GetMtime(), 0).Format("2006-01-02 15:04:05")
	ll := "\n"
	if lsnsOpts(ctx).Long {
		ll = fmt.Sprintf("\n            \t%v\n", meta.GetAcl())
	}
	if !lsnsOpts(ctx).Checksum {
		lsnsOut(ctx, fmt.Sprintf("%12d %s %s%s", meta.GetSize(), t, meta.GetPath(), ll))
	} else {
		lsnsOut(ctx, fmt.Sprintf("%12d %s %s %s%s", meta.GetSize(), t, meta.GetCh(), meta.GetPath(), ll))
	}
}

func lsnsRecurs(ctx context.Context, npath string) ([]metaOrErr, error) {
	var metas []metaOrErr
	chs, err := lsnsVars(ctx).dss.Lsns(cabridss.RemoveSlashIf(npath))
	if err != nil {
		return nil, err
	}
	var aChs []string
	var dChs []string
	for _, ch := range chs {
		aChs = append(aChs, fmt.Sprintf("%s%s", npath, ch))
		if ch[len(ch)-1] == '/' {
			dChs = append(dChs, fmt.Sprintf("%s%s", npath, ch))
		}
	}

	if !lsnsOpts(ctx).Recursive {
		iOutputs := plumber.LaunchAndWait(ctx,
			[]string{"GetMetas"},
			[]plumber.Launchable{plizedGetMetas},
			[]interface{}{aChs},
		)
		outputs := plumber.Retype[[]metaOrErr](iOutputs)
		metas1 := outputs[0]
		for _, meta := range metas1 {
			metas = append(metas, meta)
		}
	} else {
		iOutputs := plumber.LaunchAndWait(ctx,
			[]string{"GetMetas", "LsnsMetas"},
			[]plumber.Launchable{plizedGetMetas, plizedLsnsMetas},
			[]interface{}{aChs, dChs},
		)
		outputs := plumber.Retype[[]metaOrErr](iOutputs)
		metas1 := outputs[0]
		for _, meta := range metas1 {
			metas = append(metas, meta)
		}
		metas2 := outputs[1]
		for _, meta := range metas2 {
			metas = append(metas, meta)
		}
	}
	return metas, nil
}
