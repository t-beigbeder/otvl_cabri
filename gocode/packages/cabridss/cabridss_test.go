package cabridss

import (
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"testing"
)

func TestMetaBasic(t *testing.T) {
	tfs, err := testfs.CreateFs("TestMetaBasic", tfsStartup)
	if err != nil {
		t.Error(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Error(err)
	}
	meta, err := dss.GetMeta("d/b.txt", true)
	if err != nil {
		t.Error(err)
	}
	meta2, _ := dss.GetMeta("d/b.txt", true)
	if meta.GetIsNs() || meta.GetChildren() != nil || len(meta.GetAcl()) != 3 || meta.GetItime() != meta.GetMtime() || meta.GetCh() != "12efb6ee023559a9dfd8a12c5fd7faea" || !meta2.Equals(meta, true) {
		t.Errorf("TestMetaBasic %v", meta)
	}
	meta, err = dss.GetMeta("d/", true)
	if err != nil {
		t.Error(err)
	}
	if !meta.GetIsNs() || len(meta.GetChildren()) != 1 || len(meta.GetAcl()) != 3 || meta.GetItime() != meta.GetMtime() || meta.GetCh() != "c880c199d0db1b5a2018f30227dacea8" {
		t.Errorf("TestMetaBasic %v", meta)
	}
}
