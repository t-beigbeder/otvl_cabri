package cabriui

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"os"
	"testing"
)

func optionalSkip(t *testing.T) {
	if os.Getenv("CABRIUI_SKIP_DEV_TESTS") != "" {
		if t.Name() == "theBeginning" ||
			t.Name() == "TestDSSMkBase" ||
			t.Name() == "TestDSSMknsRun" ||
			t.Name() == "TestLsnsArboBase1" ||
			t.Name() == "TestLsnsArboNoFear" ||
			t.Name() == "TestSampleStartup" ||
			t.Name() == "TestSynchronizeArboTiny" ||
			t.Name() == "TestSynchronizeArboBase" ||
			t.Name() == "TestSynchronizeArboNoFear" ||
			t.Name() == "TestSynchronizeArboBiDir" ||
			t.Name() == "TestSynchronizeToObs" ||
			t.Name() == "TestCheckUiAclMap" ||
			t.Name() == "theEnd" {
			t.Skip(fmt.Sprintf("Skipping %s because you set CABRIUI_SKIP_DEV_TESTS", t.Name()))
		}
	}
}

func getObjOptions() BaseOptions {
	return BaseOptions{
		ObsRegions:    []string{os.Getenv("OVHRG"), os.Getenv("OVHRG")},
		ObsEndpoints:  []string{os.Getenv("OVHEP"), os.Getenv("OVHEP")},
		ObsContainers: []string{os.Getenv("OVHCT"), os.Getenv("OVHCT")},
		ObsAccessKeys: []string{os.Getenv("OVHAK"), os.Getenv("OVHAK")},
		ObsSecretKeys: []string{os.Getenv("OVHSK"), os.Getenv("OVHSK")},
	}
}

func getObsConfig() cabridss.ObsConfig {
	return cabridss.ObsConfig{Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK")}
}
