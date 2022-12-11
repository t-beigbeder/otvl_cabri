package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"strings"
	"time"
)

type DSSMkOptions struct {
	BaseOptions
	Size string
}

type DSSMkVars struct {
	baseVars
}

func DSSMkStartup(cr *joule.CLIRunner[DSSMkOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSMkOptions, *DSSMkVars](ctx)).vars = &DSSMkVars{baseVars: baseVars{uow: work}}
			return nil, dssMkRun(ctx)
		})
	return nil
}

func DSSMkShutdown(cr *joule.CLIRunner[DSSMkOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssMkCtx(ctx context.Context) *uiContext[DSSMkOptions, *DSSMkVars] {
	return uiCtxFrom[DSSMkOptions, *DSSMkVars](ctx)
}

func dssMkOpts(ctx context.Context) DSSMkOptions { return (*dssMkCtx(ctx)).opts }

func dssMkUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSMkOptions, *DSSMkVars](ctx)
}

func dssMkRun(ctx context.Context) error {
	opts := dssMkOpts(ctx)
	args := dssMkCtx(ctx).args
	dssType, root, _ := CheckDssSpec(args[0])
	var (
		dss cabridss.Dss
		err error
		mp  string
	)
	if mp, err = MasterPassword(dssMkUow(ctx), opts.BaseOptions, 0); err != nil {
		return err
	}
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return err
		}
	} else if dssType == "olf" || dssType == "xolf" {
		encrypted := dssType[0] == 'x'
		if dss, err = cabridss.CreateOlfDss(cabridss.OlfConfig{
			DssBaseConfig: cabridss.DssBaseConfig{LocalPath: root, Encrypted: encrypted, ConfigPassword: mp},
			Root:          root, Size: opts.Size}); err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root, mp)
		if err != nil {
			return err
		}
		if dss, err = cabridss.CreateObsDss(oc); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root, mp)
		if err != nil {
			return err
		}
		if dss, err = cabridss.CreateObsDss(sc); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	if err = dss.Close(); err != nil {
		return err
	}
	return nil
}

type DSSMknsOptions struct {
	BaseOptions
	Children []string
}

type DSSMknsVars struct {
	baseVars
}

func DSSMknsStartup(cr *joule.CLIRunner[DSSMknsOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSMknsOptions, *DSSMknsVars](ctx)).vars = &DSSMknsVars{baseVars: baseVars{uow: work}}
			return nil, dssMknsRun(ctx)
		})
	return nil
}

func DSSMknsShutdown(cr *joule.CLIRunner[DSSMknsOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssMknsCtx(ctx context.Context) *uiContext[DSSMknsOptions, *DSSMknsVars] {
	return uiCtxFrom[DSSMknsOptions, *DSSMknsVars](ctx)
}

func dssMknsOpts(ctx context.Context) DSSMknsOptions { return (*dssMknsCtx(ctx)).opts }

func dssMknsUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSMknsOptions, *DSSMknsVars](ctx)
}

func dssMknsRun(ctx context.Context) error {
	opts := dssMknsOpts(ctx)
	args := dssMknsCtx(ctx).args
	dssType, root, npath, _ := CheckDssPath(args[0])
	var (
		dss cabridss.Dss
		err error
		ure UiRunEnv
	)
	if ure, err = GetUiRunEnv[DSSMknsOptions, *DSSMknsVars](ctx, dssType[0] == 'x'); err != nil {
		return err
	}
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return err
		}
	} else if dssType == "olf" {
		oc, err := GetOlfConfig(opts.BaseOptions, 0, root, ure.MasterPassword)
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewOlfDss(oc, 0, ure.Users); err != nil {
			return err
		}
	} else if dssType == "xolf" {
		dss, err = NewXolfDss(opts.BaseOptions, 0, 0, root, ure.MasterPassword, ure.Users)
		if err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root, ure.MasterPassword)
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewObsDss(oc, 0, ure.Users); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root, ure.MasterPassword)
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewObsDss(sc, 0, ure.Users); err != nil {
			return err
		}
	} else if dssType == "webapi+http" {
		frags := strings.Split(root[2:], "/")
		wc, err := GetWebConfig(opts.BaseOptions, 0, frags[0], frags[1], ure.MasterPassword)
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewWebDss(wc, 0, ure.Users); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	acl, err := ure.ACLOrDefault()
	if err != nil {
		return err
	}
	if err = dss.Mkns(npath, time.Now().Unix(), opts.Children, acl); err != nil {
		return err
	}
	if err = dss.Close(); err != nil {
		return err
	}
	return nil
}

type DSSUnlockOptions struct {
	BaseOptions
	RepairIndex    bool
	RepairReadOnly bool
}

type DSSUnlockVars struct {
	baseVars
}

func DSSUnlockStartup(cr *joule.CLIRunner[DSSUnlockOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSUnlockOptions, *DSSUnlockVars](ctx)).vars = &DSSUnlockVars{baseVars: baseVars{uow: work}}
			return nil, dssMknsRun(ctx)
		})
	return nil
}

func DSSUnlockShutdown(cr *joule.CLIRunner[DSSUnlockOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssUnlockCtx(ctx context.Context) *uiContext[DSSUnlockOptions, *DSSUnlockVars] {
	return uiCtxFrom[DSSUnlockOptions, *DSSUnlockVars](ctx)
}

func dssUnlockOpts(ctx context.Context) DSSUnlockOptions { return (*dssUnlockCtx(ctx)).opts }

func dssUnlockUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSUnlockOptions, *DSSUnlockVars](ctx)
}

func dssUnlockOut(ctx context.Context, s string) { dssUnlockUow(ctx).UiStrOut(s) }

func dssUnlockRun(ctx context.Context) error {
	opts := dssUnlockOpts(ctx)
	args := dssUnlockCtx(ctx).args
	dssType, root, _ := CheckDssSpec(args[0])
	var (
		dss cabridss.HDss
		err error
		mp  string
	)
	if mp, err = MasterPassword(dssMknsUow(ctx), opts.BaseOptions, 0); err != nil {
		return err
	}
	if dssType == "olf" {
		oc, err := GetOlfConfig(opts.BaseOptions, 0, root, mp)
		if err != nil {
			return err
		}
		oc.DssBaseConfig.Unlock = true
		if dss, err = cabridss.NewOlfDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root, mp)
		if err != nil {
			return err
		}
		oc.DssBaseConfig.Unlock = true
		if dss, err = cabridss.NewObsDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root, mp)
		if err != nil {
			return err
		}
		sc.DssBaseConfig.Unlock = true
		if dss, err = cabridss.NewObsDss(sc, 0, nil); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	if dss.GetIndex() != nil && dss.GetIndex().IsPersistent() && opts.RepairIndex {
		ds, err := dss.GetIndex().Repair(opts.RepairReadOnly)
		if err != nil {
			return err
		}
		for _, d := range ds {
			dssUnlockOut(ctx, d)
		}
	}
	if err = dss.Close(); err != nil {
		return err
	}
	return nil

}

type DSSCleanOptions struct {
	BaseOptions
}

type DSSCleanVars struct {
	baseVars
}

func DSSCleanStartup(cr *joule.CLIRunner[DSSCleanOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSCleanOptions, *DSSCleanVars](ctx)).vars = &DSSCleanVars{baseVars: baseVars{uow: work}}
			return nil, dssMknsRun(ctx)
		})
	return nil
}

func DSSCleanShutdown(cr *joule.CLIRunner[DSSCleanOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssCleanCtx(ctx context.Context) *uiContext[DSSCleanOptions, *DSSCleanVars] {
	return uiCtxFrom[DSSCleanOptions, *DSSCleanVars](ctx)
}

func dssCleanOpts(ctx context.Context) DSSCleanOptions { return (*dssCleanCtx(ctx)).opts }

func dssCleanUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSCleanOptions, *DSSCleanVars](ctx)
}

func dssCleanOut(ctx context.Context, s string) { dssCleanUow(ctx).UiStrOut(s) }

func dssCleanRun(ctx context.Context) error {
	opts := dssCleanOpts(ctx).BaseOptions
	args := dssCleanCtx(ctx).args
	dssType, root, _ := CheckDssSpec(args[0])
	var (
		config cabridss.ObsConfig
		err    error
		mp     string
	)
	if mp, err = MasterPassword(dssMknsUow(ctx), opts, 0); err != nil {
		return err
	}
	if dssType == "obs" {
		config, err = GetObsConfig(opts, 0, root, mp)
		if err != nil {
			return err
		}
	} else if dssType == "smf" {
		config, err = GetSmfConfig(opts, 0, root, mp)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	return cabridss.CleanObsDss(config)
}
