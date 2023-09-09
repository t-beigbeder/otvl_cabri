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
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) GetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) GetContentReader(npath string) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) Remove(npath string) error {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) GetMeta(npath string, getCh bool) (IMeta, error) {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) SetCurrentTime(time int64) {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) SetMetaMockCbs(cbs *MetaMockCbs) {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) SetAfs(tfs afero.Fs) {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) GetAfs() afero.Fs {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) Close() error {
	if wdi.reducer != nil {
		return wdi.reducer.Close()
	}
	return nil
}

func (wdi *wfsDssImpl) SetSu() {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) SuEnableWrite(npath string) error {
	//TODO implement me
	panic("implement me")
}

// NewWfsDss opens a web client for a remote "wdi" DSS (data storage system)
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
