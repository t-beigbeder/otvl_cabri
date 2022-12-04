package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrisync"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"strings"
)

type SyncOptions struct {
	BaseOptions
	Recursive    bool
	DryRun       bool
	BiDir        bool
	KeepContent  bool
	NoCh         bool
	NoACL        bool
	Verbose      bool
	VerboseLevel int
	LeftTime     string
	RightTime    string
}

type SyncVars struct {
	baseVars
}

func SyncStartup(cr *joule.CLIRunner[SyncOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[SyncOptions, *SyncVars](ctx)).vars = &SyncVars{baseVars: baseVars{uow: work}}
			return nil, synchronize(ctx, cr.Args[0], cr.Args[1])
		})
	return nil
}

func SyncShutdown(cr *joule.CLIRunner[SyncOptions]) error {
	return cr.GetUow("command").GetError()
}

func syncCtx(ctx context.Context) *uiContext[SyncOptions, *SyncVars] {
	return uiCtxFrom[SyncOptions, *SyncVars](ctx)
}

func syncVars(ctx context.Context) *SyncVars { return (*syncCtx(ctx)).vars }

func syncOpts(ctx context.Context) SyncOptions { return (*syncCtx(ctx)).opts }

func syncUow(ctx context.Context) joule.UnitOfWork { return getUnitOfWork[SyncOptions, *SyncVars](ctx) }

func syncOut(ctx context.Context, s string) { syncUow(ctx).UiStrOut(s) }

func syncErr(ctx context.Context, s string) { syncUow(ctx).UiStrErr(s) }

func str2dss(ctx context.Context, dssPath string, isRight bool, obsIx *int) (dss cabridss.Dss, path string, err error) {
	var (
		mp       string
		lasttime int64
		slt      string
	)
	if mp, err = MasterPassword(syncUow(ctx), syncOpts(ctx).BaseOptions, 0); err != nil {
		return
	}
	dssType, root, path, _ := CheckDssPath(dssPath)
	if isRight {
		slt = syncOpts(ctx).RightTime
	} else {
		slt = syncOpts(ctx).LeftTime
	}
	if slt != "" {
		lasttime, _ = CheckTimeStamp(slt)
	}
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return nil, "", err
		}
	} else if dssType == "olf" {
		oc, err := GetOlfConfig(syncOpts(ctx).BaseOptions, *obsIx, root, mp)
		if err != nil {
			return nil, "", err
		}
		if dss, err = cabridss.NewOlfDss(oc, lasttime, nil); err != nil {
			return nil, "", err
		}
		*obsIx += 1
	} else if dssType == "xolf" {
		// FIXME acl
		if dss, err = NewXolfDss(syncOpts(ctx).BaseOptions, *obsIx, lasttime, root, mp, nil); err != nil {
			return nil, "", err
		}
		*obsIx += 1
	} else if dssType == "obs" {
		oc, err := GetObsConfig(syncOpts(ctx).BaseOptions, *obsIx, root, mp)
		if err != nil {
			return nil, "", err
		}
		if dss, err = cabridss.NewObsDss(oc, lasttime, nil); err != nil {
			*obsIx += 1
			return nil, "", err
		}
		*obsIx += 1
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(syncOpts(ctx).BaseOptions, *obsIx, root, mp)
		if err != nil {
			return nil, "", err
		}
		if dss, err = cabridss.NewObsDss(sc, lasttime, nil); err != nil {
			return nil, "", err
		}
	} else if dssType == "webapi+http" {
		frags := strings.Split(root[2:], "/")
		wc, err := GetWebConfig(syncOpts(ctx).BaseOptions, 0, frags[0], frags[1], mp)
		if err != nil {
			return nil, "", err
		}
		if dss, err = cabridss.NewWebDss(wc, 0, nil); err != nil {
			return nil, "", err
		}
	} else {
		return nil, "", fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	return
}

func synchronize(ctx context.Context, ldssPath, rdssPath string) error {
	var err error
	opts := syncOpts(ctx)
	obsIx := 0
	ldss, lpath, err := str2dss(ctx, ldssPath, false, &obsIx)
	if err != nil {
		return err
	}
	rdss, rpath, err := str2dss(ctx, rdssPath, true, &obsIx)
	if err != nil {
		return err
	}
	var beVerbose cabrisync.BeVerboseFunc
	if opts.VerboseLevel >= 2 {
		beVerbose = func(level int, line string) {
			if level > opts.VerboseLevel {
				return
			}
			syncErr(ctx, line+"\n")
		}
	}
	sOpts := cabrisync.SyncOptions{
		InDepth:     opts.Recursive,
		Evaluate:    opts.DryRun,
		BiDir:       opts.BiDir,
		KeepContent: opts.KeepContent,
		NoCh:        opts.NoCh,
		NoACL:       opts.NoACL,
		BeVerbose:   beVerbose,
	}
	iOutputs := plumber.LaunchAndWait(ctx,
		[]string{"Synchronized"},
		[]plumber.Launchable{cabrisync.PlizedSynchronize},
		[]interface{}{cabrisync.SyncArgs{LDss: ldss, LPath: lpath, RDss: rdss, RPath: rpath, SOpts: sOpts}},
	)
	outputs := plumber.Retype[cabrisync.SyncReport](iOutputs)
	sr := outputs[0]

	if errClose := ldss.Close(); errClose != nil {
		if err == nil {
			err = errClose
		}
	}
	if errClose := rdss.Close(); errClose != nil {
		if err == nil {
			err = errClose
		}
	}

	if sr.GErr != nil {
		return sr.GErr
	}
	stats := sr.GetStats()
	if opts.DryRun || opts.Verbose {
		sr.SortByPath().TextOutput(syncUow(ctx).UiErrWriter())
		syncErr(ctx, fmt.Sprintf(
			"created: %d, updated %d, removed %d, kept %d, touched %d, error(s) %d\n",
			stats.CreNum, stats.UpdNum, stats.RmvNum, stats.KeptNum, stats.MUpNum, stats.ErrNum))
	}
	if stats.ErrNum > 0 {
		return fmt.Errorf("some errors encountered")
	}
	return nil
}
