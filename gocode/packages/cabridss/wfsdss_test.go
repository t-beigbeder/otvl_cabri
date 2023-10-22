package cabridss

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrifsu"
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
		_, err = dss.Lsns("d2/d3/nok")
		if err == nil {
			return fmt.Errorf("TestWfsDssLsnsBase should fail with error no such ns")
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

func TestWfsDssMtime(t *testing.T) {
	err := runWfsDssTest(t, func(tfs *testfs.Fs, dss Dss) error {
		fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
		if err != nil {
			t.Fatal(err.Error())
		}
		defer fi.Close()
		tt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
		if err = dss.Updatens("d", tt, []string{"a-copy.txt"}, nil); err != nil {
			t.Fatalf(err.Error())
		}
		dfi, err := os.Stat(ufpath.Join(tfs.Path(), "d"))
		if dfi.ModTime().Unix() != tt {
			t.Fatalf("TestFsyDssMtime mtime 'd' %d != %d", dfi.ModTime().Unix(), tt)
		}
		fo, err := dss.GetContentWriter("d/a-copy.txt", tt, nil, func(err error, size int64, ch string) {
			if err != nil {
				t.Fatal(err.Error())
			}
		})
		if err != nil {
			t.Fatal(err.Error())
		}
		io.Copy(fo, fi)
		fo.Close()
		ffi, err := os.Stat(ufpath.Join(tfs.Path(), "d", "a-copy.txt"))
		if ffi.ModTime().Unix() != tt {
			t.Fatalf("TestFsyDssMtime mtime 'd/a-copy.txt' %d != %d", ffi.ModTime().Unix(), tt)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWfsDssGetContentReaderBase(t *testing.T) {
	err := runWfsDssTest(t, func(tfs *testfs.Fs, dss Dss) error {
		fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
		if err != nil {
			t.Fatal(err.Error())
		}
		defer fi.Close()
		fo, err := dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		io.Copy(fo, fi)
		fo.Close()
		fi2, err := dss.GetContentReader("a-copy.txt")
		if err != nil {
			t.Fatal(err.Error())
		}
		defer fi2.Close()
		fo2, err := dss.GetContentWriter("a-copy-copy.txt", time.Now().Unix(), nil, func(err error, size int64, ch string) {
			if err != nil {
				t.Fatal(err.Error())
			}
			if size != 241 {
				t.Fatalf("TestFsyDssGetContentReaderBase size %d != 241", size)
			}
			if ch != "484f617a695613aac4b346237aa01548" {
				t.Fatalf("TestFsyDssGetContentReaderBase hash %s != %s", ch, "484f617a695613aac4b346237aa01548")
			}
		})
		io.Copy(fo2, fi2)
		fo2.Close()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWfsDssRemoveBase(t *testing.T) {
	err := runWfsDssTest(t, func(tfs *testfs.Fs, dss Dss) (err error) {
		if err = dss.Remove("/z"); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with params error")
		}
		if err = dss.Remove("//"); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with params error")
		}
		if err = dss.Remove("/"); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with params error")
		}
		if err = dss.Remove("nosuchdir/"); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with params error")
		}
		if err = dss.Remove("nosuchfile"); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with params error")
		}
		if err = dss.Remove("e/se"); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with is a dir error")
		}
		if err = dss.Remove("e/se/c1.txt/"); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with is a file error")
		}
		if err = os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
			return err
		}
		if err = tfs.RandTextFile("e/se/c2.txt", 1); err != nil {
			return err
		}
		if err = dss.Remove("e/se/c2.txt"); err != nil {
			t.Fatalf("TestWfsDssRemoveBase %v", err)
		}
		if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/se/c2.txt")); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with no such file e/se/c2.txt")
		}
		if err = tfs.RandTextFile("e/se/c2éà.txt", 1); err != nil {
			return err
		}
		if err = dss.Remove("e/se/c2éà.txt"); err != nil {
			t.Fatalf("TestWfsDssRemoveBase %v", err)
		}
		if err = dss.Remove("e/"); err != nil {
			t.Fatalf("TestWfsDssRemoveBase %v", err)
		}
		if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/se/c1.txt")); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with no such file e/se/c2.txt")
		}
		if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/se/")); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with no such dir e/se/")
		}
		if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/")); err == nil {
			t.Fatalf("TestWfsDssRemoveBase should fail with no such dir e/")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWfsGetMetaBase(t *testing.T) {
	err := runWfsDssTest(t, func(tfs *testfs.Fs, dss Dss) (err error) {
		if _, err = dss.GetMeta("/z", true); err == nil {
			t.Fatalf("TestWfsGetMetaBase should fail with params error")
		}
		if _, err = dss.GetMeta("//", true); err == nil {
			t.Fatalf("TestWfsGetMetaBase should fail with params error")
		}
		if _, err = dss.GetMeta("nosuchdir/", true); err == nil {
			t.Fatalf("TestWfsGetMetaBase should fail with params error")
		}
		if _, err = dss.GetMeta("nosuchfile", true); err == nil {
			t.Fatalf("TestWfsGetMetaBase should fail with params error")
		}
		if _, err = dss.GetMeta("e/se", true); err == nil {
			t.Fatalf("TestWfsGetMetaBase should fail with is a dir error")
		}
		if _, err = dss.GetMeta("e/se/c1.txt/", true); err == nil {
			t.Fatalf("TestWfsGetMetaBase should fail with is a file error")
		}
		var m IMeta
		if m, err = dss.GetMeta("d/", true); err != nil {
			return err
		}
		if m, err = dss.GetMeta("d/b.txt", true); err != nil {
			return err
		}
		if m, err = dss.GetMeta("d/b.txt", false); err != nil {
			return err
		}
		if m, err = dss.GetMeta("", false); err != nil {
			return err
		}
		if err = os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
			return err
		}
		if err = tfs.RandTextFile("e/se/c2éà.txt", 1); err != nil {
			return err
		}
		if m, err = dss.GetMeta("e/se/c2éà.txt", false); err != nil {
			return err
		}
		_ = m
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWfsSetSuBase(t *testing.T) {
	err := runWfsDssTest(t, func(tfs *testfs.Fs, dss Dss) (err error) {
		if err = cabrifsu.DisableWrite(dss.GetAfs(), ufpath.Join(tfs.Path(), "a.txt"), false); err != nil {
			return err
		}
		var m IMeta
		if m, err = dss.GetMeta("a.txt", true); err != nil || m.GetAcl()[0].Rights.Write {
			return fmt.Errorf("GetMeta %v %v", m, err)
		}
		if err = dss.SuEnableWrite("a.txt"); err == nil {
			return fmt.Errorf("SuEnableWrite should fail with not in su mode")
		}
		dss.SetSu()
		if err = dss.SuEnableWrite("a.txt"); err != nil {
			return err
		}
		if m, err = dss.GetMeta("a.txt", true); err != nil || !m.GetAcl()[0].Rights.Write {
			return fmt.Errorf("GetMeta %v %v", m, err)
		}
		if err = os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
			return err
		}
		if err = dss.SuEnableWrite("e/se/"); err != nil {
			return err
		}
		if err = dss.SuEnableWrite(""); err != nil {
			return err
		}
		if err = tfs.RandTextFile("e/se/c2éà.txt", 1); err != nil {
			return err
		}
		if err = dss.SuEnableWrite("e/se/c2éà.txt"); err != nil {
			return err
		}
		if m, err = dss.GetMeta("e/se/c2éà.txt", true); err != nil || !m.GetAcl()[0].Rights.Write {
			return fmt.Errorf("GetMeta %v %v", m, err)
		}
		if err = dss.SuEnableWrite(""); err != nil {
			return err
		}
		_ = m
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
