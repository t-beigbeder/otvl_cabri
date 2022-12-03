package cabriui

import (
	"bytes"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrisync"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabritbx"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"strings"
	"testing"
)

func lsnsCreateFsyOlf(t *testing.T, tfs *testfs.Fs) (fsy, bck, olf cabridss.Dss) {
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "fsy"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "bck"), 0755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if err := os.Mkdir(ufpath.Join(tfs.Path(), fmt.Sprintf("bck/%d", i)), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "olf"), 0755); err != nil {
		t.Fatal(err)
	}
	fsy, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, ufpath.Join(tfs.Path(), "fsy"))
	if err != nil {
		t.Error(err)
	}
	bck, err = cabridss.NewFsyDss(cabridss.FsyConfig{}, ufpath.Join(tfs.Path(), "bck"))
	if err != nil {
		t.Error(err)
	}
	olf, err = cabridss.CreateOlfDss(cabridss.OlfConfig{
		DssBaseConfig: cabridss.DssBaseConfig{LocalPath: ufpath.Join(tfs.Path(), "olf")},
		Root:          ufpath.Join(tfs.Path(), "olf"), Size: "l"})
	if err != nil {
		t.Error(err)
	}
	olf, err = cabridss.NewOlfDss(cabridss.OlfConfig{
		DssBaseConfig: cabridss.DssBaseConfig{LocalPath: ufpath.Join(tfs.Path(), "olf")},
		Root:          ufpath.Join(tfs.Path(), "olf")}, 0, nil)
	if err != nil {
		t.Error(err)
	}
	olf.SetCurrentTime(-1)
	if err = olf.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	return
}

func lsnsGenArboBase(t *testing.T, dss cabridss.Dss) cabritbx.RandGen {
	const rndNb = 100 // 500
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)
	if err := rg.Create(rndNb); err != nil {
		t.Error(err)
	}
	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Create(rndNb); err != nil {
		t.Error(err)
	}
	rg.AdvTime(3600 * 24 * 42)
	if err := rg.Update(rndNb); err != nil {
		t.Error(err)
	}
	return rg
}

func TestLsnsArboBase(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboBase", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, olf := lsnsCreateFsyOlf(t, tfs)
	fsyRg := lsnsGenArboBase(t, fsy)
	olfRg := lsnsGenArboBase(t, olf)
	if fsyRg.CurTime() != olfRg.CurTime() {
		t.Fatalf("unsynchronized")
	}
	var lsnsOptions = LsnsOptions{Sorted: true, Recursive: true, LastTime: ""}
	var outBuf1 bytes.Buffer
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf1, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy"))},
		LsnsStartup, LsnsShutdown)
	if err != nil {
		t.Fatal(err)
	}
	var outBuf2 bytes.Buffer
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf2, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf"))},
		LsnsStartup, LsnsShutdown)
	if err != nil || outBuf2.String() != outBuf1.String() {
		t.Fatal(err)
	}

	bo := getObjOptions()
	bo.IndexImplems = []string{"memory"}
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: bo}, []string{"obs:"})
	if err != nil {
		t.Error(err)
	}
	if err = dssTestMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{BaseOptions: bo, Children: []string{"d1/"}}, []string{"obs:@"}); err != nil {
		t.Error(err)
	}
	if err = dssTestMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{BaseOptions: bo}, []string{"obs:@d1"}); err != nil {
		t.Error(err)
	}
	var outBuf3 bytes.Buffer
	lsnsOptions = LsnsOptions{BaseOptions: bo}
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf3, os.Stderr,
		lsnsOptions, []string{"obs:@"},
		LsnsStartup, LsnsShutdown)
	if err != nil || !strings.Contains(outBuf3.String(), "d1/\n") {
		t.Fatal(err, outBuf3.String())
	}

	if err := os.Mkdir(ufpath.Join(tfs.Path(), "smf"), 0755); err != nil {
		t.Fatal(err)
	}
	ds := fmt.Sprintf("smf:%s/smf", tfs.Path())
	err = dssTestMkRun(os.Stdin, os.Stdout, os.Stderr, DSSMkOptions{BaseOptions: bo}, []string{ds})
	if err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("smf:%s/smf@", tfs.Path())
	if err = dssTestMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{Children: []string{"d1/"}}, []string{ds}); err != nil {
		t.Error(err)
	}
	ds = fmt.Sprintf("smf:%s/smf@d1", tfs.Path())
	if err = dssTestMknsRun(os.Stdin, os.Stdout, os.Stderr, DSSMknsOptions{Children: []string{}}, []string{ds}); err != nil {
		t.Error(err)
	}
	var outBuf4 bytes.Buffer
	lsnsOptions = LsnsOptions{BaseOptions: bo}
	ds = fmt.Sprintf("smf:%s/smf@", tfs.Path())
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf4, os.Stderr,
		lsnsOptions, []string{ds},
		LsnsStartup, LsnsShutdown)
	if err != nil || !strings.Contains(outBuf4.String(), "d1/\n") {
		t.Fatal(err, outBuf4.String())
	}

}

func lsnsGenArboNoFear(t *testing.T, dss, bck cabridss.Dss) (cabritbx.RandGen, []int64) {
	const rndNb1 = 500 // 500
	const rndNb2 = 100 // 100
	cts := make([]int64, 4)
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)

	if err := rg.Create(rndNb1); err != nil {
		t.Error(err)
	}
	cts[0] = rg.CurTime() + 3600
	if bck != nil {
		rp := cabrisync.Synchronize(nil, dss, "", bck, "0", cabrisync.SyncOptions{InDepth: true, NoACL: true})
		if !(rp.GetStats() == cabrisync.SyncStats{CreNum: 1476, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}

	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Update(rndNb1 + rndNb2); err != nil {
		t.Error(err)
	}
	cts[1] = rg.CurTime() + 3600
	if bck != nil {
		rp := cabrisync.Synchronize(nil, dss, "", bck, "1", cabrisync.SyncOptions{InDepth: true, NoACL: true})
		if !(rp.GetStats() == cabrisync.SyncStats{CreNum: 1583, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}

	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Update(rndNb1 + 2*rndNb2); err != nil {
		t.Error(err)
	}
	cts[2] = rg.CurTime() + 3600
	if bck != nil {
		rp := cabrisync.Synchronize(nil, dss, "", bck, "2", cabrisync.SyncOptions{InDepth: true, NoACL: true})
		if !(rp.GetStats() == cabrisync.SyncStats{CreNum: 1671, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}

	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Update(rndNb1 + 3*rndNb2); err != nil {
		t.Error(err)
	}
	cts[3] = rg.CurTime() + 3600

	return rg, cts
}

func TestLsnsArboNoFear(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboNoFear", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, bck, olf := lsnsCreateFsyOlf(t, tfs)
	fsyRg, fsyCts := lsnsGenArboNoFear(t, fsy, bck)
	olfRg, olfCts := lsnsGenArboNoFear(t, olf, nil)
	if fsyRg.CurTime() != olfRg.CurTime() || !(fsyCts[3] == olfCts[3]) {
		t.Fatalf("unsynchronized")
	}

	var lsnsOptions = LsnsOptions{Sorted: true, Recursive: true, LastTime: ""}
	var outBuf1 bytes.Buffer
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf1, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy"))},
		LsnsStartup, LsnsShutdown)
	if err != nil {
		t.Fatal(err)
	}
	var outBuf2 bytes.Buffer
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf2, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf"))},
		LsnsStartup, LsnsShutdown)
	if err != nil || outBuf2.String() != outBuf1.String() {
		t.Fatal(err)
	}

	outBuf1 = bytes.Buffer{}
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf1, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "bck/0"))},
		LsnsStartup, LsnsShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lsnsOptions.LastTime = fmt.Sprintf("%d", olfCts[0])
	outBuf2 = bytes.Buffer{}
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf2, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf"))},
		LsnsStartup, LsnsShutdown)
	if err != nil || outBuf2.String() != outBuf1.String() {
		t.Fatal(err)
	}

	outBuf1 = bytes.Buffer{}
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf1, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "bck/1"))},
		LsnsStartup, LsnsShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lsnsOptions.LastTime = fmt.Sprintf("%d", olfCts[1])
	outBuf2 = bytes.Buffer{}
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf2, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf"))},
		LsnsStartup, LsnsShutdown)
	if err != nil || outBuf2.String() != outBuf1.String() {
		t.Fatal(err)
	}

	outBuf1 = bytes.Buffer{}
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf1, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "bck/2"))},
		LsnsStartup, LsnsShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lsnsOptions.LastTime = fmt.Sprintf("%d", olfCts[2])
	outBuf2 = bytes.Buffer{}
	err = CLIRun[LsnsOptions, *LsnsVars](
		nil, &outBuf2, os.Stderr,
		lsnsOptions, []string{fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf"))},
		LsnsStartup, LsnsShutdown)
	if err != nil || outBuf2.String() != outBuf1.String() {
		t.Fatal(err)
	}

}
