package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
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
	encrypted := dssType[0] == 'x'
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return err
		}
	} else if dssType == "olf" || dssType == "xolf" {
		oc, lerr := GetOlfConfig(opts.BaseOptions, 0, root, mp)
		if lerr != nil {
			return lerr
		}
		oc.Encrypted = encrypted
		if encrypted {
			if oc.XImpl == "" {
				oc.XImpl = "bdb"
				oc.GetIndex = cabridss.GetPIndex
			}
		}
		oc.Size = opts.Size
		if dss, err = cabridss.CreateOlfDss(oc); err != nil {
			return err
		}
	} else if dssType == "obs" || dssType == "xobs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root, mp)
		if err != nil {
			return err
		}
		oc.Encrypted = encrypted
		if dss, err = cabridss.CreateObsDss(oc); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root, mp)
		if err != nil {
			return err
		}
		sc.Encrypted = encrypted
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
	if ure, err = GetUiRunEnv[DSSMknsOptions, *DSSMknsVars](ctx, dssType[0] == 'x', false); err != nil {
		return err
	}
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return err
		}
	} else if dss, err = NewHDss[DSSMknsOptions, *DSSMknsVars](ctx, nil, NewHDssArgs{}); err != nil {
		return err
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
	LockForTest    bool
}

type DSSUnlockVars struct {
	baseVars
}

func DSSUnlockStartup(cr *joule.CLIRunner[DSSUnlockOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSUnlockOptions, *DSSUnlockVars](ctx)).vars = &DSSUnlockVars{baseVars: baseVars{uow: work}}
			return nil, dssUnlockRun(ctx)
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
	var (
		dss cabridss.HDss
		err error
	)
	if dss, err = NewHDss[DSSUnlockOptions, *DSSUnlockVars](ctx, func(bc *cabridss.DssBaseConfig) {
		bc.Unlock = !opts.LockForTest
	}, NewHDssArgs{}); err != nil {
		return err
	}
	if opts.LockForTest {
		return fmt.Errorf("leaving %s locked", dssUnlockCtx(ctx).args[0])
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

type DSSAuditOptions struct {
	BaseOptions
}

type DSSAuditVars struct {
	baseVars
}

func DSSAuditStartup(cr *joule.CLIRunner[DSSAuditOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSAuditOptions, *DSSAuditVars](ctx)).vars = &DSSAuditVars{baseVars: baseVars{uow: work}}
			return nil, dssAuditRun(ctx)
		})
	return nil
}

func DSSAuditShutdown(cr *joule.CLIRunner[DSSAuditOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssAuditCtx(ctx context.Context) *uiContext[DSSAuditOptions, *DSSAuditVars] {
	return uiCtxFrom[DSSAuditOptions, *DSSAuditVars](ctx)
}

func dssAuditOpts(ctx context.Context) DSSAuditOptions { return (*dssAuditCtx(ctx)).opts }

func dssAuditUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSAuditOptions, *DSSAuditVars](ctx)
}

func dssAuditOut(ctx context.Context, s string) { dssAuditUow(ctx).UiStrOut(s) }

func dssAuditRun(ctx context.Context) error {
	dss, err := NewHDss[DSSAuditOptions, *DSSAuditVars](ctx, nil, NewHDssArgs{})
	if err != nil {
		return err
	}
	defer dss.Close()
	_, perr := dss.AuditIndex()
	if perr != nil {
		return perr
	}
	return nil

}

type DSSScanOptions struct {
	BaseOptions
}

type DSSScanVars struct {
	baseVars
}

func DSSScanStartup(cr *joule.CLIRunner[DSSScanOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSScanOptions, *DSSScanVars](ctx)).vars = &DSSScanVars{baseVars: baseVars{uow: work}}
			return nil, dssScanRun(ctx)
		})
	return nil
}

func DSSScanShutdown(cr *joule.CLIRunner[DSSScanOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssScanCtx(ctx context.Context) *uiContext[DSSScanOptions, *DSSScanVars] {
	return uiCtxFrom[DSSScanOptions, *DSSScanVars](ctx)
}

func dssScanOpts(ctx context.Context) DSSScanOptions { return (*dssScanCtx(ctx)).opts }

func dssScanUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSScanOptions, *DSSScanVars](ctx)
}

func dssScanOut(ctx context.Context, s string) { dssScanUow(ctx).UiStrOut(s) }

func dssScanRun(ctx context.Context) error {
	dss, err := NewHDss[DSSScanOptions, *DSSScanVars](ctx, nil, NewHDssArgs{})
	if err != nil {
		return err
	}
	defer dss.Close()
	sti, perr := dss.ScanStorage()
	if perr != nil {
		return perr
	}
	_ = sti
	return nil

}

type DSSReindexOptions struct {
	BaseOptions
}

type DSSReindexVars struct {
	baseVars
}

func DSSReindexStartup(cr *joule.CLIRunner[DSSReindexOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSReindexOptions, *DSSReindexVars](ctx)).vars = &DSSReindexVars{baseVars: baseVars{uow: work}}
			return nil, dssReindexRun(ctx)
		})
	return nil
}

func DSSReindexShutdown(cr *joule.CLIRunner[DSSReindexOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssReindexCtx(ctx context.Context) *uiContext[DSSReindexOptions, *DSSReindexVars] {
	return uiCtxFrom[DSSReindexOptions, *DSSReindexVars](ctx)
}

func dssReindexOpts(ctx context.Context) DSSReindexOptions { return (*dssReindexCtx(ctx)).opts }

func dssReindexUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSReindexOptions, *DSSReindexVars](ctx)
}

func dssReindexOut(ctx context.Context, s string) { dssReindexUow(ctx).UiStrOut(s) }

func dssReindexRun(ctx context.Context) error {
	dss, err := NewHDss[DSSReindexOptions, *DSSReindexVars](ctx, nil, NewHDssArgs{})
	if err != nil {
		return err
	}
	defer dss.Close()
	sti, perr := dss.Reindex()
	if perr != nil {
		return perr
	}
	_ = sti
	return nil
}

type DSSLsHistoOptions struct {
	BaseOptions
	Recursive bool
	Sorted    bool
}

type DSSLsHistoVars struct {
	baseVars
}

func DSSLsHistoStartup(cr *joule.CLIRunner[DSSLsHistoOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSLsHistoOptions, *DSSLsHistoVars](ctx)).vars = &DSSLsHistoVars{baseVars: baseVars{uow: work}}
			return nil, dssLsHistoRun(ctx)
		})
	return nil
}

func DSSLsHistoShutdown(cr *joule.CLIRunner[DSSLsHistoOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssLsHistoCtx(ctx context.Context) *uiContext[DSSLsHistoOptions, *DSSLsHistoVars] {
	return uiCtxFrom[DSSLsHistoOptions, *DSSLsHistoVars](ctx)
}

func dssLsHistoOpts(ctx context.Context) DSSLsHistoOptions { return (*dssLsHistoCtx(ctx)).opts }

func dssLsHistoUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSLsHistoOptions, *DSSLsHistoVars](ctx)
}

func dssLsHistoOut(ctx context.Context, s string) { dssLsHistoUow(ctx).UiStrOut(s) }

func dssLsHistoRun(ctx context.Context) error {
	dss, err := NewHDss[DSSLsHistoOptions, *DSSLsHistoVars](ctx, nil, NewHDssArgs{})
	if err != nil {
		return err
	}
	defer dss.Close()
	args := dssLsHistoCtx(ctx).args
	_, _, npath, _ := CheckDssPath(args[0])

	mHes, err := dss.GetHistory(npath, dssLsHistoOpts(ctx).Recursive)
	if err != nil {
		return err
	}
	dssLsHistoOut(ctx, fmt.Sprintf("%s\n", internal.MapSliceStringer[cabridss.HistoryInfo]{Map: mHes}))
	return nil
}

type DSSRmHistoOptions struct {
	BaseOptions
	Recursive bool
	DryRun    bool
	StartTime string
	EndTime   string
}

type DSSRmHistoVars struct {
	baseVars
}

func DSSRmHistoStartup(cr *joule.CLIRunner[DSSRmHistoOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSRmHistoOptions, *DSSRmHistoVars](ctx)).vars = &DSSRmHistoVars{baseVars: baseVars{uow: work}}
			return nil, dssRmHistoRun(ctx)
		})
	return nil
}

func DSSRmHistoShutdown(cr *joule.CLIRunner[DSSRmHistoOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssRmHistoCtx(ctx context.Context) *uiContext[DSSRmHistoOptions, *DSSRmHistoVars] {
	return uiCtxFrom[DSSRmHistoOptions, *DSSRmHistoVars](ctx)
}

func dssRmHistoOpts(ctx context.Context) DSSRmHistoOptions { return (*dssRmHistoCtx(ctx)).opts }

func dssRmHistoUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSRmHistoOptions, *DSSRmHistoVars](ctx)
}

func dssRmHistoOut(ctx context.Context, s string) { dssRmHistoUow(ctx).UiStrOut(s) }

func dssRmHistoRun(ctx context.Context) error {
	dss, err := NewHDss[DSSRmHistoOptions, *DSSRmHistoVars](ctx, nil, NewHDssArgs{})
	if err != nil {
		return err
	}
	defer dss.Close()
	args := dssRmHistoCtx(ctx).args
	_, _, npath, _ := CheckDssPath(args[0])
	st, _ := CheckTimeStamp(dssRmHistoOpts(ctx).StartTime)
	et, _ := CheckTimeStamp(dssRmHistoOpts(ctx).EndTime)
	mHes, err := dss.RemoveHistory(npath, dssRmHistoOpts(ctx).Recursive, dssRmHistoOpts(ctx).DryRun, st, et)
	if err != nil {
		return err
	}
	dssRmHistoOut(ctx, fmt.Sprintf("%s\n", internal.MapSliceStringer[cabridss.HistoryInfo]{Map: mHes}))
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
			return nil, dssCleanRun(ctx)
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
	if mp, err = MasterPassword(dssCleanUow(ctx), opts, 0); err != nil {
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

type DSSConfigOptions struct {
	BaseOptions
	Raw bool
}

type DSSConfigVars struct {
	baseVars
}

func DSSConfigStartup(cr *joule.CLIRunner[DSSConfigOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[DSSConfigOptions, *DSSConfigVars](ctx)).vars = &DSSConfigVars{baseVars: baseVars{uow: work}}
			return nil, dssConfigRun(ctx)
		})
	return nil
}

func DSSConfigShutdown(cr *joule.CLIRunner[DSSConfigOptions]) error {
	return cr.GetUow("command").GetError()
}

func dssConfigCtx(ctx context.Context) *uiContext[DSSConfigOptions, *DSSConfigVars] {
	return uiCtxFrom[DSSConfigOptions, *DSSConfigVars](ctx)
}

func dssConfigOpts(ctx context.Context) DSSConfigOptions { return (*dssConfigCtx(ctx)).opts }

func dssConfigUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[DSSConfigOptions, *DSSConfigVars](ctx)
}

func dssConfigOut(ctx context.Context, s string) { dssConfigUow(ctx).UiStrOut(s) }

func dssConfigRun(ctx context.Context) error {
	opts := dssConfigOpts(ctx).BaseOptions
	args := dssConfigCtx(ctx).args
	dssType, root, _ := CheckDssSpec(args[0])
	var (
		config  cabridss.ObsConfig
		err     error
		mp      string
		changed bool
	)
	if mp, err = MasterPassword(dssConfigUow(ctx), opts, 0); err != nil {
		return err
	}
	dssSubType := dssType
	if dssType[0] == 'x' {
		dssSubType = dssType[1:]
	}
	if dssSubType == "obs" {
		config, err = GetObsConfig(opts, 0, root, mp)
		if err != nil {
			return err
		}
	} else if dssSubType == "smf" {
		config, err = GetSmfConfig(opts, 0, root, mp)
		if err != nil {
			return err
		}
	} else if dssConfigOpts(ctx).Raw && dssSubType == "olf" {
		olfConfig, err := GetOlfConfig(opts, 0, root, mp)
		if err != nil {
			return err
		}
		var pc cabridss.OlfConfig
		if err := cabridss.LoadDssConfig(olfConfig.DssBaseConfig, &pc); err != nil {
			return err
		}
		dssConfigOut(ctx, fmt.Sprintf("%+v\n", pc))
		return nil
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	var pc cabridss.ObsConfig
	if err := cabridss.LoadDssConfig(config.DssBaseConfig, &pc); err != nil {
		return err
	}
	if config.Endpoint != "" {
		pc.Endpoint = config.Endpoint
		changed = true
	}
	if config.Region != "" {
		pc.Region = config.Region
		changed = true
	}
	if config.AccessKey != "" {
		pc.AccessKey = config.AccessKey
		changed = true
	}
	if config.SecretKey != "" {
		pc.SecretKey = config.SecretKey
		changed = true
	}
	if config.Container != "" {
		pc.Container = config.Container
		changed = true
	}
	if changed {
		if err := cabridss.OverwriteDssConfig(config.DssBaseConfig, &pc); err != nil {
			return err
		}
	}
	if !dssConfigOpts(ctx).Raw {
		dssConfigOut(ctx, fmt.Sprintf(
			"--obsrg %s --obsep %s --obsct %s --obsak %s --obssk %s\n",
			pc.Region, pc.Endpoint, pc.Container, pc.AccessKey, pc.SecretKey))
	} else {
		dssConfigOut(ctx, fmt.Sprintf("%+v\n", pc))
	}
	return nil
}
