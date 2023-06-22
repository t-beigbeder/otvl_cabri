package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrisync"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"os"
	"runtime/debug"
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
	Summary      bool
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
	if ure, err = GetUiRunEnv[SyncOptions, *SyncVars](ctx, dssType[0] == 'x', !isRight); err != nil {
		return nil, "", ure, err
	}
	if isRight {
		slt = syncOpts(ctx).RightTime
	} else {
		// fix users and ACL for left-side DSS
		if ure.UiACL, err = CheckUiACL(syncOpts(ctx).LeftACL); err != nil {
			return nil, "", ure, err
		}
		ure.UiUsers = syncOpts(ctx).LeftUsers
		if _, err = ure.ACLOrDefault(); err != nil {
			return nil, "", ure, err
		}
		slt = syncOpts(ctx).LeftTime
	}
	if slt != "" {
		lasttime, _ = CheckTimeStamp(slt)
	}
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(
			cabridss.FsyConfig{
				DssBaseConfig: cabridss.DssBaseConfig{ReducerLimit: syncOpts(ctx).RedLimit},
			},
			root); err != nil {
			return nil, "", ure, err
		}
		ure.DefaultSyncUser = fmt.Sprintf("x-uid:%d", os.Getuid())
	} else {
		dx := 0
		if isRight {
			dx = 1
		}
		nhArgs := NewHDssArgs{DssIx: dx, ObsIx: *obsIx, Lasttime: lasttime}
		dss, err = NewHDss[SyncOptions, *SyncVars](ctx, nil, nhArgs)
		*obsIx += 1
		if err != nil {
			return nil, "", ure, err
		}
	}
	return dss, path, ure, nil
}

func uiMapACL(opts SyncOptions, lure, rure UiRunEnv) (lmacl, rmacl map[string][]cabridss.ACLEntry, err error) {
	lmacl = map[string][]cabridss.ACLEntry{}
	rmacl = map[string][]cabridss.ACLEntry{}
	err = nil
	for _, uim := range opts.MapACL {
		uimes := strings.Split(uim, ":")
		if len(uimes) != 2 {
			err = fmt.Errorf("ACL user mapping %s has not the form <left-user:right-user>", uim)
			return
		}
		lu, ru := uimes[0], uimes[1]
		if lu == "" {
			lu = lure.DefaultSyncUser
		}
		if ru == "" {
			ru = rure.DefaultSyncUser
		}
		lua, rua := lu, ru
		if lure.Encrypted {
			lup := lure.UserConfig.GetIdentity(lu).PKey
			if lup == "" {
				err = fmt.Errorf("no public key for left identity %s in ACL user mapping %s", lu, uim)
				return
			}
			lu = lup
		}
		if rure.Encrypted {
			rup := rure.UserConfig.GetIdentity(ru).PKey
			if rup == "" {
				err = fmt.Errorf("no public key for right identity %s in ACL user mapping %s", ru, uim)
				return
			}
			ru = rup
		}
		dr := cabridss.Rights{
			Read:    true,
			Write:   true,
			Execute: true,
		}
		lmacl[lu] = append(lmacl[lu], cabridss.ACLEntry{User: ru, Rights: cabridss.GetUserRights(rure.UiACL, rua, dr)})
		rmacl[ru] = append(rmacl[ru], cabridss.ACLEntry{User: lu, Rights: cabridss.GetUserRights(lure.UiACL, lua, dr)})
	}
	return
}

func synchronize(ctx context.Context, ldssPath, rdssPath string) error {
	var (
		err error
	)
	opts := syncOpts(ctx)
	if opts.MapACL == nil {
		opts.MapACL = []string{":"}
	}
	obsIx := 0
	ldss, lpath, lure, err := str2dss(ctx, ldssPath, false, &obsIx)
	if err != nil {
		return err
	}
	rdss, rpath, rure, err := str2dss(ctx, rdssPath, true, &obsIx)
	if err != nil {
		ldss.Close()
		return err
	}
	lmacl, rmacl, err := uiMapACL(opts, lure, rure)
	if err != nil {
		ldss.Close()
		rdss.Close()
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
		LeftMapACL:  lmacl,
		RightMapACL: rmacl,
		BeVerbose:   beVerbose,
	}
	if opts.MaxThread != 0 {
		debug.SetMaxThreads(opts.MaxThread)
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
		ssr := sr.SortByPath()
		wrt := syncUow(ctx).UiOutWriter()
		if opts.Summary {
			ssr.SummaryOutput(wrt)
		} else {
			ssr.TextOutput(wrt)
		}
		syncOut(ctx, fmt.Sprintf(
			"created: %d, updated %d, removed %d, kept %d, touched %d, error(s) %d\n",
			stats.CreNum, stats.UpdNum, stats.RmvNum, stats.KeptNum, stats.MUpNum, stats.ErrNum))
	}
	if stats.ErrNum > 0 {
		return fmt.Errorf("some errors encountered")
	}
	return nil
}
