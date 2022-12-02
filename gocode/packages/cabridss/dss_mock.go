//go:build test_cabridss

package cabridss

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

func (dss *FsyDss) SetAfs(afs afero.Fs) {
	dss.afs = afs
}

func (dss *FsyDss) SetMetaMockCbs(cbs *MetaMockCbs) {
	panic("FsyDss.SetMetaMockCbs is not implemented")
}

func OsUserHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeDir = filepath.Join(homeDir, ".cabri", "tests")
		err = os.MkdirAll(homeDir, 0o777)
	}
	return homeDir, err
}
