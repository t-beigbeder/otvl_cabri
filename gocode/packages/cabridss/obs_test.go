package cabridss

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"os"
	"testing"
	"time"
)

func getOC() ObsConfig {
	return ObsConfig{Container: os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK")}
}

func getOCWithBase(baseConfig DssBaseConfig) ObsConfig {
	return ObsConfig{
		DssBaseConfig: baseConfig,
		Container:     os.Getenv("OVHCT"), Endpoint: os.Getenv("OVHEP"), Region: os.Getenv("OVHRG"), AccessKey: os.Getenv("OVHAK"), SecretKey: os.Getenv("OVHSK"),
	}
}

func TestNewObsDssBase(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestBasic(t,
			func(tfs *testfs.Fs) error {
				return CleanObsDss(getOC())
			},
			func(tfs *testfs.Fs) (HDss, error) {
				dss, err := NewObsDss(getOC(), 0, nil)
				return dss, err
			})
	})
}

func TestNewObsDssMockFsBase(t *testing.T) {
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			return nil
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewObsDss(ObsConfig{
				GetS3Session: func() IS3Session {
					return NewS3sMockFs(tfs.Path(), nil)
				},
			}, 0, nil)
			return dss, err
		}); err != nil {
		t.Fatal(err)
	}
}

func TestNewObsDssMockFsUnlock(t *testing.T) {
	optionalSkip(t)
	tfs, err := testfs.CreateFs("TestNewObsDssMockFsUnlock", tfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()

	config := ObsConfig{
		GetS3Session: func() IS3Session {
			return NewS3sMockFs(tfs.Path(), nil)
		},
	}
	config.LocalPath = tfs.Path()
	config.DssBaseConfig.GetIndex = GetPIndex
	dss, err := CreateObsDss(config)
	if err != nil {
		t.Fatal(err)
	}

	if err := dss.Mkns("", 0, []string{"d1é/"}, nil); err != nil {
		t.Fatal(err)
	}

	dss, err = NewObsDss(config, 0, nil)
	if err == nil {
		t.Fatal("NewObsDss should fail with NewPIndex lock error")
	}
	config.DssBaseConfig.Unlock = true
	dss, err = NewObsDss(config, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewObsDssS3Mock(t *testing.T) {
	locGetConfig := func(tfs *testfs.Fs) ObsConfig {
		config := getOC()
		config.GetS3Session = func() IS3Session {
			return NewS3Session(config, func(parent IS3Session) IS3Session {
				return NewS3sMockFs(tfs.Path(), func(parent IS3Session) IS3Session {
					return NewS3sMockTests(parent, func(args ...any) interface{} {
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
		return config
	}

	internal.Retry(t, func(t *testing.T) error {
		return runTestBasic(t,
			func(tfs *testfs.Fs) error {
				return CleanObsDss(locGetConfig(tfs))
			},
			func(tfs *testfs.Fs) (HDss, error) {
				dss, err := NewObsDss(locGetConfig(tfs), 0, nil)
				return dss, err
			})
	})
}

func TestCleanObsDss(t *testing.T) {
	optionalSkip(t)
	if err := CleanObsDss(getOC()); err != nil {
		t.Fatal(err)
	}
}

func TestNewObsDssMindex(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestBasic(t,
			func(tfs *testfs.Fs) error {
				return CleanObsDss(getOC())
			},
			func(tfs *testfs.Fs) (HDss, error) {
				config := getOC()
				config.DssBaseConfig.GetIndex = func(_ DssBaseConfig, _ string) (Index, error) {
					return NewMIndex(), nil
				}
				dss, err := NewObsDss(config, 0, nil)
				return dss, err
			})
	})
	config := getOC()
	config.DssBaseConfig.GetIndex = func(_ DssBaseConfig, _ string) (Index, error) {
		return NewMIndex(), nil
	}
	dss, err := NewObsDss(config, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dss.GetMeta("", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewObsDssPindex(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestBasic(t,
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
				config := getOC()
				config.LocalPath = tfs.Path()
				config.DssBaseConfig.GetIndex = GetPIndex
				dss, err := NewObsDss(config, 0, nil)
				return dss, err
			})
	})
	config := getOC()
	config.DssBaseConfig.GetIndex = func(_ DssBaseConfig, _ string) (Index, error) {
		return NewMIndex(), nil
	}
	dss, err := NewObsDss(config, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dss.GetMeta("", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMockFsHistory(t *testing.T) {
	if err := runTestHistory(t,
		func(tfs *testfs.Fs) error {
			config := getOC()
			config.GetS3Session = func() IS3Session {
				return NewS3sMockFs(tfs.Path(), nil)
			}
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
			config := getOC()
			config.GetS3Session = func() IS3Session {
				return NewS3sMockFs(tfs.Path(), nil)
			}
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			return NewObsDss(config, 0, nil)
		}); err != nil {
		t.Fatal(err)
	}
}

func runTestObsHistory(t *testing.T, redLimit int) error {
	optionalSkip(t)
	if err := CleanObsDss(getOC()); err != nil {
		return err
	}
	if err := runTestHistory(t,
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
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			config.ReducerLimit = redLimit
			dss, err := NewObsDss(config, 0, nil)
			if err != nil {
				return nil, err
			}
			return dss, err
		}); err != nil {
		return err
	}
	return nil
}

func TestObsHistory(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestObsHistory(t, 0)
	})
}

func TestObsRedHistory(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestObsHistory(t, 2)
	})
}

func runTestObsMultiHistory(t *testing.T) error {
	optionalSkip(t)

	if err := runTestMultiHistory(t,
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
			config := getOC()
			config.LocalPath = tfs.Path()
			config.DssBaseConfig.GetIndex = GetPIndex
			dss, err := NewObsDss(config, 0, nil)
			if err != nil {
				return nil, err
			}
			return dss, err
		}); err != nil {
		return err
	}
	return nil
}

func TestObsMultiHistory(t *testing.T) {
	internal.Retry(t, func(t *testing.T) error {
		return runTestObsMultiHistory(t)
	})
}
