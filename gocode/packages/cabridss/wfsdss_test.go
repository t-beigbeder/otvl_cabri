package cabridss

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func createWfsDssServer(tfs *testfs.Fs, addr, root string) (WebServer, error) {
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		return nil, fmt.Errorf("createWfsDssServer failed with error %v", err)
	}
	httpConfig := WebServerConfig{Addr: addr, HasLog: true}
	if strings.Contains(addr, "443") {
		httpConfig.IsTls = true
		httpConfig.TlsCert = "cert.pem"
		httpConfig.TlsKey = "key.pem"
		httpConfig.BasicAuthUser = "user"
		httpConfig.BasicAuthPassword = "passw0rd"
	}
	return NewWfsDssServer(root, WfsDssServerConfig{
		WebServerConfig: httpConfig,
		Dss:             dss,
	})
}

func TestNewWfsDssServer(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestNewWfsDssServer", tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	sv, err := createWfsDssServer(tfs, ":3000", "")
	if err != nil {
		t.Fatal(err)
	}
	sv.Shutdown()
}

func TestNewWfsDssTlsServer(t *testing.T) {
	optionalSkip(t)
	if os.Getenv("CABRIDSS_KEEP_DEV_TESTS") == "" {
		t.Skip(fmt.Sprintf("Skipping %s because you didn't set CABRIDSS_KEEP_DEV_TESTS", t.Name()))
	}
	tfs, err := testfs.CreateFs("TestNewWfsDssServer", tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	sv, err := createWfsDssServer(tfs, "localhost:3443", "")
	if err != nil {
		t.Fatal(err)
	}
	sv.Shutdown()
}

func runWfsDssTestWithReducer(t *testing.T, doIt func(*testfs.Fs, Dss) error, redLimit int) error {
	optionalSkip(t)
	tfs, err := testfs.CreateFs(fmt.Sprintf("%s-%d", t.Name(), redLimit), tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	sv, err := createWfsDssServer(tfs, ":3000", "")
	if err != nil {
		t.Fatal(err)
	}
	defer sv.Shutdown()
	dss, err := NewWfsDss(WfsDssConfig{
		DssBaseConfig: DssBaseConfig{WebPort: "3000", ReducerLimit: redLimit},
		NoClientLimit: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer dss.Close()
	err = doIt(tfs, dss)
	return err
}

func runWfsDssTest(t *testing.T, doIt func(*testfs.Fs, Dss) error) error {
	if err := runWfsDssTestWithReducer(t, doIt, 0); err != nil {
		return err
	}
	return runWfsDssTestWithReducer(t, doIt, 2)
}

func TestNewWfsDssClient(t *testing.T) {
	err := runWfsDssTest(t, func(_ *testfs.Fs, dss Dss) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = runWfsDssTest(t, func(_ *testfs.Fs, dss Dss) error {
		return fmt.Errorf("bad")
	})
	if err == nil {
		t.Fatal("should fail with error")
	}
}

func TestNewWfsDssBase(t *testing.T) {
	err := runWfsDssTest(t, func(_ *testfs.Fs, dss Dss) error {
		err := dss.Updatens("", time.Now().Unix(), []string{"d"}, nil)
		if err == nil {
			return fmt.Errorf("Mkns should fail with error mkdir file exists")
		}
		err = dss.Updatens("/d", time.Now().Unix(), []string{"d2"}, nil)
		if err == nil {
			return fmt.Errorf("Mkns should fail with error namespace / (leading)")
		}
		err = dss.Updatens("d/", time.Now().Unix(), []string{"d2"}, nil)
		if err == nil {
			return fmt.Errorf("Mkns should fail with error namespace / (trailing)")
		}
		err = dss.Updatens("", time.Now().Unix(), []string{"/d2/", "d2\n/f.txt", "", "f1", "f2", "f3", "f1", "f3"}, nil)
		if err == nil || !strings.Contains(err.Error(), "name(s) [/d2/ d2\n/f.txt  f1 f3] should") {
			return fmt.Errorf("Mkns should fail with name check errors")
		}
		err = dss.Updatens("", time.Now().Unix(), []string{"d2/"}, nil)
		if err != nil {
			return fmt.Errorf("TestNewFsyDssBase failed with error %v", err)
		}
		err = dss.Mkns("d2", time.Now().Unix(), []string{"d3/", "f4"}, nil)
		if err != nil {
			return fmt.Errorf("TestNewFsyDssBase failed with error %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWfsDssLsnsBase(t *testing.T) {
	err := runWfsDssTest(t, func(_ *testfs.Fs, dss Dss) error {
		err := dss.Mkns("", time.Now().Unix(), []string{"d2/"}, nil)
		if err == nil {
			return fmt.Errorf("TestWfsDssLsnsBase should fail Mkns cannot be used non empty dir")
		}
		err = dss.Updatens("", time.Now().Unix(), []string{"d2/"}, nil)
		if err != nil {
			return fmt.Errorf("TestWfsDssLsnsBase failed with error %v", err)
		}
		err = dss.Mkns("d2", time.Now().Unix(), []string{"d3/", "f4"}, nil)
		if err != nil {
			return fmt.Errorf("TestWfsDssLsnsBase failed with error %v", err)
		}
		err = dss.Mkns("d2/d3", time.Now().Unix(), []string{"d4a/", "f5", "d4b"}, nil)
		if err != nil {
			return fmt.Errorf("TestWfsDssLsnsBase failed with error %v", err)
		}
		children0, err := dss.Lsns("")
		if err != nil || len(children0) != 1 || children0[0] != "d2/" {
			return fmt.Errorf("TestWfsDssLsnsBase failed with error %v or children %v", err, children0)
		}
		children2, err := dss.Lsns("d2")
		if err != nil || len(children2) != 2 {
			return fmt.Errorf("TestWfsDssLsnsBase failed with error %v or children %v", err, children2)
		}
		children3, err := dss.Lsns("d2/d3")
		if err != nil || len(children3) != 3 {
			return fmt.Errorf("TestWfsDssLsnsBase failed with error %v or children %v", err, children3)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

}

func TestWfsDssGetContentWriterBase(t *testing.T) {
	err := runWfsDssTest(t, func(tfs *testfs.Fs, dss Dss) error {
		fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		defer fi.Close()
		var iErr error
		fo, err := dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, func(err error, size int64, ch string) {
			if err != nil {
				iErr = fmt.Errorf(err.Error())
			}
			if size != 241 {
				iErr = fmt.Errorf("TestWfsDssGetContentWriterBase size %d != 241", size)
			}
			if ch != "484f617a695613aac4b346237aa01548" {
				iErr = fmt.Errorf("TestWfsDssGetContentWriterBase hash %s != %s", ch, "484f617a695613aac4b346237aa01548")
			}
		})
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		l, err := io.Copy(fo, fi)
		if err != nil {
			fo.Close()
			return fmt.Errorf("TestWfsDssGetContentWriterBase Copy error %v", err)
		}
		err = fo.Close()
		if err != nil {
			return fmt.Errorf("TestWfsDssGetContentWriterBase Close error %v", err)
		}
		if iErr != nil {
			return fmt.Errorf(iErr.Error())
		}
		_ = l
		fi2, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		defer fi2.Close()
		fo, err = dss.GetContentWriter("/no", time.Now().Unix(), nil, nil)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		l, err = io.Copy(fo, fi2)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		err = fo.Close()
		if err == nil {
			return fmt.Errorf("TestWfsDssGetContentWriterBase should fail with err args")
		}
		if isDup, err := dss.IsDuplicate("484f617a695613aac4b346237aa01548"); isDup || err != nil {
			return fmt.Errorf("TestWfsDssGetContentWriterBase IsDuplicate failed %v %v", isDup, err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
