package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
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
	IsSmf  bool
	ObsCfg ObsConfig
}

type WebDssConfig struct {
	DssBaseConfig
	LibApiDssConfig
	ClId          string
	NoClientLimit bool
}

type webContentWriterHandler struct {
	header  []byte
	offset  int
	rCloser io.ReadCloser
}

func (hdler *webContentWriterHandler) Read(p []byte) (n int, err error) {
	if hdler.offset < len(hdler.header) {
		n = copy(p, hdler.header[hdler.offset:])
		hdler.offset += n
		return
	}
	for {
		if n >= len(p) {
			break
		}
		nf, errf := hdler.rCloser.Read(p[n:])
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
	return hdler.rCloser.Close()
}

type webDssImpl struct {
	oDssBaseImpl
	clId         string
	apc          WebApiClient
	repoId       string
	libApi       bool
	isClientEdss bool
}

func (wdi *webDssImpl) initialize(me oDssProxy, config interface{}, lsttime int64, aclusers []string) error {
	wdi.me = me
	wdi.lsttime = lsttime
	wdi.aclusers = aclusers

	wdc := config.(webDssClientConfig)
	var (
		uc        UserConfig
		err, err2 error
		ucp       string
	)
	if wdc.ConfigDir != "" {
		uc, err = GetUserConfig(wdc.DssBaseConfig, wdc.ConfigDir)
		ucp = wdc.ConfigDir
	} else {
		uc, err = GetHomeUserConfig(wdc.DssBaseConfig)
		ucp, err2 = GetHomeConfigDir(wdc.DssBaseConfig)
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
	remoteWdc := wdc
	remoteWdc.Unlock = false
	var tlsConfig *TlsConfig
	if wdc.WebProtocol == "https" {
		tlsConfig = &TlsConfig{
			cert:              wdc.TlsCert,
			key:               wdc.TlsKey,
			noClientCheck:     wdc.TlsNoCheck,
			basicAuthUser:     wdc.BasicAuthUser,
			basicAuthPassword: wdc.BasicAuthPassword,
		}
	}
	wdi.apc, err = NewWebApiClient(wdc.WebProtocol, wdc.WebHost, wdc.WebPort, tlsConfig, wdc.WebRoot, remoteWdc, wdc.WebClientTimeout)
	if err != nil {
		return fmt.Errorf("in initialize: %v", err)
	}
	wdi.apc.SetCabriHeader("WebApi")
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
	if err = wdi.me.spUpdateClient(cix, udd.UpdatedData, !mIed.ClientIsKnown); err != nil {
		cix.Close()
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

func (wdi *webDssImpl) removeMeta(npath string, time int64) error {
	if err := cRemoveMeta(wdi.apc, npath, time); err != nil {
		return fmt.Errorf("in removeMeta: %v", err)
	}
	return nil
}

func (wdi *webDssImpl) xRemoveMeta(meta Meta) error {
	ipath := RemoveSlashIfNsIf(meta.Path, meta.IsNs)
	if err := cXRemoveMeta(wdi.apc, ipath, meta.Itime); err != nil {
		return fmt.Errorf("in xRemoveMeta: %v", err)
	}
	return wdi.index.removeMeta(ipath, meta.Itime)
}

func (wdi *webDssImpl) webPushContent(size int64, ch string, mbs []byte, emid string, cf afero.File) error {
	jsonArgs, err := json.Marshal(mPushContentIn{Size: size, Ch: ch, Mbs: mbs, Emid: emid})
	if err != nil {
		return fmt.Errorf("in webPushContent: %w", err)
	}
	lja := internal.Int64ToStr16(int64(len(jsonArgs)))
	file, err := os.Open(cf.Name())
	if err != nil {
		return fmt.Errorf("in webPushContent: %w", err)
	}
	hdler := webContentWriterHandler{header: make([]byte, 16+len(jsonArgs)), rCloser: file}
	copy(hdler.header, lja)
	copy(hdler.header[16:], jsonArgs)
	req, err := http.NewRequest(http.MethodPost, wdi.apc.Url()+"pushContent", nil)
	req.Body = &hdler
	req.Header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
	resp, err := wdi.apc.(*apiClient).client.Do(req)
	if err = NewClientErr("webPushContent", resp, err, nil); err != nil {
		return err
	}
	bs, err := io.ReadAll(resp.Body)
	var pco mError
	if err = json.Unmarshal(bs, &pco); err != nil {
		return fmt.Errorf("in webPushContent: %v", err)
	}
	if pco.Error != "" {
		return fmt.Errorf("in webPushContent: %s", pco.Error)
	}
	return nil
}

func (wdi *webDssImpl) libPushContent(size int64, ch string, mbs []byte, emid string, cf afero.File) error {
	wdc := wdi.apc.GetConfig().(webDssClientConfig)
	ccf, err := os.Open(cf.Name())
	if err != nil {
		return fmt.Errorf("in libPushContent: %v", err)
	}
	defer ccf.Close()
	proxy := wdc.libDss.(*ODss).proxy
	wter, err := proxy.spGetContentWriter(contentWriterCbs{
		getMetaBytes: func(iErr error, size int64, ch string) ([]byte, string, error) {
			return mbs, emid, nil
		},
	}, nil)
	n, err := io.Copy(wter, ccf)
	if err != nil || n != size {
		return fmt.Errorf("in libPushContent: %v %d %d", err, n, size)
	}
	if err = wter.Close(); err != nil {
		return fmt.Errorf("in libPushContent: %w", err)
	}
	return nil
}

func (wdi *webDssImpl) pushContent(size int64, ch string, mbs []byte, emid string, cf afero.File) error {
	var err error
	if wdi.libApi {
		err = wdi.libPushContent(size, ch, mbs, emid, cf)
	} else {
		err = wdi.webPushContent(size, ch, mbs, emid, cf)
	}
	if err != nil {
		return err
	}
	return nil
}

func (wdi *webDssImpl) spGetContentWriter(cwcbs contentWriterCbs, acl []ACLEntry) (io.WriteCloser, error) {
	return NewTempFileWriteCloserWithCb(wdi.getAfs(), "", "cw", func(err error, size int64, ch string, wcwc *WriteCloserWithCb) error {
		outError := err
		defer func() {
			if cwcbs.closeCb != nil {
				cwcbs.closeCb(outError, size, ch)
			}
		}()
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter: %w", err)
			return outError
		}
		mbs, emid, err := cwcbs.getMetaBytes(err, size, ch)
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter: %w", err)
			return outError
		}
		meta, err := wdi.decodeMeta(mbs)
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter: %w", err)
			return outError
		}
		cf := wcwc.Underlying.(afero.File)
		// FIXME: check if upload is required
		if err := wdi.pushContent(size, ch, mbs, emid, cf); err != nil {
			outError = fmt.Errorf("in spGetContentWriter: %w", err)
			return outError
		}
		if err := wdi.index.storeMeta(meta.Path, meta.Itime, mbs); err != nil {
			outError = fmt.Errorf("in spGetContentWriter: %w", err)
			return outError
		}
		return nil
	})
}

func (wdi *webDssImpl) spWebGetContentReader(ch string) (io.ReadCloser, error) {
	reqBody, err := json.Marshal(mSpGetContentReader{Ch: ch})
	if err != nil {
		return nil, fmt.Errorf("in spWebGetContentReader: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, wdi.apc.Url()+"spGetContentReader", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("in spWebGetContentReader: %w", err)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp, err := wdi.apc.(*apiClient).client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("in spWebGetContentReader: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		bs, err := io.ReadAll(resp.Body)
		return nil, NewClientErr("spWebGetContentReader", resp, err, bs)
	}
	slj := make([]byte, 16)
	if n, err := resp.Body.Read(slj); n != 16 || err != nil {
		return nil, fmt.Errorf("in spWebGetContentReader: %w", err)
	}
	lj, err := internal.Str16ToInt64(string(slj))
	if err != nil {
		return nil, fmt.Errorf("in spWebGetContentReader: %w", err)
	}
	if lj != 0 {
		sErr := make([]byte, lj)
		if n, err := resp.Body.Read(sErr); n != int(lj) || (err != nil && err != io.EOF) {
			return nil, fmt.Errorf("in spWebGetContentReader: %w", err)
		}
		return nil, fmt.Errorf("in spWebGetContentReader: %s", sErr)
	}
	return resp.Body, nil
}

func (wdi *webDssImpl) spLibGetContentReader(ch string) (io.ReadCloser, error) {
	wdc := wdi.apc.GetConfig().(webDssClientConfig)
	proxy := wdc.libDss.(*ODss).proxy
	return proxy.spGetContentReader(ch)
}

func (wdi *webDssImpl) spGetContentReader(ch string) (io.ReadCloser, error) {
	if wdi.libApi {
		return wdi.spLibGetContentReader(ch)
	}
	return wdi.spWebGetContentReader(ch)
}

func (wdi *webDssImpl) doGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	return wdi.spGetContentReader(meta.Ch)
}

func (wdi *webDssImpl) queryContent(ch string) (exist bool, err error) {
	ex, err := cQueryContent(wdi.apc, ch)
	if err != nil {
		return false, fmt.Errorf("in queryContent: %v", err)
	}
	return ex.Exist, nil
}

func (wdi *webDssImpl) removeContent(ch string) error {
	err := cRemoveContent(wdi.apc, ch)
	if err != nil {
		return fmt.Errorf("in removeContent: %w", err)
	}
	return nil

}

func (wdi *webDssImpl) spClose() error {
	if !wdi.libApi {
		return nil
	}
	wdc := wdi.apc.GetConfig().(webDssClientConfig)
	proxy := wdc.libDss.(*ODss).proxy
	return proxy.close()
}

func (wdi *webDssImpl) dumpIndex() string {
	rdi, err := cDumpIndex(wdi.apc)
	if err != nil {
		return fmt.Errorf("in queryContent: %v", err).Error()
	}
	return strings.Join([]string{"Remote", rdi.Dump, "Local", wdi.getIndex().Dump()}, "\n")
}

func (wdi *webDssImpl) setAfs(tfs afero.Fs) { panic("inconsistent") }

func (wdi *webDssImpl) getAfs() afero.Fs { return appFs }

func copyMap[T any](dst map[string]T, src map[string]T) {
	for k, v := range src {
		dst[k] = v
	}
}

func (wdi *webDssImpl) spScanPhysicalStorageClient(checksum bool, sts *mSPS, sti StorageInfo, errs *ErrorCollector) {
	copyMap(sti.Path2Meta, sts.Sti.Path2Meta)
	copyMap(sti.Path2Content, sts.Sti.Path2Content)
	copyMap(sti.Path2CContent, sts.Sti.Path2CContent)
	copyMap(sti.ExistingCs, sts.Sti.ExistingCs)
	copyMap(sti.ExistingEcs, sts.Sti.ExistingEcs)
	copyMap(sti.Path2Error, sts.Sti.Path2Error)
	errs = &sts.Errs
}

func (wdi *webDssImpl) spAuditIndexFromRemote(sti StorageInfo, mai map[string][]AuditIndexInfo) error {
	appMai := func(k string, aii AuditIndexInfo) {
		if _, ok := mai[k]; !ok {
			mai[k] = []AuditIndexInfo{aii}
		}
		mai[k] = append(mai[k], aii)
	}

	rmetas, err := wdi.me.spLoadRemoteIndex(mai)
	if err != nil {
		return err
	}
	for k, mm := range rmetas {
		for t, m := range mm {
			var meta Meta
			if err = json.Unmarshal(m, &meta); err != nil {
				appMai(k, AuditIndexInfo{"RemoteInconsistent", err, t, m})
				continue
			}
			path := RemoveSlashIfNsIf(meta.Path, meta.IsNs)
			_, err, ok := wdi.index.loadMeta(path, t)
			if err != nil {
				appMai(k, AuditIndexInfo{"LocalError", err, t, m})
			}
			if !ok {
				appMai(k, AuditIndexInfo{"LocalMissing", nil, t, m})
			}
		}
	}
	_, lmetas, _, err := wdi.index.(*pIndex).loadInMemory()
	for k, mm := range lmetas {
		rmm, ok := rmetas[k]
		if !ok {
			appMai(k, AuditIndexInfo{"RemoteHashMissing", nil, 0, nil})
		}
		for t, m := range mm {
			var meta Meta
			if err = json.Unmarshal(m, &meta); err != nil {
				appMai(k, AuditIndexInfo{"LocalInconsistent", err, t, m})
				continue
			}
			_, ok := rmm[t]
			if !ok {
				appMai(k, AuditIndexInfo{"RemoteTimeMissing", nil, t, m})
			}
		}
	}
	return nil
}

func (wdi *webDssImpl) scanPhysicalStorage(checksum bool, sti StorageInfo, errs *ErrorCollector) {
	sts, err := cScanPhysicalStorage(wdi.apc, checksum)
	if err != nil {
		errs.Collect(fmt.Errorf("in scanPhysicalStorage: %v", err))
		return
	}
	// opportunity to decrypt if applicable
	wdi.me.spScanPhysicalStorageClient(checksum, sts, sti, errs)
}

func (wdi *webDssImpl) spLoadRemoteIndex(mai map[string][]AuditIndexInfo) (map[string]map[int64][]byte, error) {
	remx, err := cLoadIndex(wdi.apc)
	if err != nil {
		return map[string]map[int64][]byte{}, err
	}
	return remx.Metas, nil
}

func newWebDssProxy(config WebDssConfig, lsttime int64, aclusers []string, isClientEdss bool) (oDssProxy, HDss, error) {
	slsttime, _ := internal.Nano2SecNano(lsttime)
	impl := webDssImpl{isClientEdss: isClientEdss}
	var (
		dss HDss
		err error
	)
	if config.LibApi {
		impl.libApi = true
		if config.IsOlf {
			config.OlfCfg.DssBaseConfig.Encrypted = config.DssBaseConfig.Encrypted
			if dss, err = NewOlfDss(config.OlfCfg, slsttime, aclusers); err != nil {
				return nil, nil, err
			}
		} else if config.IsObs {
			config.ObsCfg.DssBaseConfig.Encrypted = config.DssBaseConfig.Encrypted
			if dss, err = NewObsDss(config.ObsCfg, slsttime, aclusers); err != nil {
				return nil, nil, err
			}
		} else if config.IsSmf {
			config.ObsCfg.DssBaseConfig.Encrypted = config.DssBaseConfig.Encrypted
			config.ObsCfg.GetS3Session = func() IS3Session {
				return NewS3sMockFs(config.ObsCfg.LocalPath, nil)
			}
			if dss, err = NewObsDss(config.ObsCfg, slsttime, aclusers); err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, fmt.Errorf("LibApi configuration is neither olf, obs or even smf")
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
func NewWebDss(config WebDssConfig, slsttime int64, aclusers []string) (HDss, error) {
	lsttime := slsttime * 1e9
	proxy, libDss, err := newWebDssProxy(config, lsttime, aclusers, false)
	if err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	wdcc := webDssClientConfig{WebDssConfig: config, libDss: libDss}
	if err := proxy.initialize(proxy, wdcc, lsttime, aclusers); err != nil {
		proxy.close()
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	if proxy.isRepoEncrypted() {
		proxy.close()
		return nil, fmt.Errorf("in NewWebDss: reposirory is encrypted")
	}
	var red plumber.Reducer = nil
	if config.ReducerLimit != 0 {
		red = plumber.NewReducer(config.ReducerLimit, 0)
	}
	proxy.setReducer(red)
	return &ODss{proxy: proxy}, nil
}
