package cabriui

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"io"
	"strings"
	"time"
)

type DSSMkOptions struct {
	BaseOptions
	Size string
}

func DSSMkRun(
	cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts DSSMkOptions, args []string,
) error {
	dssType, root, _ := CheckDssSpec(args[0])
	var dss cabridss.Dss
	var err error
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return err
		}
	} else if dssType == "olf" {
		if dss, err = cabridss.CreateOlfDss(cabridss.OlfConfig{
			DssBaseConfig: cabridss.DssBaseConfig{LocalPath: root},
			Root:          root, Size: opts.Size}); err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root)
		if err != nil {
			return err
		}
		if dss, err = cabridss.CreateObsDss(oc); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root)
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

func DSSMknsRun(
	cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts DSSMknsOptions, args []string,
) error {
	dssType, root, npath, _ := CheckDssPath(args[0])
	var dss cabridss.Dss
	var err error
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return err
		}
	} else if dssType == "olf" {
		oc, err := GetOlfConfig(opts.BaseOptions, 0, root)
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewOlfDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root)
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewObsDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root)
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewObsDss(sc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "webapi+http" {
		frags := strings.Split(root[2:], "/")
		wc, err := GetWebConfig(opts.BaseOptions, 0, frags[0], frags[1])
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewWebDss(wc, 0, nil); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	if err = dss.Mkns(npath, time.Now().Unix(), opts.Children, nil); err != nil {
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

func DSSUnlockRun(cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts DSSUnlockOptions, args []string,
) error {
	dssType, root, _ := CheckDssSpec(args[0])

	var dss cabridss.HDss
	var err error
	if dssType == "olf" {
		oc, err := GetOlfConfig(opts.BaseOptions, 0, root)
		if err != nil {
			return err
		}
		oc.DssBaseConfig.Unlock = true
		if dss, err = cabridss.NewOlfDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root)
		if err != nil {
			return err
		}
		oc.DssBaseConfig.Unlock = true
		if dss, err = cabridss.NewObsDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root)
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
			fmt.Fprintln(cliOut, d)
		}
	}
	if err = dss.Close(); err != nil {
		return err
	}
	return nil

}

func DSSCleanRun(cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts BaseOptions, args []string,
) error {
	dssType, root, _ := CheckDssSpec(args[0])
	var config cabridss.ObsConfig
	var err error
	if dssType == "obs" {
		config, err = GetObsConfig(opts, 0, root)
		if err != nil {
			return err
		}
	} else if dssType == "smf" {
		config, err = GetSmfConfig(opts, 0, root)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	return cabridss.CleanObsDss(config)
}
