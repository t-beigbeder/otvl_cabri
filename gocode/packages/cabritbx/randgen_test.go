package cabritbx

import (
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"testing"
)

func TestRandGenCreateFsy(t *testing.T) {
	tfs, err := testfs.CreateFs("TestRandGenCreate", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	dss, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfs.Path())
	if err != nil {
		t.Error(err)
	}
	rg := NewRanGen(GetDefaultConfig(), dss)
	if err = rg.Create(50); err != nil {
		t.Error(err)
	}
	rg.AdvTime(3600)
	if err = rg.Update(1000); err != nil {
		t.Error(err)
	}
}

func TestRandGenCreateOlf(t *testing.T) {
	tfs, err := testfs.CreateFs("TestRandGenCreate", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	dss, err := cabridss.CreateOlfDss(cabridss.OlfConfig{DssBaseConfig: cabridss.DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "s"})
	if err != nil {
		t.Error(err)
	}
	dss.SetCurrentTime(-1)
	if err = dss.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	rg := NewRanGen(GetDefaultConfig(), dss)
	if err = rg.Create(100); err != nil {
		t.Error(err)
	}
}
