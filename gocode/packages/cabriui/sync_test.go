package cabriui

import (
	"bytes"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrisync"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabritbx"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"strings"
	"testing"
	"time"
)

func syncCreateFsyOlf(t *testing.T, tfs *testfs.Fs) (fsy, bck, olf cabridss.Dss) {
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

func syncGenArboTiny(t *testing.T, dss cabridss.Dss) cabritbx.RandGen {
	const rndNb = 10
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

func TestSynchronizeArboTiny(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboTiny", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, olf := syncCreateFsyOlf(t, tfs)
	fsyRg := syncGenArboTiny(t, fsy)
	olfRg := syncGenArboTiny(t, olf)
	if fsyRg.CurTime() != olfRg.CurTime() {
		t.Fatalf("unsynchronized")
	}
	syncOptions := SyncOptions{Recursive: true, DryRun: true, NoACL: true}
	var outBuf bytes.Buffer
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 80 || lines[len(lines)-2] != "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" {
		t.Fatalf("stats %v %d", lines, len(lines))
	}

	syncOptions.VerboseLevel = 3
	err = CLIRun[SyncOptions, *SyncVars](
		nil, os.Stdout, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}

}

func syncGenArboBase(t *testing.T, dss cabridss.Dss) cabritbx.RandGen {
	const rndNb = 500
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

func TestSynchronizeArboBase(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboBase", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, olf := syncCreateFsyOlf(t, tfs)
	fsyRg := syncGenArboBase(t, fsy)
	olfRg := syncGenArboBase(t, olf)
	if fsyRg.CurTime() != olfRg.CurTime() {
		t.Fatalf("unsynchronized")
	}
	syncOptions := SyncOptions{Recursive: true, DryRun: true, NoACL: true}
	var outBuf bytes.Buffer
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 2838 || lines[len(lines)-2] != "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" {
		t.Fatalf("stats %v", lines[len(lines)-2])
	}
}

func syncGenArboNoFear(t *testing.T, dss, bck cabridss.Dss) (cabritbx.RandGen, []int64) {
	cts := make([]int64, 4)
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)

	if err := rg.Create(500); err != nil {
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
	if err := rg.Update(600); err != nil {
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
	if err := rg.Update(700); err != nil {
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
	if err := rg.Update(800); err != nil {
		t.Error(err)
	}
	cts[3] = rg.CurTime() + 3600

	return rg, cts
}

func TestSynchronizeArboNoFear(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboNoFear", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, bck, olf := syncCreateFsyOlf(t, tfs)
	fsyRg, fsyCts := syncGenArboNoFear(t, fsy, bck)
	olfRg, olfCts := syncGenArboNoFear(t, olf, nil)
	if fsyRg.CurTime() != olfRg.CurTime() || !(fsyCts[3] == olfCts[3]) {
		t.Fatalf("unsynchronized")
	}

	syncOptions := SyncOptions{Recursive: true, DryRun: true, NoACL: true}
	var outBuf bytes.Buffer
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 1814 || lines[len(lines)-2] != "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" {
		t.Fatalf("stats %d %v", len(lines), lines[len(lines)-2])
	}

	syncOptions = SyncOptions{Recursive: true, DryRun: true, RightTime: "2002-01-19T02:56:32Z", NoACL: true}
	outBuf = bytes.Buffer{}
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "bck/0")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 1479 || lines[len(lines)-2] != "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" {
		t.Fatalf("stats %d %v", len(lines), lines[len(lines)-2])
	}

	syncOptions = SyncOptions{Recursive: true, DryRun: true, RightTime: "2002-02-21T22:52:31Z", NoACL: true}
	outBuf = bytes.Buffer{}
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "bck/1")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 1586 || lines[len(lines)-2] != "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" {
		t.Fatalf("stats %d %v", len(lines), lines[len(lines)-2])
	}

	syncOptions = SyncOptions{Recursive: true, DryRun: true, RightTime: "2002-03-29T13:26:54Z", NoACL: true}
	outBuf = bytes.Buffer{}
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "bck/2")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 1674 || lines[len(lines)-2] != "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" {
		t.Fatalf("stats %d %v", len(lines), lines[len(lines)-2])
	}
}

func syncGenArboBiDir(t *testing.T, dss, bck cabridss.Dss) []int64 {
	cts := make([]int64, 4)
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)

	if err := rg.Create(500); err != nil {
		t.Error(err)
	}
	cts[0] = rg.CurTime() + 3600
	if bck != nil {
		rp := cabrisync.Synchronize(nil, dss, "", bck, "0", cabrisync.SyncOptions{InDepth: true, NoACL: true})
		if !(rp.GetStats() == cabrisync.SyncStats{CreNum: 1476, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}
	rgConfig := cabritbx.GetDefaultConfig()
	if bck != nil {
		rgConfig.Seed = 22
		rgConfig.TimeOrigin = rg.CurTime() + 3600*24*22
	} else {
		rgConfig.Seed = 23
		rgConfig.TimeOrigin = rg.CurTime() + 3600*24*23
	}
	rg = cabritbx.NewRanGen(rgConfig, dss)
	if err := rg.Update(250); err != nil {
		t.Error(err)
	}
	cts[1] = rg.CurTime() + 3600
	if bck != nil {
		rp := cabrisync.Synchronize(nil, dss, "", bck, "1", cabrisync.SyncOptions{InDepth: true, NoACL: true})
		if !(rp.GetStats() == cabrisync.SyncStats{CreNum: 1533, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}
	return cts
}

func TestSynchronizeArboBiDir(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboBiDir", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, bck, olf := syncCreateFsyOlf(t, tfs)
	fsyCts := syncGenArboBiDir(t, fsy, bck)
	olfCts := syncGenArboBiDir(t, olf, nil)
	if !(fsyCts[0] == olfCts[0]) {
		t.Fatalf("unsynchronized")
	}
	//rp := Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true})
	//if !(rp.GetStats() == SyncStats{CreNum: 210, UpdNum: 321, RmvNum: 0, MUpNum: 4}) || len(rp.Entries) != 1642 {
	//	t.Fatal(rp.GetStats(), len(rp.Entries))
	//}

	syncOptions := SyncOptions{Recursive: true, DryRun: true, BiDir: true, NoACL: true}
	var outBuf bytes.Buffer
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 1644 || lines[len(lines)-2] != "created: 210, updated 321, removed 0, kept 0, touched 4, error(s) 0" {
		t.Fatalf("stats %d %v", len(lines), lines[len(lines)-2])
	}

	syncOptions = SyncOptions{Recursive: true, DryRun: false, BiDir: true, Verbose: true, NoACL: true}
	outBuf = bytes.Buffer{}
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 1644 || lines[len(lines)-2] != "created: 210, updated 321, removed 0, kept 0, touched 4, error(s) 0" {
		t.Fatalf("stats %d %v", len(lines), lines[len(lines)-2])
	}

	syncOptions = SyncOptions{Recursive: true, DryRun: true, BiDir: true, NoACL: true}
	outBuf = bytes.Buffer{}
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", ufpath.Join(tfs.Path(), "fsy")),
			fmt.Sprintf("olf:%s@", ufpath.Join(tfs.Path(), "olf")),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 1644 || lines[len(lines)-2] != "created: 0, updated 0, removed 0, kept 0, touched 0, error(s) 0" {
		t.Fatalf("stats %d %v", len(lines), lines[len(lines)-2])
	}
}

func runTestSynchronizeToObs(t *testing.T) error {
	// object storage is eventually consistent, so this test may fail
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeToObs", func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	defer tfs.Delete()

	if err := cabridss.CleanObsDss(getObsConfig()); err != nil {
		return err
	}
	ods, err := cabridss.NewObsDss(getObsConfig(), 0, nil)
	if err != nil {
		return err
	}
	if err := ods.Mkns("", time.Now().Unix(), []string{}, nil); err != nil {
		return err
	}
	cs, err := ods.Lsns("")
	if err != nil || len(cs) != 0 {
		return fmt.Errorf("%s %s", cs, err)
	}
	if err := ods.Mkns("", time.Now().Unix(), []string{"s1/", "s2/"}, nil); err != nil {
		return err
	}
	if err := ods.Mkns("s1", time.Now().Unix(), []string{}, nil); err != nil {
		return err
	}
	if err := ods.Mkns("s2", time.Now().Unix(), []string{}, nil); err != nil {
		return err
	}

	bo := getObjOptions()
	bo.IndexImplems = []string{"no", "memory"}
	syncOptions := SyncOptions{BaseOptions: bo, Recursive: true, DryRun: true, NoACL: true}
	var outBuf bytes.Buffer
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", tfs.Path()),
			fmt.Sprintf("obs:@s1"),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		return err
	}
	lines := strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 6 || lines[len(lines)-2] != "created: 3, updated 1, removed 0, kept 0, touched 0, error(s) 0" {
		return fmt.Errorf("stats %d %v", len(lines), lines[len(lines)-2])
	}

	syncOptions = SyncOptions{BaseOptions: bo, Recursive: true, Verbose: true, NoACL: true}
	outBuf = bytes.Buffer{}
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("fsy:%s@", tfs.Path()),
			fmt.Sprintf("obs:@s1"),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		return err
	}
	lines = strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 6 || lines[len(lines)-2] != "created: 3, updated 1, removed 0, kept 0, touched 0, error(s) 0" {
		return fmt.Errorf("stats %d %v", len(lines), lines)
	}

	syncOptions = SyncOptions{BaseOptions: bo, Recursive: true, Verbose: true, NoACL: true}
	outBuf = bytes.Buffer{}
	err = CLIRun[SyncOptions, *SyncVars](
		nil, &outBuf, os.Stderr,
		syncOptions, []string{
			fmt.Sprintf("obs:@s1"),
			fmt.Sprintf("obs:@s2"),
		},
		SyncStartup, SyncShutdown)
	if err != nil {
		return err
	}
	lines = strings.Split(outBuf.String(), "\n")
	_ = lines
	if len(lines) != 6 || lines[len(lines)-2] != "created: 3, updated 1, removed 0, kept 0, touched 0, error(s) 0" {
		return fmt.Errorf("stats %d %v", len(lines), lines)
	}
	return nil
}

func TestSynchronizeToObs(t *testing.T) {
	internal.Retry(t, runTestSynchronizeToObs)
}
