package cabridss

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"io"
)

type WfsDssConfig struct {
	DssBaseConfig
	NoClientLimit bool
}

type wfsDssImpl struct {
	// like webDssImpl
	apc     WebApiClient
	reducer plumber.Reducer
	su      bool
}

func (wdi *wfsDssImpl) Mkns(npath string, mtime int64, children []string, acl []ACLEntry) error {
	if wdi.reducer == nil {
		return cfsMkns(wdi.apc, npath, mtime, children, acl)
	}
	return wdi.reducer.Launch(
		fmt.Sprintf("Mkns %s", npath),
		func() error {
			return cfsMkns(wdi.apc, npath, mtime, children, acl)
		})
}

func (wdi *wfsDssImpl) Updatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
	if wdi.reducer == nil {
		return cfsUpdatens(wdi.apc, npath, mtime, children, acl)
	}
	return wdi.reducer.Launch(
		fmt.Sprintf("Updatens %s", npath),
		func() error {
			return cfsUpdatens(wdi.apc, npath, mtime, children, acl)
		})
}

func (wdi *wfsDssImpl) Lsns(npath string) (children []string, err error) {
	if wdi.reducer == nil {
		return cfsLsns(wdi.apc, npath)
	}
	if err = wdi.reducer.Launch(
		fmt.Sprintf("Lsns %s", npath),
		func() error {
			var iErr error
			if children, iErr = cfsLsns(wdi.apc, npath); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (wdi *wfsDssImpl) IsDuplicate(ch string) (bool, error) {
	return false, nil
}

func (wdi *wfsDssImpl) GetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (wc io.WriteCloser, err error) {
	if wdi.reducer == nil {
		return cfsGetContentWriter(wdi.apc, npath, mtime, acl, cb)
	}
	if err = wdi.reducer.Launch(
		fmt.Sprintf("GetContentWriter %s", npath),
		func() error {
			var iErr error
			if wc, iErr = cfsGetContentWriter(wdi.apc, npath, mtime, acl, cb); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (wdi *wfsDssImpl) GetContentReader(npath string) (rc io.ReadCloser, err error) {
	if wdi.reducer == nil {
		return cfsGetContentReader(wdi.apc, npath)
	}
	if err = wdi.reducer.Launch(
		fmt.Sprintf("GetContentReader %s", npath),
		func() error {
			var iErr error
			if rc, iErr = cfsGetContentReader(wdi.apc, npath); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return

}

func (wdi *wfsDssImpl) Remove(npath string) (err error) {
	if wdi.reducer == nil {
		return cfsRemove(wdi.apc, npath)
	}
	if err = wdi.reducer.Launch(
		fmt.Sprintf("Lsns %s", npath),
		func() error {
			var iErr error
			if iErr = cfsRemove(wdi.apc, npath); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (wdi *wfsDssImpl) GetMeta(npath string, getCh bool) (meta IMeta, err error) {
	if wdi.reducer == nil {
		return cfsGetMeta(wdi.apc, npath, getCh)
	}
	if err = wdi.reducer.Launch(
		fmt.Sprintf("GetMeta %s", npath),
		func() error {
			var iErr error
			if meta, iErr = cfsGetMeta(wdi.apc, npath, getCh); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (wdi *wfsDssImpl) SetCurrentTime(time int64) {
	panic("not (yet) implemented")
}

func (wdi *wfsDssImpl) SetMetaMockCbs(cbs *MetaMockCbs) {
	panic("not (yet) implemented")
}

func (wdi *wfsDssImpl) SetAfs(tfs afero.Fs) {
	panic("not (yet) implemented")
}

func (wdi *wfsDssImpl) GetAfs() afero.Fs {
	return appFs
}

func (wdi *wfsDssImpl) Close() error {
	if wdi.reducer != nil {
		return wdi.reducer.Close()
	}
	return nil
}

func (wdi *wfsDssImpl) SetSu() { wdi.su = true }

func (wdi *wfsDssImpl) SuEnableWrite(npath string) (err error) {
	if !wdi.su {
		return fmt.Errorf("in SuEnableWrite: not in su mode")
	}
	if wdi.reducer == nil {
		return cfsSuEnableWrite(wdi.apc, npath)
	}
	if err = wdi.reducer.Launch(
		fmt.Sprintf("SuEnableWrite %s", npath),
		func() error {
			var iErr error
			if iErr = cfsSuEnableWrite(wdi.apc, npath); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

// NewWfsDss opens a web client for a remote "fsy" DSS (data storage system)
// wdc provides the web client configuration
// returns a pointer to the ready to use DSS or an error if any occur
func NewWfsDss(wdc WfsDssConfig) (Dss, error) {
	wdi := &wfsDssImpl{}
	var tlsConfig *TlsConfig
	var err error
	if wdc.WebProtocol == "https" {
		tlsConfig = &TlsConfig{
			cert:              wdc.TlsCert,
			key:               wdc.TlsKey,
			noClientCheck:     wdc.TlsNoCheck,
			basicAuthUser:     wdc.BasicAuthUser,
			basicAuthPassword: wdc.BasicAuthPassword,
		}
	}
	remoteWdc := wdc
	wdi.apc, err = NewWebApiClient(wdc.WebProtocol, wdc.WebHost, wdc.WebPort, tlsConfig, wdc.WebRoot, remoteWdc, wdc.WebClientTimeout)
	if err != nil {
		return nil, fmt.Errorf("in NewWfsDss: %w", err)
	}
	err = cfsInitialize(wdi.apc)
	if err != nil {
		return nil, fmt.Errorf("in NewWfsDss: %w", err)
	}
	if wdc.ReducerLimit != 0 {
		wdi.reducer = plumber.NewReducer(wdc.ReducerLimit, 0)
	}
	return wdi, nil
}
