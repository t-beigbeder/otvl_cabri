package cabridss

import (
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"testing"
)

func TestEDssClientOlfBase(t *testing.T) {
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
			sv, err = createWebDssServer(tfs, ":3000", "",
				CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex, Encrypted: true},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
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

func runTestEDssClientObsBase(t *testing.T) error {
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
			sv, err = createWebDssServer(tfs, ":3000", "",
				CreateNewParams{
					Create: true, DssType: "obs", LocalPath: tfs.Path(), GetIndex: getPIndex, Encrypted: true,
					Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK"),
				},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
							ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
							WebPort:   "3000",
						}, NoClientLimit: true},
				},
				0, nil)
			return dss, err
		}); err != nil {
		return err
	}
	return nil
}

func TestEDssClientObsBase(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestEDssClientObsBase(t)
	})
}

func TestEDssClientSmfBase(t *testing.T) {
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
			sv, err = createWebDssServer(tfs, ":3000", "",
				CreateNewParams{
					Create: true, DssType: "smf", LocalPath: tfs.Path(), GetIndex: getPIndex, Encrypted: true,
				},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
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

func TestEDssApiClientOlfBase(t *testing.T) {
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{
				DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path(), Encrypted: true},
				Root:          tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			ucp, uc, _ := newUcp(tfs)
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
							LibApi:    true,
							ConfigDir: ucp,
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
				},
				0, IdPkeys(uc))
			return dss, err

		}); err != nil {
		t.Fatal(err)
	}
}

func runTestEDssApiClientObsBase(t *testing.T) error {
	return runTestBasic(t,
		func(tfs *testfs.Fs) error {
			if err := CleanObsDss(getOC()); err != nil {
				return err
			}
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			config.Encrypted = true
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
			dbc.DssBaseConfig.Encrypted = true
			dss, err := NewEDss(
				EDssConfig{
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
				},
				0, nil)
			return dss, err
		})
}

func TestEDssApiClientObsBase(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestEDssApiClientObsBase(t)
	})
}

func TestEDssApiClientSmfBase(t *testing.T) {
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			config.Encrypted = true
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
			dbc.DssBaseConfig.Encrypted = true
			dss, err := NewEDss(
				EDssConfig{
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
				},
				0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func TestEDssClientOlfHistory(t *testing.T) {
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
			sv, err = createWebDssServer(tfs, ":3000", "",
				CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex, Encrypted: true},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
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

func TestEDssApiClientOlfHistory(t *testing.T) {
	if err := runTestHistory(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{
				DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path(), Encrypted: true},
				Root:          tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			ucp, uc, _ := newUcp(tfs)
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
							LibApi:    true,
							ConfigDir: ucp,
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
				},
				0, IdPkeys(uc))
			return dss, err

		}); err != nil {
		t.Fatal(err)
	}
}

func TestEDssClientOlfMultiHistory(t *testing.T) {
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
			sv, err = createWebDssServer(tfs, ":3000", "",
				CreateNewParams{Create: true, DssType: "olf", Root: tfs.Path(), Size: "s", GetIndex: getPIndex, Encrypted: true},
			)
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
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

func TestEDssApiClientOlfMultiHistory(t *testing.T) {
	if err := runTestMultiHistory(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{
				DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path(), Encrypted: true},
				Root:          tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			ucp, uc, _ := newUcp(tfs)
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
							LibApi:    true,
							ConfigDir: ucp,
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
				},
				0, IdPkeys(uc))
			return dss, err

		}); err != nil {
		t.Fatal(err)
	}
}
