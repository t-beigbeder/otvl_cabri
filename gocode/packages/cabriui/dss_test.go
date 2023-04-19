package cabriui

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func dssTestMkRun(
	cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts DSSMkOptions, args []string,
) error { // FIXME: migrate to use new code
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
		oc, err := GetObsConfig(opts.BaseOptions, 0, root, "")
		if err != nil {
			return err
		}
		if dss, err = cabridss.CreateObsDss(oc); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root, "")
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

func dssTestMknsRun(
	cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts DSSMknsOptions, args []string,
) error { // FIXME: migrate to use new code
	dssType, root, npath, _ := CheckDssPath(args[0])
	var dss cabridss.Dss
	var err error
	if dssType == "fsy" {
		if dss, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, root); err != nil {
			return err
		}
	} else if dssType == "olf" {
		oc, err := GetOlfConfig(opts.BaseOptions, 0, root, "")
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewOlfDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root, "")
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewObsDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root, "")
		if err != nil {
			return err
		}
		if dss, err = cabridss.NewObsDss(sc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "webapi+http" {
		frags := strings.Split(root[2:], "/")
		wc, err := GetWebConfig(opts.BaseOptions, 0, false, frags[0], frags[1], UiRunEnv{})
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

func dssTestUnlockRun(cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts DSSUnlockOptions, args []string,
) error { // FIXME: migrate to use new code
	dssType, root, _ := CheckDssSpec(args[0])

	var dss cabridss.HDss
	var err error
	if dssType == "olf" {
		oc, err := GetOlfConfig(opts.BaseOptions, 0, root, "")
		if err != nil {
			return err
		}
		oc.DssBaseConfig.Unlock = true
		if dss, err = cabridss.NewOlfDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(opts.BaseOptions, 0, root, "")
		if err != nil {
			return err
		}
		oc.DssBaseConfig.Unlock = true
		if dss, err = cabridss.NewObsDss(oc, 0, nil); err != nil {
			return err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(opts.BaseOptions, 0, root, "")
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

func TestDSSMkBase(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestDSSMkBase", func(f *testfs.Fs) error {
		os.Mkdir(filepath.Join(f.Path(), "fsy"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "solf"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "molf"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "lolf"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "obs"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "smf"), 0o777)
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	ds := fmt.Sprintf("fsy:%s/fsy", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("olf:%s/solf", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "s"}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("olf:%s/molf", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "m"}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("olf:%s/lolf", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "l"}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("obs:%s/obs", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: getObjOptions()}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("smf:%s/smf", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{}, []string{ds})
	if err != nil {
		t.Error(err)
	}
}

func TestdssTestMknsRun(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestDSSMkBase", func(f *testfs.Fs) error {
		os.Mkdir(filepath.Join(f.Path(), "fsy"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "olf"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "obs"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "smf"), 0o777)
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()

	ds := fmt.Sprintf("fsy:%s/fsy@", tfs.Path())
	if err = dssTestMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{Children: []string{"d1/", "a.txt"}}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("olf:%s/olf", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "s"}, []string{ds})
	if err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("obs:%s/obs", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: getObjOptions()}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("obs:%s/obs@", tfs.Path())
	if err = dssTestMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{BaseOptions: getObjOptions(), Children: []string{"d1/", "a.txt"}}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("smf:%s/smf", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: BaseOptions{}}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("smf:%s/smf@", tfs.Path())
	if err = dssTestMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{Children: []string{"d1/", "a.txt"}}, []string{ds}); err != nil {
		t.Error(err)
	}

}

func TestdssTestUnlockRun(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestdssTestUnlockRun", func(f *testfs.Fs) error {
		os.Mkdir(filepath.Join(f.Path(), "fsy"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "olf"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "obs"), 0o777)
		os.Mkdir(filepath.Join(f.Path(), "smf"), 0o777)
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()

	ds := fmt.Sprintf("fsy:%s/fsy", tfs.Path())
	if err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

	bo := BaseOptions{IndexImplems: []string{"bdb"}}
	ds = fmt.Sprintf("olf:%s/olf", tfs.Path())
	if err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: bo, Size: "s"}, []string{ds}); err != nil {
		t.Error(err)
	}
	if err = dssTestUnlockRun(os.Stdin, os.Stdout, os.Stderr, DSSUnlockOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("obs:%s/obs", tfs.Path())
	if err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: getObjOptions()}, []string{ds}); err != nil {
		t.Error(err)
	}
	if err = dssTestUnlockRun(os.Stdin, os.Stdout, os.Stderr, DSSUnlockOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("smf:%s/smf", tfs.Path())
	if err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: bo}, []string{ds}); err != nil {
		t.Error(err)
	}
	if err = dssTestUnlockRun(os.Stdin, os.Stdout, os.Stderr, DSSUnlockOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

}
