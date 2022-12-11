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
		ure UiRunEnv
	)
	vars := lsnsVars(ctx)
	vars.dssType, vars.root, vars.npath, _ = CheckDssPath(dssPath)
	if ure, err = GetUiRunEnv[LsnsOptions, *LsnsVars](ctx, vars.dssType[0] == 'x'); err != nil {
		return err
	}
	var lasttime int64
	slt := lsnsOpts(ctx).LastTime
	if slt != "" {
		lasttime, _ = CheckTimeStamp(slt)
	}
	if vars.dssType == "fsy" {
		if vars.dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, vars.root); err != nil {
			return err
		}
	} else if vars.dssType == "olf" {
		oc, err := GetOlfConfig(lsnsOpts(ctx).BaseOptions, 0, vars.root, ure.MasterPassword)
		if err != nil {
			return err
		}
		if vars.dss, err = cabridss.NewOlfDss(oc, lasttime, ure.Users); err != nil {
			return err
		}
	} else if vars.dssType == "xolf" {
		vars.dss, err = NewXolfDss(lsnsOpts(ctx).BaseOptions, 0, lasttime, vars.root, ure.MasterPassword, ure.Users)
		if err != nil {
			return err
		}
	} else if vars.dssType == "obs" {
		oc, err := GetObsConfig(lsnsOpts(ctx).BaseOptions, 0, vars.root, ure.MasterPassword)
		if err != nil {
			return err
		}
		if vars.dss, err = cabridss.NewObsDss(oc, lasttime, ure.Users); err != nil {
			return err
		}
	} else if vars.dssType == "smf" {
		sc, err := GetSmfConfig(lsnsOpts(ctx).BaseOptions, 0, vars.root, ure.MasterPassword)
		if err != nil {
			return err
		}
		if vars.dss, err = cabridss.NewObsDss(sc, lasttime, ure.Users); err != nil {
			return err
		}
	} else if vars.dssType == "webapi+http" {
		frags := strings.Split(vars.root[2:], "/")
		wc, err := GetWebConfig(lsnsOpts(ctx).BaseOptions, 0, frags[0], frags[1], ure.MasterPassword)
		if err != nil {
			return err
		}
		if vars.dss, err = cabridss.NewWebDss(wc, 0, ure.Users); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", vars.dssType)
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
	getCh := lsnsOpts(ctx).Checksum
	t := time.Unix(meta.GetMtime(), 0).Format("2006-01-02 15:04:05")
	if !getCh {
		lsnsOut(ctx, fmt.Sprintf("%12d %s %s\n", meta.GetSize(), t, meta.GetPath()))
	} else {
		lsnsOut(ctx, fmt.Sprintf("%12d %s %s %s\n", meta.GetSize(), t, meta.GetCh(), meta.GetPath()))
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
