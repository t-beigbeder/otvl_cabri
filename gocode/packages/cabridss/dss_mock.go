//go:build test_cabridss

package cabridss

import "github.com/spf13/afero"

func (dss *FsyDss) SetAfs(afs afero.Fs) {
	dss.afs = afs
}

func (dss *FsyDss) SetMetaMockCbs(cbs *MetaMockCbs) {
	panic("FsyDss.SetMetaMockCbs is not implemented")
}
