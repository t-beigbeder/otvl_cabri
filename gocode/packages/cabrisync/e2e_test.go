package cabrisync

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabritbx"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"testing"
	"time"
)

var dbgOut *os.File

func getMockFsObsConfigDbg(tfs *testfs.Fs) cabridss.ObsConfig {
	var err error
	if dbgOut == nil {
		dbgOut, err = os.Create("/tmp/dbgOut.txt")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	return cabridss.ObsConfig{
		GetS3Session: func() cabridss.IS3Session {
			return cabridss.NewS3sMockFs(ufpath.Join(tfs.Path(), "obs"), func(parent cabridss.IS3Session) cabridss.IS3Session {
				return cabridss.NewS3sMockTests(parent, func(args ...any) interface{} {
					if len(args) > 2 {
						fmt.Fprintf(dbgOut, "%s S3sMockTests %v %v\n", time.Now().Format("2006-01-02 15:04:05.000"), args[1], args[2])
					} else {
						fmt.Fprintf(dbgOut, "%s S3sMockTests %v\n", time.Now().Format("2006-01-02 15:04:05.000"), args[1])
					}
					return nil
				})
			})
		},
	}
}

func getMockFsObsConfig(tfs *testfs.Fs) cabridss.ObsConfig {
	return cabridss.ObsConfig{
		GetS3Session: func() cabridss.IS3Session {
			return cabridss.NewS3sMockFs(ufpath.Join(tfs.Path(), "obs"), nil)
		},
	}
}

func createFsyOlf(t *testing.T, tfs *testfs.Fs, hasObs bool) (fsy, bck, olf, obs cabridss.Dss) {
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
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "obs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "smf"), 0755); err != nil {
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
	olf, err = cabridss.CreateOlfDss(cabridss.OlfConfig{DssBaseConfig: cabridss.DssBaseConfig{LocalPath: ufpath.Join(tfs.Path(), "olf")}, Root: ufpath.Join(tfs.Path(), "olf"), Size: "s"})
	if err != nil {
		t.Error(err)
	}
	olf.SetCurrentTime(1)
	if err = olf.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	olfConfig := cabridss.OlfConfig{
		DssBaseConfig: cabridss.DssBaseConfig{
			LocalPath: ufpath.Join(tfs.Path(), "olf"),
			GetIndex: func(_ cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
				return cabridss.NewMIndex(), nil
			},
		},
		Root: ufpath.Join(tfs.Path(), "olf"),
	}
	olf, err = cabridss.NewOlfDss(olfConfig, 0, nil)
	if err != nil {
		t.Error(err)
	}

	if !hasObs {
		return
	}
	cabridss.CleanObsDss(getOC())
	obsConfig := getMockFsObsConfig(tfs)
	obsConfig.DssBaseConfig.GetIndex = func(_ cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
		return cabridss.NewMIndex(), nil
	}
	obs, err = cabridss.NewObsDss(obsConfig, 0, nil)
	obs.SetCurrentTime(1)
	if err = obs.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	return
}

func genArboTiny(t *testing.T, dss cabridss.Dss) cabritbx.RandGen {
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)
	if err := rg.Create(50); err != nil {
		t.Error(err)
	}
	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Create(50); err != nil {
		t.Error(err)
	}
	rg.AdvTime(3600 * 24 * 42)
	if err := rg.Update(50); err != nil {
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
	fsy, _, olf, obs := createFsyOlf(t, tfs, true)

	fsyRg := genArboTiny(t, fsy)
	olfRg := genArboTiny(t, olf)
	obsRg := genArboTiny(t, obs)
	if fsyRg.CurTime() != olfRg.CurTime() || fsyRg.CurTime() != obsRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}

	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
}

func TestSynchronizeArboSmfPix(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboSmfPix", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, _, _ := createFsyOlf(t, tfs, true)

	obsConfig := getMockFsObsConfig(tfs)
	obsConfig.LocalPath = ufpath.Join(tfs.Path(), "obs")
	obsConfig.DssBaseConfig.GetIndex = cabridss.GetPIndex
	obs, err := cabridss.CreateObsDss(obsConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err = obs.Close(); err != nil {
		t.Fatal(err)
	}
	obs, err = cabridss.NewObsDss(obsConfig, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	obs.SetCurrentTime(1)
	if err = obs.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	fsyRg := genArboTiny(t, fsy)
	obsRg := genArboTiny(t, obs)
	if fsyRg.CurTime() != obsRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
}

func TestSynchronizeArboObsPix(t *testing.T) {
	optionalSkip(t)
	if os.Getenv("CABRISYNC_KEEP_DEV_TESTS") == "" {
		t.Skip(fmt.Sprintf("Skipping %s because you didn't set CABRISYNC_KEEP_DEV_TESTS", t.Name()))
	}
	tfs, err := testfs.CreateFs("TestSynchronizeArboObsPix", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, _, _ := createFsyOlf(t, tfs, true)

	obsConfig := getOC()
	obsConfig.LocalPath = ufpath.Join(tfs.Path(), "obs")
	obsConfig.DssBaseConfig.GetIndex = cabridss.GetPIndex
	obs, err := cabridss.CreateObsDss(obsConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err = obs.Close(); err != nil {
		t.Fatal(err)
	}
	obs, err = cabridss.NewObsDss(obsConfig, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	obs.SetCurrentTime(1)
	if err = obs.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	fsyRg := genArboTiny(t, fsy)
	obsRg := genArboTiny(t, obs)
	if fsyRg.CurTime() != obsRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
}

func TestSynchronizeArboWebDssClientOlf(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboWebDssClientOlf", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, _, _ := createFsyOlf(t, tfs, true)

	getPIndex := func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
		return cabridss.NewPIndex(ufpath.Join(tfs.Path(), "olf", "index.bdb"), false, false)
	}

	sv, err := createWebDssServer(":3000", "",
		cabridss.CreateNewParams{Create: false, DssType: "olf", Root: ufpath.Join(tfs.Path(), "olf"), GetIndex: getPIndex},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sv.Shutdown()

	wolf, err := cabridss.NewWebDss(
		cabridss.WebDssConfig{
			DssBaseConfig: cabridss.DssBaseConfig{
				UserConfigPath: ufpath.Join(tfs.Path(), ".cabri"),
				WebPort:        "3000",
			}, NoClientLimit: true},
		0, nil)
	if err != nil {
		t.Fatal(err)
	}
	wolf.SetCurrentTime(1)
	if err = wolf.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}

	fsyRg := genArboTiny(t, fsy)
	wolfRg := genArboTiny(t, wolf)
	if fsyRg.CurTime() != wolfRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", wolf, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wolf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wolf, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wolf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
}

func TestSynchronizeArboWebDssClientObs(t *testing.T) {
	optionalSkip(t)
	if os.Getenv("CABRISYNC_KEEP_DEV_TESTS") == "" {
		t.Skip(fmt.Sprintf("Skipping %s because you didn't set CABRISYNC_KEEP_DEV_TESTS", t.Name()))
	}
	tfs, err := testfs.CreateFs("TestSynchronizeArboWebDssClientObs", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, _, _ := createFsyOlf(t, tfs, false)

	cabridss.CleanObsDss(getOC())
	getPIndex := func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
		return cabridss.NewPIndex(ufpath.Join(tfs.Path(), "obs", "index.bdb"), false, false)
	}
	sv, err := createWebDssServer(":3000", "",
		cabridss.CreateNewParams{
			Create: true, DssType: "obs", LocalPath: ufpath.Join(tfs.Path(), "obs"), GetIndex: getPIndex,
			Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sv.Shutdown()

	wobs, err := cabridss.NewWebDss(
		cabridss.WebDssConfig{
			DssBaseConfig: cabridss.DssBaseConfig{
				UserConfigPath: ufpath.Join(tfs.Path(), ".cabri"),
				WebPort:        "3000",
			}, NoClientLimit: true},
		0, nil)
	if err != nil {
		t.Fatal(err)
	}
	wobs.SetCurrentTime(1)
	if err = wobs.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}

	fsyRg := genArboTiny(t, fsy)
	wobsRg := genArboTiny(t, wobs)
	if fsyRg.CurTime() != wobsRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", wobs, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wobs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wobs, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wobs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
}

func TestSynchronizeArboWebDssClientSmf(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboWebDssClientSmf", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, _, _, _ := createFsyOlf(t, tfs, false)

	cabridss.CleanObsDss(getOC())
	getPIndex := func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
		return cabridss.NewPIndex(ufpath.Join(tfs.Path(), "smf", "index.bdb"), false, false)
	}
	sv, err := createWebDssServer(":3000", "",
		cabridss.CreateNewParams{
			Create: true, DssType: "smf", LocalPath: ufpath.Join(tfs.Path(), "smf"), GetIndex: getPIndex,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sv.Shutdown()

	wsmf, err := cabridss.NewWebDss(
		cabridss.WebDssConfig{
			DssBaseConfig: cabridss.DssBaseConfig{
				UserConfigPath: ufpath.Join(tfs.Path(), ".cabri"),
				WebPort:        "3000",
			}, NoClientLimit: true},
		0, nil)
	if err != nil {
		t.Fatal(err)
	}
	wsmf.SetCurrentTime(1)
	if err = wsmf.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}

	fsyRg := genArboTiny(t, fsy)
	wsmfRg := genArboTiny(t, wsmf)
	if fsyRg.CurTime() != wsmfRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 375}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
}

func genArboBase(t *testing.T, dss cabridss.Dss) cabritbx.RandGen {
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)
	if err := rg.Create(500); err != nil {
		t.Error(err)
	}
	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Create(500); err != nil {
		t.Error(err)
	}
	rg.AdvTime(3600 * 24 * 42)
	if err := rg.Update(500); err != nil {
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
	fsy, _, olf, obs := createFsyOlf(t, tfs, true)
	fsyRg := genArboBase(t, fsy)
	olfRg := genArboBase(t, olf)
	obsRg := genArboBase(t, obs)
	if fsyRg.CurTime() != olfRg.CurTime() || fsyRg.CurTime() != obsRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 2836}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 2836}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}

	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 2836}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 2836}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}

	getPIndex := func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
		return cabridss.NewPIndex(ufpath.Join(tfs.Path(), "smf", "index.bdb"), false, false)
	}
	sv, err := createWebDssServer(":3000", "",
		cabridss.CreateNewParams{
			Create: true, DssType: "smf", LocalPath: ufpath.Join(tfs.Path(), "smf"), GetIndex: getPIndex,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sv.Shutdown()

	wsmf, err := cabridss.NewWebDss(
		cabridss.WebDssConfig{
			DssBaseConfig: cabridss.DssBaseConfig{
				UserConfigPath: ufpath.Join(tfs.Path(), ".cabri"),
				WebPort:        "3000",
			}, NoClientLimit: true},
		0, nil)
	if err != nil {
		t.Fatal(err)
	}
	wsmf.SetCurrentTime(1)
	if err = wsmf.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	wsmfRg := genArboBase(t, wsmf)
	if fsyRg.CurTime() != wsmfRg.CurTime() {
		t.Fatalf("unsynchronized")
	}

	rp = Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true, Evaluate: true, NoACL: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 2836}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 2836}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
	rp = Synchronize(nil, fsy, "", wsmf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{}) {
		rp.TextOutput(os.Stdout)
		t.Fatal(rp.GetStats())
	}
}

func genArboNoFear(t *testing.T, dss, bck cabridss.Dss) (cabritbx.RandGen, []int64) {
	cts := make([]int64, 4)
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)

	if err := rg.Create(500); err != nil {
		t.Error(err)
	}
	cts[0] = rg.CurTime() + 3600
	if bck != nil {
		rp := Synchronize(nil, dss, "", bck, "0", SyncOptions{InDepth: true})
		if !(rp.GetStats() == SyncStats{CreNum: 1476, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}

	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Update(600); err != nil {
		t.Error(err)
	}
	cts[1] = rg.CurTime() + 3600
	if bck != nil {
		rp := Synchronize(nil, dss, "", bck, "1", SyncOptions{InDepth: true})
		if !(rp.GetStats() == SyncStats{CreNum: 1583, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}

	rg.AdvTime(3600 * 24 * 21)
	if err := rg.Update(700); err != nil {
		t.Error(err)
	}
	cts[2] = rg.CurTime() + 3600
	if bck != nil {
		rp := Synchronize(nil, dss, "", bck, "2", SyncOptions{InDepth: true})
		if !(rp.GetStats() == SyncStats{CreNum: 1671, UpdNum: 1}) {
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
	fsy, bck, olf, obs := createFsyOlf(t, tfs, true)
	fsyRg, fsyCts := genArboNoFear(t, fsy, bck)
	olfRg, olfCts := genArboNoFear(t, olf, nil)
	obsRg, obsCts := genArboNoFear(t, obs, nil)
	if fsyRg.CurTime() != olfRg.CurTime() || !(fsyCts[3] == olfCts[3]) {
		t.Fatalf("unsynchronized")
	}
	if fsyRg.CurTime() != obsRg.CurTime() || !(fsyCts[3] == obsCts[3]) {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 1812}) || len(rp.Entries) != 1812 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

	olf, err = cabridss.NewOlfDss(cabridss.OlfConfig{
		DssBaseConfig: cabridss.DssBaseConfig{LocalPath: ufpath.Join(tfs.Path(), "olf")},
		Root:          ufpath.Join(tfs.Path(), "olf")}, olfCts[0], nil)
	if err != nil {
		t.Error(err)
	}
	rp = Synchronize(nil, bck, "0", olf, "", SyncOptions{InDepth: true, Evaluate: true})

	if !(rp.GetStats() == SyncStats{MUpNum: 1477}) || len(rp.Entries) != 1477 {
		rp.TextOutput(os.Stderr)
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

	olf, err = cabridss.NewOlfDss(cabridss.OlfConfig{
		DssBaseConfig: cabridss.DssBaseConfig{LocalPath: ufpath.Join(tfs.Path(), "olf")},
		Root:          ufpath.Join(tfs.Path(), "olf")}, olfCts[1], nil)
	if err != nil {
		t.Error(err)
	}
	rp = Synchronize(nil, bck, "1", olf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 1584}) || len(rp.Entries) != 1584 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

	olf, err = cabridss.NewOlfDss(cabridss.OlfConfig{
		DssBaseConfig: cabridss.DssBaseConfig{LocalPath: ufpath.Join(tfs.Path(), "olf")},
		Root:          ufpath.Join(tfs.Path(), "olf")}, olfCts[2], nil)
	if err != nil {
		t.Error(err)
	}
	rp = Synchronize(nil, bck, "2", olf, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 1672}) || len(rp.Entries) != 1672 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 1812}) || len(rp.Entries) != 1812 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

	obs, err = cabridss.NewObsDss(getMockFsObsConfig(tfs), obsCts[0], nil)
	if err != nil {
		t.Error(err)
	}
	rp = Synchronize(nil, bck, "0", obs, "", SyncOptions{InDepth: true, Evaluate: true})

	if !(rp.GetStats() == SyncStats{MUpNum: 1477}) || len(rp.Entries) != 1477 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

	obs, err = cabridss.NewObsDss(getMockFsObsConfig(tfs), obsCts[1], nil)
	if err != nil {
		t.Error(err)
	}
	rp = Synchronize(nil, bck, "1", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 1584}) || len(rp.Entries) != 1584 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

	obs, err = cabridss.NewObsDss(getMockFsObsConfig(tfs), obsCts[2], nil)
	if err != nil {
		t.Error(err)
	}
	rp = Synchronize(nil, bck, "2", obs, "", SyncOptions{InDepth: true, Evaluate: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 1672}) || len(rp.Entries) != 1672 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}

}

func genArboBiDir(t *testing.T, dss, bck cabridss.Dss) []int64 {
	cts := make([]int64, 4)
	rg := cabritbx.NewRanGen(cabritbx.GetDefaultConfig(), dss)

	if err := rg.Create(500); err != nil {
		t.Error(err)
	}
	cts[0] = rg.CurTime() + 3600
	if bck != nil {
		rp := Synchronize(nil, dss, "", bck, "0", SyncOptions{InDepth: true})
		if !(rp.GetStats() == SyncStats{CreNum: 1476, UpdNum: 1}) {
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
		rp := Synchronize(nil, dss, "", bck, "1", SyncOptions{InDepth: true})
		if !(rp.GetStats() == SyncStats{CreNum: 1533, UpdNum: 1}) {
			t.Fatal(rp.GetStats())
		}
	}
	return cts
}

func TestSynchronizeArboBiDirOlf(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboBiDir", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, bck, olf, _ := createFsyOlf(t, tfs, true)

	fsyCts := genArboBiDir(t, fsy, bck)
	olfCts := genArboBiDir(t, olf, nil)
	if !(fsyCts[0] == olfCts[0]) {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{CreNum: 210, UpdNum: 321, RmvNum: 0, MUpNum: 1111}) || len(rp.Entries) != 1642 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{CreNum: 210, UpdNum: 321, RmvNum: 0, MUpNum: 1111}) || len(rp.Entries) != 1642 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}
	rp = Synchronize(nil, fsy, "", olf, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 257}) || len(rp.Entries) != 1642 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}
}

func TestSynchronizeArboBiDirObs(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestSynchronizeArboBiDirObs", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, bck, _, obs := createFsyOlf(t, tfs, true)

	fsyCts := genArboBiDir(t, fsy, bck)
	obsCts := genArboBiDir(t, obs, nil)
	if !(fsyCts[0] == obsCts[0]) {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{CreNum: 210, UpdNum: 321, RmvNum: 0, MUpNum: 1111}) || len(rp.Entries) != 1642 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{CreNum: 210, UpdNum: 321, RmvNum: 0, MUpNum: 1111}) || len(rp.Entries) != 1642 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 257}) || len(rp.Entries) != 1642 {
		t.Fatal(rp.GetStats(), len(rp.Entries))
	}
}

func runSynchronizeArboBiDirObs(t *testing.T) (int, SyncReport, SyncReport) {
	tfs, err := testfs.CreateFs("TestSynchronizeArboBiDirObs", nil)
	if err != nil {
		t.Error(err)
	}
	defer tfs.Delete()
	fsy, bck, _, obs := createFsyOlf(t, tfs, true)

	fsyCts := genArboBiDir(t, fsy, bck)
	obsCts := genArboBiDir(t, obs, nil)
	if !(fsyCts[0] == obsCts[0]) {
		t.Fatalf("unsynchronized")
	}

	rp := Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{CreNum: 210, UpdNum: 321, RmvNum: 0, MUpNum: 1111}) || len(rp.Entries) != 1642 {
		return 1, rp, SyncReport{}
	}
	rd := rp.GetRefDiag()
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, BiDir: true, RefDiag: &rd})
	if !(rp.GetStats() == SyncStats{CreNum: 210, UpdNum: 321, RmvNum: 0, MUpNum: 1111}) || len(rp.Entries) != 1642 {
		return 2, rp, SyncReport{}
	}
	rp2 := rp
	rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true})
	if !(rp.GetStats() == SyncStats{MUpNum: 257}) || len(rp.Entries) != 1642 {
		fmt.Fprintln(os.Stderr, "return 3, SyncReport{}, rp!!!")
		rd := rp.GetRefDiag()
		rp = Synchronize(nil, fsy, "", obs, "", SyncOptions{InDepth: true, Evaluate: true, BiDir: true, RefDiag: &rd})
		if !(rp.GetStats() == SyncStats{MUpNum: 257}) || len(rp.Entries) != 1642 {
			return 3, SyncReport{}, rp
		}
	}
	return 4, rp2, rp
}

func TestLoopSynchronizeArboBiDirObs(t *testing.T) {
	t.Skip("TestLoopSynchronizeArboBiDirObs currently unused")
	optionalSkip(t)
	var rp2ok, rp4ok, rp2nok, rp4nok SyncReport
	for {
		step, rp2, rp4 := runSynchronizeArboBiDirObs(t)
		if step == 1 {
			t.Fatalf("unattended %v", rp2)
		}
		if step == 2 {
			rp2nok = rp2
		}
		if step == 3 {
			rp4nok = rp4
		}
		if step == 4 {
			rp2ok, rp4ok = rp2, rp4
		}
		if len(rp2ok.Entries) > 0 && len(rp4ok.Entries) > 0 && len(rp2nok.Entries) > 0 && len(rp4nok.Entries) > 0 {
			break
		}
		fmt.Printf("%d %d %d %d %d\n", step, len(rp2ok.Entries), len(rp4ok.Entries), len(rp2nok.Entries), len(rp4nok.Entries))
	}
	fmt.Printf("%d %d %d %d\n", len(rp2ok.Entries), len(rp4ok.Entries), len(rp2nok.Entries), len(rp4nok.Entries))
	fmt.Println("RP2OK")
	rp2ok.SortByPath().TextOutput(os.Stdout)
	fmt.Println("RP2NOK")
	rp2nok.SortByPath().TextOutput(os.Stdout)
	fmt.Println("RP4OK")
	rp4ok.SortByPath().TextOutput(os.Stdout)
	fmt.Println("RP4NOK")
	rp4nok.SortByPath().TextOutput(os.Stdout)
}
