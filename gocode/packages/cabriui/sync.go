package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrisync"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"os"
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
	MapACL       []string
	LeftUsers    []string
	LeftACL      []string
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

func str2dss(ctx context.Context, dssPath string, isRight bool, obsIx *int) (cabridss.Dss, string, UiRunEnv, error) {
	var (
		dss      cabridss.Dss
		path     string
		ure      UiRunEnv
		err      error
		lasttime int64
		slt      string
	)
	dssType, root, path, _ := CheckDssPath(dssPath)
	// will setup users and ACL for right-side DSS
	if ure, err = GetUiRunEnv[SyncOptions, *SyncVars](ctx, dssType[0] == 'x'); err != nil {
		return nil, "", ure, err
	}
	if isRight {
		slt = syncOpts(ctx).RightTime
	} else {
		// fix users and ACL for left-side DSS
		if ure.UiACL, err = CheckUiACL(syncOpts(ctx).LeftACL); err != nil {
			return nil, "", ure, err
		}
		ure.Users = syncOpts(ctx).LeftUsers
		if _, err = ure.ACLOrDefault(); err != nil {
			return nil, "", ure, err
		}
		slt = syncOpts(ctx).LeftTime
	}
	if slt != "" {
		lasttime, _ = CheckTimeStamp(slt)
	}
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return nil, "", ure, err
		}
		ure.DefaultSyncUser = fmt.Sprintf("x-uid:%d", os.Getuid())
	} else if dssType == "olf" {
		oc, err := GetOlfConfig(syncOpts(ctx).BaseOptions, *obsIx, root, ure.MasterPassword)
		if err != nil {
			return nil, "", ure, err
		}
		if dss, err = cabridss.NewOlfDss(oc, lasttime, ure.Users); err != nil {
			return nil, "", ure, err
		}
		*obsIx += 1
	} else if dssType == "xolf" {
		if dss, err = NewXolfDss(syncOpts(ctx).BaseOptions, *obsIx, lasttime, root, ure.MasterPassword, ure.Users); err != nil {
			return nil, "", ure, err
		}
		*obsIx += 1
	} else if dssType == "obs" {
		oc, err := GetObsConfig(syncOpts(ctx).BaseOptions, *obsIx, root, ure.MasterPassword)
		if err != nil {
			return nil, "", ure, err
		}
		if dss, err = cabridss.NewObsDss(oc, lasttime, ure.Users); err != nil {
			*obsIx += 1
			return nil, "", ure, err
		}
		*obsIx += 1
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(syncOpts(ctx).BaseOptions, *obsIx, root, ure.MasterPassword)
		if err != nil {
			return nil, "", ure, err
		}
		if dss, err = cabridss.NewObsDss(sc, lasttime, ure.Users); err != nil {
			return nil, "", ure, err
		}
	} else if dssType == "webapi+http" {
		frags := strings.Split(root[2:], "/")
		wc, err := GetWebConfig(syncOpts(ctx).BaseOptions, 0, frags[0], frags[1], ure.MasterPassword)
		if err != nil {
			return nil, "", ure, err
		}
		if dss, err = cabridss.NewWebDss(wc, 0, ure.Users); err != nil {
			return nil, "", ure, err
		}
	} else {
		err = fmt.Errorf("DSS type %s is not (yet) supported", dssType)
		return nil, "", ure, err
	}
	return dss, path, ure, nil
}

func uiMapACL(opts SyncOptions, lure, rure UiRunEnv) (map[string]string, error) {
	macl := map[string]string{}
	for _, uim := range opts.MapACL {
		uimes := strings.Split(uim, ":")
		if len(uimes) != 2 {
			return nil, fmt.Errorf("ACL user mapping %s has not the form <left-user:right-user>", uim)
		}
		lu, ru := uimes[0], uimes[1]
		if lu == "" {
			lu = lure.DefaultSyncUser
		}
		if ru == "" {
			ru = rure.DefaultSyncUser
		}
		if lure.Encrypted {
			lup := lure.UserConfig.GetIdentity(lu).PKey
			if lup == "" {
				return nil, fmt.Errorf("no public key for left identity %s in ACL user mapping %s", lu, uim)
			}
			lu = lup
		}
		if rure.Encrypted {
			rup := rure.UserConfig.GetIdentity(ru).PKey
			if rup == "" {
				return nil, fmt.Errorf("no public key for left identity %s in ACL user mapping %s", ru, uim)
			}
			ru = rup
		}
		macl[lu] = ru
	}
	return macl, nil
}

func synchronize(ctx context.Context, ldssPath, rdssPath string) error {
	var (
		err error
	)
	opts := syncOpts(ctx)
	obsIx := 0
	ldss, lpath, lure, err := str2dss(ctx, ldssPath, false, &obsIx)
	if err != nil {
		return err
	}
	rdss, rpath, rure, err := str2dss(ctx, rdssPath, true, &obsIx)
	if err != nil {
		return err
	}
	macl, err := uiMapACL(opts, lure, rure)
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
		MapACL:      macl,
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
