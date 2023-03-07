package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"testing"
)

func createWebDssServer(addr, root string, params CreateNewParams) (WebServer, error) {
	dss, err := CreateOrNewDss(params)
	_ = dss
	if err != nil {
		return nil, fmt.Errorf("createWebDssServer failed with error %v", err)
	}
	return NewWebDssServer(addr, root, WebDssServerConfig{Dss: dss.(HDss), HasLog: true})
}

func TestNewWebDssServer(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestNewWebDssServer", tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()

	getPIndex := func(config DssBaseConfig, _ string) (Index, error) {
		return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
	}

	sv, err := createWebDssServer(":3000", "",
		CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex},
	)
	if err != nil {
		t.Fatal(err)
	}
	sv.Shutdown()
	// check unlocked
	sv, err = createWebDssServer(":3000", "",
		CreateNewParams{Create: false, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sv.Shutdown()
	sv, err = createWebDssServer(":3000", "",
		CreateNewParams{Create: false, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex},
	)
	if err == nil {
		t.Fatal("should fail with error index.bdb locked")
	}
}

func TestWebDssStoreMeta(t *testing.T) {
	optionalSkip(t)
	sm := mStoreMeta{Npath: "é", Time: 255, Bs: []byte("ç")}
	bs, err := json.Marshal(sm)
	if err != nil {
		t.Fatal(err)
	}
	sm = mStoreMeta{Npath: "é", Time: -1, Bs: []byte("ç")}
	bs, err = json.Marshal(sm)
	if err != nil {
		t.Fatal(err)
	}
	var sm2 mStoreMeta
	if err = json.Unmarshal(bs, &sm2); err != nil || sm2.Time != -1 || string(sm2.Bs) != "ç" {
		t.Fatal(err, sm2)
	}
}

func TestNewWebDssClientOlf(t *testing.T) {
	ucpCount := 0
	var sv WebServer
	var err error
	defer func() {
		if sv != nil {
			sv.Shutdown()
		}
	}()
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			getPIndex := func(config DssBaseConfig, _ string) (Index, error) {
				return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
			}
			sv, err = createWebDssServer(":3000", "",
				CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			ucpCount += 1
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						ConfigDir: ufpath.Join(tfs.Path(), fmt.Sprintf(".cabri-i%d", ucpCount)),
						WebPort:   "3000",
					}, NoClientLimit: true},
				0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func runTestNewWebDssClientObs(t *testing.T) error {
	var sv WebServer
	var err error
	defer func() {
		if sv != nil {
			sv.Shutdown()
		}
	}()
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			if err := CleanObsDss(getOC()); err != nil {
				t.Fatal(err)
			}
			getPIndex := func(config DssBaseConfig, _ string) (Index, error) {
				return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
			}
			sv, err = createWebDssServer(":3000", "",
				CreateNewParams{
					Create: true, DssType: "obs", LocalPath: tfs.Path(), GetIndex: getPIndex,
					Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK"),
				},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
						WebPort:   "3000",
					}, NoClientLimit: true},
				0, nil)
			return dss, err
		}); err != nil {
		return err
	}
	return nil
}

func TestNewWebDssClientObs(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestNewWebDssClientObs(t)
	})
}

func TestNewWebDssClientSmf(t *testing.T) {
	var sv WebServer
	var err error
	defer func() {
		if sv != nil {
			sv.Shutdown()
		}
	}()
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			getPIndex := func(config DssBaseConfig, _ string) (Index, error) {
				return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
			}
			sv, err = createWebDssServer(":3000", "",
				CreateNewParams{
					Create: true, DssType: "smf", LocalPath: tfs.Path(), GetIndex: getPIndex,
				},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
						WebPort:   "3000",
					}, NoClientLimit: true},
				0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func TestNewWebDssApiClientOlf(t *testing.T) {
	ucpCount := 0
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			ucpCount += 1
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						LibApi:    true,
						ConfigDir: ufpath.Join(tfs.Path(), fmt.Sprintf(".cabri-i%d", ucpCount)),
					},
					LibApiDssConfig: LibApiDssConfig{
						IsOlf: true,
						OlfCfg: OlfConfig{
							DssBaseConfig: DssBaseConfig{
								LocalPath: tfs.Path(),
								GetIndex: func(config DssBaseConfig, _ string) (Index, error) {
									return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
								},
							}, Root: tfs.Path(), Size: "s"},
					},
				},
				0, nil)
			return dss, err

		}); err != nil {
		t.Fatal(err)
	}
}

func runTestNewWebDssApiClientObs(t *testing.T) error {
	return runTestBasic(t,
		func(tfs *testfs.Fs) error {
			if err := CleanObsDss(getOC()); err != nil {
				return err
			}
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			dss, err := CreateObsDss(config)
			if err != nil {
				return err
			}
			dss.Close()
			return nil
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dbc := getOC()
			dbc.LocalPath = tfs.Path()
			dbc.DssBaseConfig.GetIndex = GetPIndex
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						LibApi:    true,
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
					},
					LibApiDssConfig: LibApiDssConfig{
						IsObs:  true,
						ObsCfg: dbc,
					},
				},
				0, nil)
			return dss, err
		})
}

func TestNewWebDssApiClientObs(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestNewWebDssApiClientObs(t)
	})
}

func TestNewWebDssApiClientSmf(t *testing.T) {
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			dss, err := CreateObsDss(config)
			if err != nil {
				return err
			}
			dss.Close()
			return nil
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dbc := getOC()
			dbc.LocalPath = tfs.Path()
			dbc.DssBaseConfig.GetIndex = GetPIndex
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						LibApi:    true,
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
					},
					LibApiDssConfig: LibApiDssConfig{
						IsSmf:  true,
						ObsCfg: dbc,
					},
				},
				0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func TestWebClientOlfHistory(t *testing.T) {
	var sv WebServer
	var err error
	defer func() {
		if sv != nil {
			sv.Shutdown()
		}
	}()
	if err := runTestHistory(t,
		func(tfs *testfs.Fs) error {
			getPIndex := func(config DssBaseConfig, _ string) (Index, error) {
				return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
			}
			sv, err = createWebDssServer(":3000", "",
				CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
						WebPort:   "3000",
					}, NoClientLimit: true},
				0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func runTestWebClientObsHistory(t *testing.T) error {
	var sv WebServer
	var err error
	defer func() {
		if sv != nil {
			sv.Shutdown()
		}
	}()
	if err := runTestHistory(t,
		func(tfs *testfs.Fs) error {
			if err := CleanObsDss(getOC()); err != nil {
				t.Fatal(err)
			}
			getPIndex := func(config DssBaseConfig, _ string) (Index, error) {
				return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
			}
			sv, err = createWebDssServer(":3000", "",
				CreateNewParams{
					Create: true, DssType: "obs", LocalPath: tfs.Path(), GetIndex: getPIndex,
					Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK"),
				},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
						WebPort:   "3000",
					}, NoClientLimit: true},
				0, nil)
			return dss, err
		}); err != nil {
		return err
	}
	return nil
}

func TestWebClientObsHistory(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestWebClientObsHistory(t)
	})
}

func TestWebDssApiClientOlfHistory(t *testing.T) {
	if err := runTestHistory(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						LibApi:    true,
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
					},
					LibApiDssConfig: LibApiDssConfig{
						IsOlf: true,
						OlfCfg: OlfConfig{
							DssBaseConfig: DssBaseConfig{
								LocalPath: tfs.Path(),
								GetIndex:  GetPIndex,
							}, Root: tfs.Path(), Size: "s"},
					},
				},
				0, nil)
			return dss, err

		}); err != nil {
		t.Fatal(err)
	}
}

func runTestWebDssApiClientObsHistory(t *testing.T) error {
	return runTestHistory(t,
		func(tfs *testfs.Fs) error {
			if err := CleanObsDss(getOC()); err != nil {
				return err
			}
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			dss, err := CreateObsDss(config)
			if err != nil {
				return err
			}
			dss.Close()
			return nil
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dbc := getOC()
			dbc.LocalPath = tfs.Path()
			dbc.DssBaseConfig.GetIndex = GetPIndex
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						LibApi:    true,
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
					},
					LibApiDssConfig: LibApiDssConfig{
						IsObs:  true,
						ObsCfg: dbc,
					},
				},
				0, nil)
			return dss, err
		})
}

func TestWebDssApiClientObsHistory(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestWebDssApiClientObsHistory(t)
	})
}

func TestWebClientOlfMultiHistory(t *testing.T) {
	var sv WebServer
	var err error
	defer func() {
		if sv != nil {
			sv.Shutdown()
		}
	}()
	if err := runTestMultiHistory(t,
		func(tfs *testfs.Fs) error {
			getPIndex := func(config DssBaseConfig, _ string) (Index, error) {
				return NewPIndex(ufpath.Join(tfs.Path(), "index.bdb"), false, false)
			}
			sv, err = createWebDssServer(":3000", "",
				CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewWebDss(
				WebDssConfig{
					DssBaseConfig: DssBaseConfig{
						ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
						WebPort:   "3000",
					}, NoClientLimit: true},
				0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}
