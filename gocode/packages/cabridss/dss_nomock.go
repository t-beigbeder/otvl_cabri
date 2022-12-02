//go:build !test_cabridss

package cabridss

import (
	"github.com/spf13/afero"
	"os"
)

func (dss *FsyDss) SetAfs(afs afero.Fs) {
	panic("FsyDss.SetAfs is forbidden")
}

func (dss *FsyDss) SetMetaMockCbs(cbs *MetaMockCbs) {
	panic("FsyDss.SetMetaMockCbs is forbidden")
}

func OsUserHomeDir() (string, error) {
	return os.UserHomeDir()
}
