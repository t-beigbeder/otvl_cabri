package cabridss

import (
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"testing"
)

func TestCreateOlfDssErr(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestCreateOlfDssErr", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	_, err = CreateOlfDss(OlfConfig{Root: ufpath.Join(tfs.Path(), "no"), Size: "l"})
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with no such file or directory error")
	}
	_, err = CreateOlfDss(OlfConfig{Root: "/dev/null", Size: "l"})
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with not a diretory error")
	}
	_ = os.Mkdir(ufpath.Join(tfs.Path(), "meta"), 0o777)
	_, err = CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "l"})
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with cannot create error")
	}
	_ = os.Remove(ufpath.Join(tfs.Path(), "meta"))
	_ = os.Mkdir(ufpath.Join(tfs.Path(), "content"), 0o777)
	_, err = CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "l"})
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with cannot create error")
	}
	_ = os.Remove(ufpath.Join(tfs.Path(), "meta"))
	_ = os.Remove(ufpath.Join(tfs.Path(), "content"))
	_ = os.Mkdir(ufpath.Join(tfs.Path(), "tmp"), 0o777)
	_, err = CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "l"})
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with cannot create error")
	}
}

func runTestNewOlfVarSizesDssOk(t *testing.T, size string) error {
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: size})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewOlfDss(OlfConfig{
				Root: tfs.Path(),
				DssBaseConfig: DssBaseConfig{
					LocalPath: tfs.Path(),
					GetIndex: func(config DssBaseConfig, _ string) (Index, error) {
						return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
					}}}, 0, nil)
			return dss, err
		}); err != nil {
		return err
	}
	return nil
}

func TestNewOlfSmallDssOk(t *testing.T) {
	if err := runTestNewOlfVarSizesDssOk(t, "s"); err != nil {
		t.Fatal(err)
	}
}

func TestNewOlfMediumDssOk(t *testing.T) {
	if err := runTestNewOlfVarSizesDssOk(t, "m"); err != nil {
		t.Fatal(err)
	}
}

func TestNewOlfLargeDssOk(t *testing.T) {
	if err := runTestNewOlfVarSizesDssOk(t, "l"); err != nil {
		t.Fatal(err)
	}
}

func TestOlfDssMindex(t *testing.T) {
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewOlfDss(OlfConfig{
				Root: tfs.Path(),
				DssBaseConfig: DssBaseConfig{
					LocalPath: tfs.Path(),
					GetIndex: func(config DssBaseConfig, _ string) (Index, error) {
						return NewMIndex(), nil
					}}}, 0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func TestOlfHistory(t *testing.T) {
	if err := runTestHistory(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewOlfDss(OlfConfig{
				Root: tfs.Path(),
				DssBaseConfig: DssBaseConfig{
					LocalPath: tfs.Path(),
					GetIndex: func(config DssBaseConfig, _ string) (Index, error) {
						return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
					}}}, 0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}

}

func TestOlfMultiHistory(t *testing.T) {
	if err := runTestMultiHistory(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewOlfDss(OlfConfig{
				Root: tfs.Path(),
				DssBaseConfig: DssBaseConfig{
					LocalPath: tfs.Path(),
					GetIndex: func(config DssBaseConfig, _ string) (Index, error) {
						return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
					}}}, 0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}
