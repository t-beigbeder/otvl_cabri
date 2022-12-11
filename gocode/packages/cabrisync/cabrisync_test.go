package cabrisync

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/mockfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

var ucpCount = 0
var ids []cabridss.IdentityConfig
var currentUserConfig cabridss.UserConfig
var mtCount = time.Date(2018, time.April, 24, 23, 0, 0, 0, time.UTC).Unix() - 1

func mtimeCount() int64 { mtCount += 1; return mtCount }

func newUcp(tfs *testfs.Fs) (ucp string, uc cabridss.UserConfig, err error) {
	ucpCount += 1
	ucp = ufpath.Join(tfs.Path(), fmt.Sprintf(".cabri-i%d", ucpCount))
	if ucpCount == 1 {
		uc1, err1 := cabridss.GetUserConfig(cabridss.DssBaseConfig{}, ucp)
		if err1 != nil {
			return
		}
		ids = uc1.Identities
	}
	id, err := cabridss.GenIdentity(fmt.Sprintf("id-%d", ucpCount))
	ids = append(ids, id)
	for _, id = range ids {
		cabridss.UserConfigPutIdentity(cabridss.DssBaseConfig{}, ucp, id)
	}
	uc, _ = cabridss.GetUserConfig(cabridss.DssBaseConfig{}, ucp)
	return
}

func dumpIx(six, cix cabridss.Index) {
	if cix == nil {
		return
	}
	println("six")
	println(six.Dump())
	println("cix")
	println(cix.Dump())
	println()
}

func basicTfsStartup(tfs *testfs.Fs) error {
	if err := tfs.RandTextFile("a.txt", 41); err != nil {
		return err
	}
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
		return err
	}
	if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
		return err
	}
	if err := os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(ufpath.Join(tfs.Path(), "f", "sf"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(ufpath.Join(tfs.Path(), "g", "sg"), 0755); err != nil {
		return err
	}
	if err := tfs.RandTextFile("e/se/c3.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("e/se/c1.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("e/se/c2.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("f/sf/d4.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("g/sg/e5.txt", 20); err != nil {
		return err
	}
	return nil
}

func runTestSynchronizeBasic(t *testing.T, tfsl *testfs.Fs, dssl, dssr cabridss.Dss, noAcl bool, verbose bool) error {
	var err error

	optionalSleep(t)
	if err = dssr.Mkns("", time.Now().Unix(), []string{"step1/", "step2/", "step3/", "step4/"}, nil); err != nil {
		t.Fatalf("runTestSynchronizeBasic error %v", err)
	}
	if err = dssr.Mkns("step1", time.Now().Unix(), nil, nil); err != nil {
		t.Fatalf("runTestSynchronizeBasic error %v", err)
	}
	report1 := Synchronize(nil, dssl, "", dssr, "step1", SyncOptions{InDepth: true, Evaluate: true, NoACL: noAcl})
	report1.TextOutput(io.Discard)
	rs1 := report1.GetStats()
	if report1.HasErrors() || rs1.CreNum != 14 || rs1.UpdNum != 1 || rs1.MUpNum != 0 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs1)
	}

	report2 := Synchronize(nil, dssl, "", dssr, "step1", SyncOptions{InDepth: true, NoACL: noAcl})
	rs2 := report2.GetStats()
	if rs2.ErrNum != 0 || rs2.CreNum != 14 || rs2.UpdNum != 1 || rs2.MUpNum != 0 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs2)
	}
	if err = dssr.Mkns("step2", time.Now().Unix(), nil, nil); err != nil {
		t.Fatalf("runTestSynchronizeBasic error %v", err)
	}
	rs2bis := Synchronize(nil, dssl, "", dssr, "step2", SyncOptions{InDepth: true, NoACL: noAcl}).GetStats()
	if rs2bis.ErrNum != 0 || rs2bis.CreNum != 14 || rs2bis.UpdNum != 1 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs2bis)
	}
	if err = dssr.Mkns("step3", time.Now().Unix(), nil, nil); err != nil {
		t.Fatalf("runTestSynchronizeBasic error %v", err)
	}
	rs2ter := Synchronize(nil, dssl, "", dssr, "step3", SyncOptions{InDepth: true, NoACL: noAcl}).GetStats()
	if rs2ter.ErrNum != 0 || rs2ter.CreNum != 14 || rs2ter.UpdNum != 1 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs2ter)
	}
	if err = dssr.Mkns("step4", time.Now().Unix(), nil, nil); err != nil {
		t.Fatalf("runTestSynchronizeBasic error %v", err)
	}
	rs2d := Synchronize(nil, dssl, "", dssr, "step4", SyncOptions{InDepth: true, NoACL: noAcl}).GetStats()
	if rs2d.ErrNum != 0 || rs2d.CreNum != 14 || rs2d.UpdNum != 1 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs2d)
	}

	report3 := Synchronize(nil, dssl, "", dssr, "step1", SyncOptions{InDepth: true, Evaluate: true, NoACL: noAcl})
	report3.TextOutput(io.Discard)
	rs3 := report3.GetStats()
	if rs3.ErrNum != 0 || rs3.CreNum != 0 || rs3.UpdNum != 0 || rs3.MUpNum != 0 {
		report3.TextOutput(os.Stdout)
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs3)
	}

	if err := tfsl.RandTextFile("d/b.txt", 22); err != nil {
		t.Fatalf("runTestSynchronizeBasic failed %v", err)
	}
	if err := tfsl.RandTextFile("d/bc.txt", 23); err != nil {
		t.Fatalf("runTestSynchronizeBasic failed %v", err)
	}
	if err := tfsl.RandTextFile("e/se/c2.txt", 24); err != nil {
		t.Fatalf("runTestSynchronizeBasic failed %v", err)
	}
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	err = dssl.GetAfs().Chtimes(ufpath.Join(tfsl.Path(), "e/se/c2.txt"), time.Now(), time.Unix(ttt, 0))
	err = dssl.GetAfs().Chtimes(ufpath.Join(tfsl.Path(), "g/sg/e5.txt"), time.Now(), time.Unix(ttt, 0))
	err = dssl.GetAfs().Chtimes(ufpath.Join(tfsl.Path(), "g"), time.Now(), time.Unix(ttt, 0))
	err = dssl.GetAfs().RemoveAll(ufpath.Join(tfsl.Path(), "f"))
	optionalSleep(t)

	report4 := Synchronize(nil, dssl, "", dssr, "step2", SyncOptions{InDepth: true, Evaluate: true, NoACL: noAcl})
	report4.TextOutput(io.Discard)
	rs4 := report4.GetStats()
	if rs4.ErrNum != 0 || rs4.CreNum != 1 || rs4.UpdNum != 4 || rs4.RmvNum != 3 || rs4.KeptNum != 0 || rs4.MUpNum != 2 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs4)
	}
	report5 := Synchronize(nil, dssl, "", dssr, "step2", SyncOptions{InDepth: true, NoACL: noAcl})
	report5.TextOutput(io.Discard)
	rs5 := report5.GetStats()
	if rs5.ErrNum != 0 || rs5.CreNum != 1 || rs5.UpdNum != 4 || rs5.RmvNum != 3 || rs5.KeptNum != 0 || rs5.MUpNum != 2 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs5)
	}
	report6 := Synchronize(nil, dssl, "", dssr, "step2", SyncOptions{InDepth: true, Evaluate: true, NoACL: noAcl})
	report6.TextOutput(io.Discard)
	rs6 := report6.GetStats()
	if rs6.ErrNum != 0 || rs6.CreNum != 0 || rs6.UpdNum != 0 || rs6.RmvNum != 0 || rs6.KeptNum != 0 || rs6.MUpNum != 0 {
		report6.TextOutput(os.Stdout)
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs6)
	}

	report7 := Synchronize(nil, dssl, "", dssr, "step3", SyncOptions{InDepth: true, Evaluate: true, KeepContent: true, NoACL: noAcl})
	report7.TextOutput(io.Discard)
	rs7 := report7.GetStats()
	if rs7.ErrNum != 0 || rs7.CreNum != 1 || rs7.UpdNum != 4 || rs7.RmvNum != 0 || rs7.KeptNum != 3 || rs7.MUpNum != 2 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs7)
	}
	report8 := Synchronize(nil, dssl, "", dssr, "step3", SyncOptions{InDepth: true, KeepContent: true, NoACL: noAcl})
	report8.TextOutput(io.Discard)
	rs8 := report8.GetStats()
	if rs8.ErrNum != 0 || rs8.CreNum != 1 || rs8.UpdNum != 4 || rs8.RmvNum != 0 || rs8.KeptNum != 3 || rs8.MUpNum != 2 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs8)
	}
	report9 := Synchronize(nil, dssl, "", dssr, "step3", SyncOptions{InDepth: true, KeepContent: true, Evaluate: true, NoACL: noAcl})
	report9.TextOutput(io.Discard)
	rs9 := report9.GetStats()
	if rs9.ErrNum != 0 || rs9.CreNum != 0 || rs9.UpdNum != 1 || rs9.RmvNum != 0 || rs9.KeptNum != 3 || rs9.MUpNum != 0 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs9)
	}

	report10 := Synchronize(nil, dssl, "", dssr, "step4", SyncOptions{InDepth: true, BiDir: true, Evaluate: true, NoACL: noAcl})
	report10.TextOutput(io.Discard)
	rs10 := report10.GetStats()
	if rs10.ErrNum != 0 || rs10.CreNum != 4 || rs10.UpdNum != 4 || rs10.RmvNum != 0 || rs10.KeptNum != 0 || rs10.MUpNum != 2 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs10)
	}
	report11 := Synchronize(nil, dssl, "", dssr, "step4", SyncOptions{InDepth: true, BiDir: true, NoACL: noAcl})
	report11.TextOutput(io.Discard)
	rs11 := report11.GetStats()
	if rs11.ErrNum != 0 || rs11.CreNum != 4 || rs11.UpdNum != 4 || rs11.RmvNum != 0 || rs11.KeptNum != 0 || rs11.MUpNum != 2 {
		println("report10")
		report10.TextOutput(os.Stdout)
		println("report11")
		report11.TextOutput(os.Stdout)
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs11)
	}
	report12 := Synchronize(nil, dssl, "", dssr, "step4", SyncOptions{InDepth: true, BiDir: true, Evaluate: true, NoACL: noAcl})
	report12.TextOutput(io.Discard)
	rs12 := report12.GetStats()
	if rs12.ErrNum != 0 || rs12.CreNum != 0 || rs12.UpdNum != 0 || rs12.RmvNum != 0 || rs12.KeptNum != 0 || rs12.MUpNum != 0 {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs12)
	}

	beVerbose := func(level int, line string) {
		if !verbose || level > 3 {
			return
		}
		fmt.Fprintln(os.Stderr, line)
	}
	report13 := Synchronize(nil, dssl, "", dssr, "step4", SyncOptions{InDepth: true, BiDir: true, Evaluate: true, NoACL: noAcl, BeVerbose: beVerbose})
	es := SyncStats{}
	rs13 := report13.GetStats()
	if rs13 != es {
		t.Fatalf("runTestSynchronizeBasic failed %+v", rs13)
	}
	hdss, _ := dssr.(cabridss.HDss)
	six, cix := cabridss.GetServerIndexesForTests(hdss)
	dumpIx(six, cix)
	return nil
}

func TestSynchronizeBasic(t *testing.T) {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSynchronizeBasicLeft", basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsr, err := testfs.CreateFs("TestSynchronizeBasicRight", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsr.Delete()
	dssr, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsr.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	dssr.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	runTestSynchronizeBasic(t, tfsl, dssl, dssr, true, true)
}

func TestSynchronizeBasicACL(t *testing.T) {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSynchronizeBasicACLLeft", basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsr, err := testfs.CreateFs("TestSynchronizeBasicACLRight", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsr.Delete()
	dssr, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsr.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	dssr.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	runTestSynchronizeBasic(t, tfsl, dssl, dssr, false, false)
}

func TestSynchronizeBasicFsyOlf(t *testing.T) {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSyncBasicFsyOlfLeft", basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsr, err := testfs.CreateFs("TestSyncBasicFsyOlfRight", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsr.Delete()
	dssr, err := cabridss.CreateOlfDss(cabridss.OlfConfig{DssBaseConfig: cabridss.DssBaseConfig{LocalPath: tfsr.Path()}, Root: tfsr.Path(), Size: "s"})
	if err != nil {
		t.Fatal(err.Error())
	}
	dssr.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	runTestSynchronizeBasic(t, tfsl, dssl, dssr, true, false)
}

func TestSynchronizeBasicFsyOlfACL(t *testing.T) {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSynchronizeBasicFsyOlfACLLeft", basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsr, err := testfs.CreateFs("TestSynchronizeBasicFsyOlfACLRight", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsr.Delete()
	dssr, err := cabridss.CreateOlfDss(cabridss.OlfConfig{DssBaseConfig: cabridss.DssBaseConfig{LocalPath: tfsr.Path()}, Root: tfsr.Path(), Size: "s"})
	if err != nil {
		t.Fatal(err.Error())
	}
	dssr.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	runTestSynchronizeBasic(t, tfsl, dssl, dssr, false, false)
}

func TestSynchronizeBasicFsyObs(t *testing.T) {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSynchronizeBasicFsyObsLeft", basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsm, err := testfs.CreateFs("TestSynchronizeBasicFsyObsMock", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsm.Delete()

	config := getOC()

	cabridss.CleanObsDss(getOC())
	config.GetS3Session = func() cabridss.IS3Session {
		return cabridss.NewS3Session(config, func(parent cabridss.IS3Session) cabridss.IS3Session {
			return cabridss.NewS3sMockFs(tfsm.Path(), func(parent cabridss.IS3Session) cabridss.IS3Session {
				return cabridss.NewS3sMockTests(parent, func(args ...any) interface{} {
					if len(args) > 2 {
						fmt.Printf("%s S3sMockTests %v %v\n", time.Now().Format("2006-01-02 15:04:05.000"), args[1], args[2])
					} else {
						fmt.Printf("%s S3sMockTests %v\n", time.Now().Format("2006-01-02 15:04:05.000"), args[1])
					}
					return nil
				})
			})
		})
	}
	config.GetS3Session = func() cabridss.IS3Session {
		return cabridss.NewS3sMockFs(tfsm.Path(), nil)
	}
	//config.GetS3Session = nil
	dssr, err := cabridss.NewObsDss(config, 0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	runTestSynchronizeBasic(t, tfsl, dssl, dssr, true, false)
}

func TestSynchronizeBasicFsyObsACL(t *testing.T) {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSynchronizeBasicFsyObsACLLeft", basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsm, err := testfs.CreateFs("TestSynchronizeBasicFsyObsACLMock", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsm.Delete()

	config := getOC()

	cabridss.CleanObsDss(getOC())
	config.GetS3Session = func() cabridss.IS3Session {
		return cabridss.NewS3Session(config, func(parent cabridss.IS3Session) cabridss.IS3Session {
			return cabridss.NewS3sMockFs(tfsm.Path(), func(parent cabridss.IS3Session) cabridss.IS3Session {
				return cabridss.NewS3sMockTests(parent, func(args ...any) interface{} {
					fmt.Printf("%s S3sMockTests %v\n", time.Now().Format("2006-01-02 15:04:05.000"), args[1])
					return nil
				})
			})
		})
	}
	config.GetS3Session = func() cabridss.IS3Session {
		return cabridss.NewS3sMockFs(tfsm.Path(), nil)
	}
	//config.GetS3Session = nil
	dssr, err := cabridss.NewObsDss(config, 0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	runTestSynchronizeBasic(t, tfsl, dssl, dssr, false, false)
}

func TestSynchronizeBasicFsyWebOlf(t *testing.T) {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSynchronizeBasicFsyWebOlfLeft", basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsr, err := testfs.CreateFs("TestSynchronizeBasicFsyWebOlfRight", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsr.Delete()

	getPIndex := func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
		return cabridss.NewPIndex(ufpath.Join(tfsr.Path(), "index.bdb"), false, false)
	}

	sv, err := createWebDssServer(":3000", "",
		cabridss.CreateNewParams{Create: true, DssType: "olf", Root: tfsr.Path(), Size: "s", GetIndex: getPIndex},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sv.Shutdown()

	dssr, err := cabridss.NewWebDss(
		cabridss.WebDssConfig{
			DssBaseConfig: cabridss.DssBaseConfig{
				ConfigDir: ufpath.Join(tfsr.Path(), ".cabri"),
				WebPort:   "3000",
			}, NoClientLimit: true},
		0, nil)
	if err != nil {
		t.Fatal(err)
	}

	//dssr, err := cabridss.CreateOlfDss(cabridss.OlfConfig{DssBaseConfig: cabridss.DssBaseConfig{LocalPath: tfsr.Path()}, Root: tfsr.Path(), Size: "s"})
	//if err != nil {
	//	t.Fatal(err.Error())
	//}
	//dssr.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	runTestSynchronizeBasic(t, tfsl, dssl, dssr, true, false)
}

func runTestSynchronizeWith(t *testing.T, createDssCb func(*testfs.Fs) error, newDssCb func(*testfs.Fs) (cabridss.HDss, error)) error {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs(fmt.Sprintf("%sLeft", t.Name()), basicTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsr, err := testfs.CreateFs(fmt.Sprintf("%sRight", t.Name()), basicTfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfsr.Delete()
	if err = createDssCb(tfsr); err != nil {
		return err
	}
	dssr, err := newDssCb(tfsr)
	if err != nil {
		t.Fatal(err)
	}
	defer dssr.Close()
	return runTestSynchronizeBasic(t, tfsl, dssl, dssr, true, false)
}

func TestSynchronizeBasicFsyEDssApiOlf(t *testing.T) {
	if err := runTestSynchronizeWith(t,
		func(tfs *testfs.Fs) error {
			_, err := cabridss.CreateOlfDss(cabridss.OlfConfig{
				DssBaseConfig: cabridss.DssBaseConfig{LocalPath: tfs.Path(), Encrypted: true},
				Root:          tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (cabridss.HDss, error) {
			ucp, uc, _ := newUcp(tfs)
			dss, err := cabridss.NewEDss(
				cabridss.EDssConfig{
					WebDssConfig: cabridss.WebDssConfig{
						DssBaseConfig: cabridss.DssBaseConfig{
							LibApi:    true,
							ConfigDir: ucp,
						},
						LibApiDssConfig: cabridss.LibApiDssConfig{
							IsOlf: true,
							OlfCfg: cabridss.OlfConfig{
								DssBaseConfig: cabridss.DssBaseConfig{
									LocalPath: tfs.Path(),
									GetIndex: func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
										return cabridss.NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
									},
								}, Root: tfs.Path(), Size: "s"},
						},
					},
				},
				0, cabridss.IdPkeys(uc))
			return dss, err

		}); err != nil {
		t.Fatal(err)
	}
}

func TestSynchronizeBasicFsyEDssWebOlf(t *testing.T) {
	var sv cabridss.WebServer
	var err error
	defer func() {
		if sv != nil {
			sv.Shutdown()
		}
	}()

	if err := runTestSynchronizeWith(t,
		func(tfs *testfs.Fs) error {
			getPIndex := func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
				return cabridss.NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
			}
			sv, err = createWebDssServer(":3000", "",
				cabridss.CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex, Encrypted: true},
			)
			return err
		},
		func(tfs *testfs.Fs) (cabridss.HDss, error) {
			dss, err := cabridss.NewEDss(
				cabridss.EDssConfig{
					WebDssConfig: cabridss.WebDssConfig{
						DssBaseConfig: cabridss.DssBaseConfig{
							ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
							WebPort:   "3000",
						}, NoClientLimit: true},
				},
				0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func runTestSynchronizeBasicFsyEDssApiObs(t *testing.T) error {
	return runTestSynchronizeWith(t,
		func(tfs *testfs.Fs) error {
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = cabridss.GetPIndex
			config.Encrypted = true
			dss, err := cabridss.CreateObsDss(config)
			if err != nil {
				return err
			}
			dss.Close()
			return nil
		},
		func(tfs *testfs.Fs) (cabridss.HDss, error) {
			dbc := getOC()
			dbc.LocalPath = tfs.Path()
			dbc.DssBaseConfig.GetIndex = cabridss.GetPIndex
			dbc.Encrypted = true
			dss, err := cabridss.NewEDss(
				cabridss.EDssConfig{
					cabridss.WebDssConfig{
						DssBaseConfig: cabridss.DssBaseConfig{
							LibApi:    true,
							ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
						},
						LibApiDssConfig: cabridss.LibApiDssConfig{
							IsObs:  true,
							ObsCfg: dbc,
						},
					},
				},
				0, nil)
			return dss, err
		})
}

func TestSynchronizeBasicFsyEDssApiObs(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestSynchronizeBasicFsyEDssApiObs(t)
	})
}

func basicPlusTfsStartup(tfs *testfs.Fs) error {
	if err := tfs.RandTextFile("a.txt", 41); err != nil {
		return err
	}
	if err := tfs.RandTextFile("b.txt", 412); err != nil {
		return err
	}
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
		return err
	}
	if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
		return err
	}
	if err := os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(ufpath.Join(tfs.Path(), "f", "sf"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(ufpath.Join(tfs.Path(), "g", "sg"), 0755); err != nil {
		return err
	}
	if err := tfs.RandTextFile("e/se/c3.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("e/se/c1.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("e/se/c2.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("f/sf/d4.txt", 20); err != nil {
		return err
	}
	if err := tfs.RandTextFile("g/sg/e5.txt", 20); err != nil {
		return err
	}
	return nil
}

func runTestSynchroInconsistentChildren(t *testing.T) error {
	optionalSkip(t)
	tfsl, err := testfs.CreateFs("TestSynchroInconsistentChildrenFsy", basicPlusTfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsl.Delete()
	dssl, err := cabridss.NewFsyDss(cabridss.FsyConfig{}, tfsl.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dssl.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	tfsm, err := testfs.CreateFs("TestSynchroInconsistentChildrenSmf", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfsm.Delete()

	config := getOC()
	config.GetS3Session = func() cabridss.IS3Session {
		return cabridss.NewS3sMockFs(tfsm.Path(), nil)
	}
	dssr, err := cabridss.NewObsDss(config, 0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err = dssr.Mkns("", time.Now().Unix(), nil, nil); err != nil {
		t.Fatal(err)
	}
	rp1 := Synchronize(nil, dssl, "", dssr, "", SyncOptions{})
	if rp1.HasErrors() || len(rp1.Entries) != 3 {
		rp1.TextOutput(os.Stdout)
		t.Fatalf("%d", len(rp1.Entries))
	}
	des, _ := os.ReadDir(tfsm.Path())
	for _, de := range des {
		if strings.HasPrefix(de.Name(), "meta-ffa0da5d885fba09d903c782713b6b09") { // b.txt
			os.Remove(ufpath.Join(tfsm.Path(), de.Name()))
		}
	}
	rp2r := Synchronize(nil, dssl, "", dssr, "", SyncOptions{InDepth: true, Evaluate: true})
	if rp2r.HasErrors() || len(rp2r.Entries) != 16 {
		rp2r.TextOutput(os.Stdout)
		return fmt.Errorf("evaluate %d", len(rp2r.Entries))
	}
	rp2 := Synchronize(nil, dssl, "", dssr, "", SyncOptions{InDepth: true})
	if rp2.HasErrors() || len(rp2.Entries) != 16 {
		rp2.TextOutput(os.Stdout)
		return fmt.Errorf("synchronize %d", len(rp2.Entries))
	}
	rp3 := Synchronize(nil, dssl, "", dssr, "", SyncOptions{InDepth: true, Evaluate: true})
	s3 := rp3.GetStats()
	if rp3.HasErrors() || len(rp3.Entries) != 16 || s3 != (SyncStats{}) {
		rp3.TextOutput(os.Stdout)
		return fmt.Errorf("reevaluate %d %v", len(rp3.Entries), s3)
	}
	return nil
}

func TestSynchroInconsistentChildren(t *testing.T) {
	optionalSkip(t)
	if err := runTestSynchroInconsistentChildren(t); err != nil {
		t.Fatal(err)
	}
}

func TestLoopSynchroInconsistentChildren(t *testing.T) {
	optionalSkip(t)
	for i := 0; i < 10; i++ {
		if err := runTestSynchroInconsistentChildren(t); err != nil {
			t.Fatalf("round #%d: %v", i, err)
		}
	}
}
