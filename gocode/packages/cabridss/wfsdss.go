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
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) Updatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
	//TODO implement me
	panic("implement me")
}

func (wdi *wfsDssImpl) Lsns(npath string) (children []string, err error) {
	//TODO implement me
	panic("implement me")
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
	if wdc.ReducerLimit != 0 {
		wdi.reducer = plumber.NewReducer(wdc.ReducerLimit, 0)
	}
	return wdi, nil
}
