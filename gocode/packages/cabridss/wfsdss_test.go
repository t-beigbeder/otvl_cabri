package cabridss

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
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

func runWfsDssTest(t *testing.T, doIt func(Dss) error) error {
	optionalSkip(t)
	tfs, err := testfs.CreateFs(t.Name(), tfsStartup)
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
		DssBaseConfig: DssBaseConfig{},
		NoClientLimit: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer dss.Close()
	err = doIt(dss)
	return err
}

func TestNewWfsDssClient(t *testing.T) {
	err := runWfsDssTest(t, func(dss Dss) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = runWfsDssTest(t, func(dss Dss) error {
		return fmt.Errorf("bad")
	})
	if err == nil {
		t.Fatal("should fail with error")
	}
}

func TestNewWfsDssBase(t *testing.T) {
	err := runWfsDssTest(t, func(dss Dss) error {
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
