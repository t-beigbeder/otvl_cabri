package cabrisync

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"os"
	"testing"
	"time"
)

func getOC() cabridss.ObsConfig {
	return cabridss.ObsConfig{Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK")}
}

func createWebDssServer(addr, root string, params cabridss.CreateNewParams) (cabridss.WebServer, error) {
	dss, err := cabridss.CreateOrNewDss(params)
	_ = dss
	if err != nil {
		return nil, fmt.Errorf("createWebDssServer failed with error %v", err)
	}
	httpConfig := cabridss.WebServerConfig{Addr: addr, HasLog: false}
	s, err := cabridss.NewWebDssServer(root, cabridss.WebDssServerConfig{WebServerConfig: httpConfig, Dss: dss.(cabridss.HDss)})
	return s, err
}

func createWfsDssServer(addr, root string, dss cabridss.Dss) (cabridss.WebServer, error) {
	httpConfig := cabridss.WebServerConfig{Addr: addr, HasLog: false}
	return cabridss.NewWfsDssServer(root, cabridss.WfsDssServerConfig{
		WebServerConfig: httpConfig,
		Dss:             dss,
	})
}

func optionalSleep(t *testing.T) {
	if os.Getenv("CABRISYNC_FAST_TESTS") == "" {
		//fmt.Println("Sleeping 1.1s to check mtimes correctness")
		time.Sleep(1100 * time.Millisecond)
	}
}

func optionalSkip(t *testing.T) {
	if os.Getenv("CABRISYNC_SKIP_DEV_TESTS") != "" {
		if t.Name() == "TestSynchronizeBasic" ||
			t.Name() == "TestSynchronizeBasicACL" ||
			t.Name() == "TestSynchronizeBasicRed" ||
			t.Name() == "TestSynchronizeBasicFsyOlf" ||
			t.Name() == "TestSynchronizeBasicFsyOlfACL" ||
			t.Name() == "TestSynchronizeBasicFsyOlfRed" ||
			t.Name() == "TestSynchronizeBasicFsyObs" ||
			t.Name() == "TestSynchronizeBasicFsyObsACL" ||
			t.Name() == "TestSynchronizeBasicFsyWebOlf" ||
			t.Name() == "TestSynchronizeBasicFsyWebFsy" ||
			t.Name() == "TestSynchronizeBasicFsyEDssApiOlf" ||
			t.Name() == "TestSynchronizeBasicFsyEDssWebOlf" ||
			t.Name() == "TestSynchronizeBasicFsyEDssApiObs" ||
			t.Name() == "TestSynchroInconsistentChildren" ||
			t.Name() == "TestLoopSynchroInconsistentChildren" ||
			t.Name() == "TestMappedAcl" ||
			t.Name() == "TestMappedEncryptedAcl" ||
			t.Name() == "TestFsy2Fsy1" ||
			t.Name() == "TestSynchronizeArboTiny" ||
			t.Name() == "TestSynchronizeArboSmfPix" ||
			t.Name() == "TestSynchronizeArboObsPix" ||
			t.Name() == "TestSynchronizeArboWebDssClientOlf" ||
			t.Name() == "TestSynchronizeArboWebDssClientObs" ||
			t.Name() == "TestSynchronizeArboWebDssClientSmf" ||
			t.Name() == "TestSynchronizeArboBase" ||
			t.Name() == "TestSynchronizeArboNoFear" ||
			t.Name() == "TestSynchronizeArboBiDirOlf" ||
			t.Name() == "TestSynchronizeArboBiDirObs" ||
			t.Name() == "TestLoopSynchronizeArboBiDirObs" ||
			t.Name() == "theEnd" {
			t.Skip(fmt.Sprintf("Skipping %s because you set CABRISYNC_SKIP_DEV_TESTS", t.Name()))
		}
	}
}
