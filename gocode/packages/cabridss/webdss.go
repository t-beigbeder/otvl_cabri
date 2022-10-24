package cabridss

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type LibApiDssConfig struct {
	IsOlf  bool
	OlfCfg OlfConfig
	IsObs  bool
	ObsCfg ObsConfig
}

type WebDssConfig struct {
	DssBaseConfig
	LibApiDssConfig
	ClId          string
	NoClientLimit bool
}

type webContentWriterHandler struct {
	header []byte
	offset int
	file   *os.File
}

func (hdler *webContentWriterHandler) Read(p []byte) (n int, err error) {
	if hdler.offset < len(hdler.header) {
		n = copy(p, hdler.header[hdler.offset:])
		hdler.offset += n
	}
	for {
		if n >= len(p) {
			break
		}
		nf, errf := hdler.file.Read(p[n:])
		if errf != nil {
			err = errf
			return
		}
		hdler.offset += nf
		n += nf
	}
	return
}

func (hdler *webContentWriterHandler) Close() error {
	return hdler.file.Close()
}

type webDssImpl struct {
	oDssBaseImpl
	clId   string
	apc    WebApiClient
	repoId string
	libApi bool
}

func (wdi *webDssImpl) initialize(config interface{}, lsttime int64, aclusers []string) error {
	wdi.me = wdi
	wdi.lsttime = lsttime
	wdi.aclusers = aclusers

	wdc := config.(webDssClientConfig)
	var (
		uc        UserConfig
		err, err2 error
		ucp       string
	)
	if wdc.UserConfigPath != "" {
		uc, err = GetUserConfig(wdc.DssBaseConfig, wdc.UserConfigPath)
		ucp = wdc.UserConfigPath
	} else {
		uc, err = GetHomeUserConfig(wdc.DssBaseConfig)
		ucp, err2 = GetHomeUserConfigPath(wdc.DssBaseConfig)
	}
	if err != nil {
		return fmt.Errorf("in initialize: %v", err)
	}
	if err2 != nil {
		return fmt.Errorf("in initialize: %v", err2)
	}
	wdc.identities = uc.Identities
	wdi.clId = uc.ClientId

	var mIed *mInitialized
	wdc.ClId = wdi.clId
	wdi.apc = NewWebApiClient(wdc.WebProtocol, wdc.WebHost, wdc.WebPort, wdc.WebRoot, wdc)
	if wdc.NoClientLimit {
		wdi.apc.SetNoLimit()
	}
	mIed, err = cInitialize(wdi.apc)
	if err != nil {
		return fmt.Errorf("in initialize: %v", err)
	}
	wdi.repoId = mIed.RepoId
	wdi.repoEncrypted = mIed.Encrypted
	if wdi.repoId == "" {
		return fmt.Errorf("in initialize: the repository has no id")
	}
	if !mIed.PersistentIndex {
		return fmt.Errorf("in initialize: the repository has no persistent index")
	}

	var udd *mUpdatedData
	if !mIed.ClientIsKnown {
		udd, err = cRecordClient(wdi.apc)
	} else {
		udd, err = cUpdateClient(wdi.apc, wdc.ReIndex)
	}
	if err != nil {
		return fmt.Errorf("in initialize: %v", err)
	}
	cixf := filepath.Join(ucp, fmt.Sprintf("%s-%s.bdb", wdc.ClId, mIed.RepoId))
	cix, err := NewPIndex(cixf, wdc.Unlock, wdc.AutoRepair)
	if err != nil {
		return fmt.Errorf("in initialize: %v", err)
	}
	if err = cix.updateData(udd.UpdatedData, !mIed.ClientIsKnown); err != nil {
		return fmt.Errorf("in initialize: %v", err)
	}
	wdi.index = cix
	return nil
}

func (wdi *webDssImpl) loadMeta(npath string, time int64) ([]byte, error) {
	mo, err := cLoadMeta(wdi.apc, npath, time)
	if err != nil {
		return nil, fmt.Errorf("in loadMeta: %v", err)
	}
	return mo.Bs, nil
}

func (wdi *webDssImpl) queryMetaTimes(npath string) (times []int64, err error) {
	mt, err := cQueryMetaTimes(wdi.apc, npath)
	if err != nil {
		return nil, fmt.Errorf("in queryMetaTimes: %v", err)
	}
	return mt.Times, nil
}

func (wdi *webDssImpl) storeMeta(npath string, time int64, bs []byte) error {
	if err := cStoreMeta(wdi.apc, npath, time, bs); err != nil {
		return fmt.Errorf("in storeMeta: %v", err)
	}
	return nil
}

func (wdi *webDssImpl) xStoreMeta(npath string, time int64, bs []byte, acl []ACLEntry) error {
	if err := cXStoreMeta(wdi.apc, npath, time, bs, acl); err != nil {
		return fmt.Errorf("in xStoreMeta: %v", err)
	}
	return wdi.index.storeMeta(npath, time, bs)
}

func (wdi *webDssImpl) removeMeta(npath string, time int64) error {
	if err := cRemoveMeta(wdi.apc, npath, time); err != nil {
		return fmt.Errorf("in removeMeta: %v", err)
	}
	return nil
}

func (wdi *webDssImpl) xRemoveMeta(npath string, time int64) error {
	if err := cXRemoveMeta(wdi.apc, npath, time); err != nil {
		return fmt.Errorf("in xRemoveMeta: %v", err)
	}
	return wdi.index.removeMeta(npath, time)
}

func (wdi *webDssImpl) onWebCloseContent(npath string, mtime int64, cf afero.File, size int64, sha256 []byte, acl []ACLEntry, smCb storeMetaCallback) error {
	jsonArgs, err := json.Marshal(mOnCloseContentIn{Npath: npath, Mtime: mtime, Size: size, Ch: internal.Sha256ToStr32(sha256), ACL: acl})
	if err != nil {
		return fmt.Errorf("in onWebCloseContent: %v", err)
	}
	lja := internal.Int64ToStr16(int64(len(jsonArgs)))
	file, err := os.Open(cf.Name())
	if err != nil {
		return fmt.Errorf("in onWebCloseContent: %v", err)
	}
	hdler := webContentWriterHandler{header: make([]byte, 16+len(jsonArgs)), file: file}
	copy(hdler.header, lja)
	copy(hdler.header[16:], jsonArgs)
	req, err := http.NewRequest(http.MethodPost, wdi.apc.Url()+"onCloseContent", nil)
	req.Body = &hdler
	req.Header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
	resp, err := wdi.apc.(*apiClient).client.Do(req)
	if err = NewClientErr("onWebCloseContent", resp, err, nil); err != nil {
		return err
	}
	bs, err := io.ReadAll(resp.Body)
	var occo mOnCloseContentOut
	if err = json.Unmarshal(bs, &occo); err != nil {
		return fmt.Errorf("in onWebCloseContent: %v", err)
	}
	if occo.Error != "" {
		return fmt.Errorf("in onWebCloseContent: %s", occo.Error)
	}
	if err = smCb(occo.Npath, occo.Time, occo.Bs); err != nil {
		return fmt.Errorf("in onWebCloseContent: %v", err)
	}
	return nil
}

func (wdi *webDssImpl) onLibCloseContent(npath string, mtime int64, cf afero.File, size int64, sha256trunc []byte, acl []ACLEntry, smCb storeMetaCallback) error {
	wdc := wdi.apc.GetConfig().(webDssClientConfig)
	scf, err := os.CreateTemp("", "scw")
	if err != nil {
		return fmt.Errorf("in onLibCloseContent: %v", err)
	}
	ccf, err := os.Open(cf.Name())
	if err != nil {
		return fmt.Errorf("in onLibCloseContent: %v", err)
	}
	defer ccf.Close()

	var cbErr error
	var cbOut mOnCloseContentOut
	lcb := func(err error, size int64, sha256trunc []byte) {
		if err != nil {
			return
		}
		proxy := wdc.libDss.(*ODss).proxy
		cbErr = proxy.onCloseContent(npath, mtime, scf, size, sha256trunc, acl, func(npath string, time int64, bs []byte) error {
			cbOut = mOnCloseContentOut{Npath: npath, Time: time, Bs: bs}
			if err = proxy.xStoreMeta(npath, time, bs, acl); err != nil {
				return fmt.Errorf("in onLibCloseContent: %w", err)
			}
			return proxy.storeMeta(npath, time, bs)
		})
	}
	wter := &ContentHandle{cb: lcb, cf: scf, h: sha256.New()}
	n, err := io.Copy(wter, ccf)
	if err != nil || n != size {
		return fmt.Errorf("in onLibCloseContent: %v %d %d", err, n, size)
	}
	if err = wter.Close(); err != nil {
		return fmt.Errorf("in onLibCloseContent: %w", err)
	}
	if cbErr != nil {
		return fmt.Errorf("in onLibCloseContent: %w", cbErr)
	}
	if err = smCb(npath, cbOut.Time, cbOut.Bs); err != nil {
		return fmt.Errorf("in onLibCloseContent: %v", err)
	}
	return nil
}

func (wdi *webDssImpl) onCloseContent(npath string, mtime int64, cf afero.File, size int64, sha256 []byte, acl []ACLEntry, smCb storeMetaCallback) error {
	defer os.Remove(cf.Name())
	if wdi.libApi {
		return wdi.onLibCloseContent(npath, mtime, cf, size, sha256, acl, smCb)
	}
	return wdi.onWebCloseContent(npath, mtime, cf, size, sha256, acl, smCb)
}

func (wdi *webDssImpl) doGetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	cf, err := os.CreateTemp("", "ccw")
	if err != nil {
		return nil, fmt.Errorf("in doGetContentWriter: %w", err)
	}
	lcb := func(err error, size int64, sha256trunc []byte) {
		if err == nil {
			err = wdi.onCloseContent(npath, mtime, cf, size, sha256trunc, acl, func(npath string, time int64, bs []byte) error {
				if err = wdi.xStoreMeta(npath, time, bs, acl); err != nil {
					return fmt.Errorf("in doGetContentWriter: %w", err)
				}
				return nil
			})
		}
		if cb != nil {
			cb(err, size, sha256trunc)
		}
	}
	return &ContentHandle{cb: lcb, cf: cf, h: sha256.New()}, nil
}

func (wdi *webDssImpl) doWebGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	reqBody, err := json.Marshal(mDoGetContentReader{Npath: npath, MData: meta})
	if err != nil {
		return nil, fmt.Errorf("in doGetContentReader: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, wdi.apc.Url()+"doGetContentReader", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("in doGetContentReader: %w", err)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp, err := wdi.apc.(*apiClient).client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("in doGetContentReader: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		bs, err := io.ReadAll(resp.Body)
		return nil, NewClientErr("doGetContentReader", resp, err, bs)
	}
	slj := make([]byte, 16)
	if n, err := resp.Body.Read(slj); n != 16 || err != nil {
		return nil, fmt.Errorf("in doGetContentReader: %w", err)
	}
	lj, err := internal.Str16ToInt64(string(slj))
	if err != nil {
		return nil, fmt.Errorf("in doGetContentReader: %w", err)
	}
	if lj != 0 {
		sErr := make([]byte, lj)
		if n, err := resp.Body.Read(sErr); n != int(lj) || (err != nil && err != io.EOF) {
			return nil, fmt.Errorf("in doGetContentReader: %w", err)
		}
		return nil, fmt.Errorf("in doGetContentReader: %s", sErr)
	}
	return resp.Body, nil
}

func (wdi *webDssImpl) doLibGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	wdc := wdi.apc.GetConfig().(webDssClientConfig)
	proxy := wdc.libDss.(*ODss).proxy
	return proxy.doGetContentReader(npath, meta)
}

func (wdi *webDssImpl) doGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	if wdi.libApi {
		return wdi.doLibGetContentReader(npath, meta)
	}
	return wdi.doWebGetContentReader(npath, meta)
}

func (wdi *webDssImpl) queryContent(ch string) (exist bool, err error) {
	ex, err := cQueryContent(wdi.apc, ch)
	if err != nil {
		return false, fmt.Errorf("in queryContent: %v", err)
	}
	return ex.Exist, nil
}

func (wdi *webDssImpl) dumpIndex() string {
	rdi, err := cDumpIndex(wdi.apc)
	if err != nil {
		return fmt.Errorf("in queryContent: %v", err).Error()
	}
	return strings.Join([]string{"Remote", rdi.Dump, "Local", wdi.getIndex().Dump()}, "\n")
}

func (wdi *webDssImpl) setAfs(tfs afero.Fs) { panic("inconsistent") }

func (wdi *webDssImpl) getAfs() afero.Fs { panic("inconsistent") }

func copyMap[T any](dst map[string]T, src map[string]T) {
	for k, v := range src {
		dst[k] = v
	}
}

func (wdi *webDssImpl) scanPhysicalStorage(sti StorageInfo, errs *ErrorCollector) {
	sts, err := cScanPhysicalStorage(wdi.apc)
	if err != nil {
		errs.Collect(fmt.Errorf("in scanPhysicalStorage: %v", err))
		return
	}

	copyMap(sti.Path2Meta, sts.Sti.Path2Meta)
	copyMap(sti.Path2Content, sts.Sti.Path2Content)
	copyMap(sti.ExistingCs, sts.Sti.ExistingCs)
	copyMap(sti.Path2Error, sts.Sti.Path2Error)
	errs = &sts.Errs
}

func newWebDssProxy(config WebDssConfig, lsttime int64, aclusers []string) (oDssProxy, HDss, error) {
	impl := webDssImpl{}
	var (
		dss HDss
		err error
	)
	if config.LibApi {
		impl.libApi = true
		if config.IsOlf {
			if dss, err = NewOlfDss(config.OlfCfg, lsttime, aclusers); err != nil {
				return nil, nil, err
			}
		} else if config.IsObs {
			if dss, err = NewObsDss(config.ObsCfg, lsttime, aclusers); err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, fmt.Errorf("LibApi configuration is neither olf or obs")
		}
	}
	return &impl, dss, nil
}

// NewWebDss opens a web client for an "object-storage" DSS (data storage system)
// config provides the object store specification
// lsttime if not zero is the upper time of entries retrieved in it
// aclusers if not nil is a List of ACL users for access check
// returns a pointer to the ready to use DSS or an error if any occur
// If lsttime is not zero, access will be read-only
func NewWebDss(config WebDssConfig, lsttime int64, aclusers []string) (HDss, error) {
	proxy, libDss, err := newWebDssProxy(config, lsttime, aclusers)
	if err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	wdcc := webDssClientConfig{WebDssConfig: config, libDss: libDss}
	if err := proxy.initialize(wdcc, lsttime, aclusers); err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	return &ODss{proxy: proxy}, nil
}
