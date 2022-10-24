package cabridss

import (
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"testing"
)

func TestNewEDssApiClientOlf(t *testing.T) {
	t.Skip("TestNewEDssApiClientOlf WIP") // FIXME
	if err := runTestBasic(t,
		func(tfs *testfs.Fs) error {
			_, err := CreateOlfDss(OlfConfig{
				DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path(), Encrypted: true},
				Root:          tfs.Path(), Size: "s"})
			return err
		},
		func(tfs *testfs.Fs) (HDss, error) {
			dss, err := NewEDss(
				EDssConfig{
					WebDssConfig: WebDssConfig{
						DssBaseConfig: DssBaseConfig{
							LibApi:         true,
							UserConfigPath: ufpath.Join(tfs.Path(), ".cabri"),
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
				0, nil)
			return dss, err

		}); err != nil {
		t.Fatal(err)
	}
}
