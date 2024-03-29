package cabridss

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/mockfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"os"
	"runtime"
	"testing"
)

func doTestMkallNs(dss Dss, t *testing.T, isFsy bool) {
	var err error
	if err = MkallNs(dss, "", 0); err != nil {
		t.Fatal(err)
	}
	if err = MkallNs(dss, "a", 0); err != nil {
		t.Fatal(err)
	}
	_, err = dss.Lsns("a")
	if err != nil {
		t.Fatal(err)
	}
	if err = MkallNs(dss, "a", 0); err != nil {
		t.Fatal(err)
	}
	if err = MkallNs(dss, "dd", 0); err == nil && isFsy {
		t.Fatal("should fail dd not a dir")
	}
	if err = MkallNs(dss, "a/b1", 0); err != nil {
		t.Fatal(err)
	}
	if err = MkallNs(dss, "a/b2", 0); err != nil {
		t.Fatal(err)
	}
	if err = MkallNs(dss, "a/b3/c3/d3", 0); err != nil {
		t.Fatal(err)
	}
	cs, err := dss.Lsns("a")
	if err != nil || len(cs) != 3 {
		t.Fatalf("%v %d", err, len(cs))
	}
	cs, err = dss.Lsns("a/b3")
	if err != nil || len(cs) != 1 {
		t.Fatalf("%v %d", err, len(cs))
	}

}

func TestMkallNsFsy(t *testing.T) {
	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("dd", 20); err != nil {
			return err
		}
		return nil
	}

	tfs, err := testfs.CreateFs("TestMkallNs", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), nil))
	doTestMkallNs(dss, t, true)
}

func TestMkallNsOlf(t *testing.T) {
	tfs, err := testfs.CreateFs("TestMkallNs", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "l"})
	if err != nil {
		t.Fatal(err.Error())
	}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), nil))
	if err = dss.Mkns("", 0, nil, nil); err != nil {
		t.Error(err)
	}
	doTestMkallNs(dss, t, false)
}

func doTestMkallContent(dss Dss, t *testing.T) {
	var err error
	if err = MkallContent(dss, "0.c", 0); err != nil {
		t.Fatal(err)
	}
	if err = MkallNs(dss, "a/1.c", 0); err != nil {
		t.Fatal(err)
	}

}

func TestMkallContentFsy(t *testing.T) {
	tfs, err := testfs.CreateFs("TestMkallContent", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), nil))
	doTestMkallContent(dss, t)
}

func TestErrorCollector(t *testing.T) {
	errs := &ErrorCollector{}
	errs.Collect(fmt.Errorf("error #1 is %v", fmt.Errorf("embeded 1")))
	errs.Collect(fmt.Errorf("error #2 is %v", fmt.Errorf("embeded 2")))
	errHat := fmt.Errorf("the hat is %w", errs)
	fmt.Printf("errHat is %s\n", errHat)
	ue := errors.Unwrap(errHat)
	fmt.Printf("ue is %s\n", ue)
	tue := ue.(*ErrorCollector)
	fmt.Printf("tue is %s\n", tue)
}

type bfcl struct {
	bytes.Buffer
}

func (b bfcl) Close() error {
	return nil
}

func TestNewWriteCloserWithCb(t *testing.T) {
	bsa := bfcl{}
	wcwc := NewWriteCloserWithCb(&bsa, func(err error, size int64, ch string, wcwc *WriteCloserWithCb) error {
		if err == nil && size == 24 && ch == "6b33a33017f120c522a983001abf6967" {
			return nil
		}
		return fmt.Errorf("in TestNewWriteCloserWithCb cb: %v %d, %s", err, size, ch)
	})
	if _, err := wcwc.Write([]byte("TestNewWriteCloserWithCb")); err != nil {
		t.Fatal(err)
	}
	if err := wcwc.Close(); err != nil {
		t.Fatal(err)
	}
	bsa = bfcl{}
	wcwc = NewWriteCloserWithCb(&bsa, func(err error, size int64, ch string, wcwc *WriteCloserWithCb) error {
		return fmt.Errorf("in TestNewWriteCloserWithCb cb: %v", err)
	})
	if _, err := wcwc.Write([]byte("TestNewWriteCloserWithCb")); err != nil {
		t.Fatal(err)
	}
	var err error
	if err = wcwc.Close(); err == nil {
		t.Fatal("should fail with error")
	}
	wcwc, err = NewTempFileWriteCloserWithCb(appFs, os.TempDir(), "ntfwcwc", func(err error, size int64, ch string, me *WriteCloserWithCb) error {
		if err != nil || size != 24 || ch != "6b33a33017f120c522a983001abf6967" {
			return fmt.Errorf("in TestNewWriteCloserWithCb cb: %v %d, %s", err, size, ch)
		}
		if st, err := appFs.Stat(me.Underlying.(afero.File).Name()); err != nil || st.Size() != 24 {
			return fmt.Errorf("in TestNewWriteCloserWithCb cb: %+v %v", st, err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wcwc.Write([]byte("TestNewWriteCloserWithCb")); err != nil {
		t.Fatal(err)
	}
	if err := wcwc.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewReadCloserWithCb(t *testing.T) {
	bf := bytes.Buffer{}
	bf.Write([]byte("TestNewReadCloserWithCb"))
	rcwc, err := NewReadCloserWithCb(&bf, func() error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = rcwc.Close(); err != nil {
		t.Fatal(err)
	}
}

func optionalSkip(t *testing.T) {
	if runtime.GOOS == "windows" {
		if t.Name() == "InTheBeginning" ||
			t.Name() == "TestMetaBasic" ||
			t.Name() == "TestNewFsyDssOk" ||
			t.Name() == "TestNewFsyDssErr" ||
			t.Name() == "TestIrregular" ||
			t.Name() == "TestFsyStat" ||
			t.Name() == "TheEnd" {
			t.Skip(fmt.Sprintf("Skipping %s because it won't work on windows", t.Name()))
		}
	}
	if os.Getenv("CABRIDSS_SKIP_DEV_TESTS") != "" {
		if t.Name() == "InTheBeginning" ||
			t.Name() == "TestNewFsyDssOk" ||
			t.Name() == "TestNewFsyDssErr" ||
			t.Name() == "TestNewFsyDssBase" ||
			t.Name() == "TestNewFsyDssOsErrors" ||
			t.Name() == "TestFsyDssLsnsBase" ||
			t.Name() == "TestFsyDssLsnsErr" ||
			t.Name() == "TestFsyDssGetContentWriterBase" ||
			t.Name() == "TestFsyDssMtime" ||
			t.Name() == "TestFsyDssGetContentReaderBase" ||
			t.Name() == "TestFsyDssOsErrors" ||
			t.Name() == "TestFsyDssRemoveBase" ||
			t.Name() == "TestFsyDssGetMetaBasic" ||
			t.Name() == "TestFsyDssUpdateNsBasic" ||
			t.Name() == "TestIrregular" ||
			t.Name() == "TestFsyStat1" ||
			t.Name() == "TestSetSysAcl" ||
			t.Name() == "TestFsyDssRed" ||
			t.Name() == "InTheBeginning" ||
			t.Name() == "TestNewObsDssBase" ||
			t.Name() == "TestNewObsDssMockFsBase" ||
			t.Name() == "TestNewObsDssMockFsUnlock" ||
			t.Name() == "TestNewObsDssS3Mock" ||
			t.Name() == "TestCleanObsDss" ||
			t.Name() == "TestNewObsDssMindex" ||
			t.Name() == "TestNewObsDssPindex" ||
			t.Name() == "TestMockFsHistory" ||
			t.Name() == "TestObsHistory" ||
			t.Name() == "TestObsRedHistory" ||
			t.Name() == "TestObsMultiHistory" ||
			t.Name() == "TestCreateOlfDssErr" ||
			t.Name() == "TestNewOlfSmallDssOk1" ||
			t.Name() == "TestNewOlfSmallDssRedOk" ||
			t.Name() == "TestNewOlfMediumDssOk" ||
			t.Name() == "TestNewOlfLargeDssOk" ||
			t.Name() == "TestOlfDssMindex" ||
			t.Name() == "TestOlfHistory" ||
			t.Name() == "TestOlfRedHistory" ||
			t.Name() == "TestOlfMultiHistory" ||
			t.Name() == "TestNewWebDssServer" ||
			t.Name() == "TestNewWebDssTlsServer" ||
			t.Name() == "TestWebDssStoreMeta" ||
			t.Name() == "TestNewWebDssClientOlf" ||
			t.Name() == "TestNewWebDssTlsClientOlf" ||
			t.Name() == "TestNewWebDssClientOlfRed" ||
			t.Name() == "TestNewWebDssClientObs" ||
			t.Name() == "TestNewWebDssClientSmf" ||
			t.Name() == "TestNewWebDssApiClientOlf" ||
			t.Name() == "TestNewWebDssApiClientOlfRed" ||
			t.Name() == "TestNewWebDssApiClientObs" ||
			t.Name() == "TestNewWebDssApiClientSmf" ||
			t.Name() == "TestWebClientOlfHistory1" ||
			t.Name() == "TestWebClientObsHistory" ||
			t.Name() == "TestWebDssApiClientOlfHistory1" ||
			t.Name() == "TestWebDssApiClientObsHistory" ||
			t.Name() == "TestWebClientOlfMultiHistory" ||
			t.Name() == "TestNewWebApiClient" ||
			t.Name() == "TestWebApiStream" ||
			t.Name() == "TestWebTestSleep" ||
			t.Name() == "TestEDssClientOlfBase" ||
			t.Name() == "TestEDssClientOlfBaseRed" ||
			t.Name() == "TestEDssClientObsBase" ||
			t.Name() == "TestEDssClientSmfBase" ||
			t.Name() == "TestEDssApiClientOlfBase" ||
			t.Name() == "TestEDssApiClientOlfBaseRed" ||
			t.Name() == "TestEDssApiClientObsBase" ||
			t.Name() == "TestEDssApiClientSmfBase" ||
			t.Name() == "TestEDssClientOlfHistory" ||
			t.Name() == "TestEDssClientOlfRedHistory" ||
			t.Name() == "TestEDssApiClientOlfHistory" ||
			t.Name() == "TestEDssApiClientOlfRedHistory" ||
			t.Name() == "TestEDssClientOlfMultiHistory" ||
			t.Name() == "TestEDssApiClientOlfMultiHistory" ||
			t.Name() == "TheEnd" {
			t.Skip(fmt.Sprintf("Skipping %s because you set CABRIDSS_SKIP_DEV_TESTS", t.Name()))
		}
	}
}
