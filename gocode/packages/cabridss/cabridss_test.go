package cabridss

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"io"
	"testing"
	"time"
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
	if meta.GetIsNs() || meta.GetChildren() != nil || len(meta.GetAcl()) != 3 || meta.GetItime()/1e9 != meta.GetMtime() || meta.GetCh() != "12efb6ee023559a9dfd8a12c5fd7faea" || !meta2.Equals(meta, true) {
		t.Errorf("TestMetaBasic %v", meta)
	}
	meta, err = dss.GetMeta("d/", true)
	if err != nil {
		t.Error(err)
	}
	if !meta.GetIsNs() || len(meta.GetChildren()) != 1 || len(meta.GetAcl()) != 3 || meta.GetItime()/1e9 != meta.GetMtime() || meta.GetCh() != "c880c199d0db1b5a2018f30227dacea8" {
		t.Errorf("TestMetaBasic %v", meta)
	}
}

func TestTimeResolutionAlign(t *testing.T) {
	trs := []TimeResolution{"s", "m", "h", "d"}
	display := func(tm time.Time) {
		ns := tm.UnixNano()
		for _, tr := range trs {
			dsp := fmt.Sprintf("%s %s\n", tr, UnixNanoUTC(tr.Align(ns)))
			fmt.Fprint(io.Discard, dsp)
			//fmt.Fprint(os.Stdout, dsp)

		}
	}
	anniv := time.Date(1918, time.April, 24, 23, 22, 21, 20, time.UTC)
	display(anniv)
	anniv2 := time.Date(2018, time.April, 24, 23, 22, 21, 20, time.UTC)
	display(anniv2)
	origin := time.Unix(0, 0)
	display(origin)
	justBefore := time.Unix(0, -1)
	display(justBefore)
	justAfter := time.Unix(0, 1)
	display(justAfter)
	dayJustBefore := time.Unix(-24*3600+3661, 1)
	display(dayJustBefore)
	dayJustAfter := time.Unix(24*3600+3661, 1)
	display(dayJustAfter)
}
