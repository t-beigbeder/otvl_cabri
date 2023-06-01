//go:build test_cabridss

package cabridss

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

func (fsy *FsyDss) SetAfs(afs afero.Fs) {
	fsy.afs = afs
}

func (fsy *FsyDss) SetMetaMockCbs(cbs *MetaMockCbs) {
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
