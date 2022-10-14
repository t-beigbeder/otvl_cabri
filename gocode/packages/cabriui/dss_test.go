package cabriui

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"os"
	"path/filepath"
	"testing"
)

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
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("olf:%s/solf", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "s"}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("olf:%s/molf", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "m"}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("olf:%s/lolf", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "l"}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("obs:%s/obs", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: getObjOptions()}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("smf:%s/smf", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{}, []string{ds})
	if err != nil {
		t.Error(err)
	}
}

func TestDSSMknsRun(t *testing.T) {
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
	if err = DSSMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{Children: []string{"d1/", "a.txt"}}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("olf:%s/olf", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{Size: "s"}, []string{ds})
	if err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("obs:%s/obs", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: getObjOptions()}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("obs:%s/obs@", tfs.Path())
	if err = DSSMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{BaseOptions: getObjOptions(), Children: []string{"d1/", "a.txt"}}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("smf:%s/smf", tfs.Path())
	err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: BaseOptions{}}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("smf:%s/smf@", tfs.Path())
	if err = DSSMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{Children: []string{"d1/", "a.txt"}}, []string{ds}); err != nil {
		t.Error(err)
	}

}

func TestDSSUnlockRun(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestDSSUnlockRun", func(f *testfs.Fs) error {
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
	if err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

	bo := BaseOptions{IndexImplems: []string{"bdb"}}
	ds = fmt.Sprintf("olf:%s/olf", tfs.Path())
	if err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: bo, Size: "s"}, []string{ds}); err != nil {
		t.Error(err)
	}
	if err = DSSUnlockRun(os.Stdin, os.Stdout, os.Stderr, DSSUnlockOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("obs:%s/obs", tfs.Path())
	if err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: getObjOptions()}, []string{ds}); err != nil {
		t.Error(err)
	}
	if err = DSSUnlockRun(os.Stdin, os.Stdout, os.Stderr, DSSUnlockOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

	ds = fmt.Sprintf("smf:%s/smf", tfs.Path())
	if err = DSSMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: bo}, []string{ds}); err != nil {
		t.Error(err)
	}
	if err = DSSUnlockRun(os.Stdin, os.Stdout, os.Stderr, DSSUnlockOptions{}, []string{ds}); err != nil {
		t.Error(err)
	}

}
