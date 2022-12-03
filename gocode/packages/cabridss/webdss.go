package cabridss

import (
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
	if err = wdi.me.spUpdateClient(cix, udd.UpdatedData, !mIed.ClientIsKnown); err != nil {
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

func (wdi *webDssImpl) xRemoveMeta(npath string, time int64) error {
	if err := cXRemoveMeta(wdi.apc, npath, time); err != nil {
		return fmt.Errorf("in xRemoveMeta: %v", err)
	}
	return wdi.index.removeMeta(npath, time)
}

func (wdi *webDssImpl) webPushContent(size int64, ch string, mbs []byte, cf afero.File) error {
	jsonArgs, err := json.Marshal(mPushContentIn{Size: size, Ch: ch, Mbs: mbs})
	if err != nil {
		return fmt.Errorf("in webPushContent: %w", err)
	}
	lja := internal.Int64ToStr16(int64(len(jsonArgs)))
	file, err := os.Open(cf.Name())
	if err != nil {
		return fmt.Errorf("in webPushContent: %w", err)
	}
	hdler := webContentWriterHandler{header: make([]byte, 16+len(jsonArgs)), file: file}
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

func (wdi *webDssImpl) libPushContent(size int64, ch string, mbs []byte, cf afero.File) error {
	wdc := wdi.apc.GetConfig().(webDssClientConfig)
	ccf, err := os.Open(cf.Name())
	if err != nil {
		return fmt.Errorf("in libPushContent: %v", err)
	}
	defer ccf.Close()
	proxy := wdc.libDss.(*ODss).proxy
	wter, err := proxy.spGetContentWriter(contentWriterCbs{
		getMetaBytes: func(iErr error, size int64, ch string) ([]byte, error) {
			return mbs, nil
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

func (wdi *webDssImpl) pushContent(size int64, ch string, mbs []byte, cf afero.File) error {
	var err error
	if wdi.libApi {
		err = wdi.libPushContent(size, ch, mbs, cf)
	} else {
		err = wdi.webPushContent(size, ch, mbs, cf)
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
		mbs, err := cwcbs.getMetaBytes(err, size, ch)
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
		if err := wdi.pushContent(size, ch, mbs, cf); err != nil {
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

func (wdi *webDssImpl) spScanPhysicalStorageClient(sts *mSPS, sti StorageInfo, errs *ErrorCollector) {
	copyMap(sti.Path2Meta, sts.Sti.Path2Meta)
	copyMap(sti.Path2Content, sts.Sti.Path2Content)
	copyMap(sti.ExistingCs, sts.Sti.ExistingCs)
	copyMap(sti.Path2Error, sts.Sti.Path2Error)
	errs = &sts.Errs
}

func (wdi *webDssImpl) scanPhysicalStorage(sti StorageInfo, errs *ErrorCollector) {
	sts, err := cScanPhysicalStorage(wdi.apc)
	if err != nil {
		errs.Collect(fmt.Errorf("in scanPhysicalStorage: %v", err))
		return
	}
	// opportunity to decrypt if applicable
	wdi.me.spScanPhysicalStorageClient(sts, sti, errs)
}

func newWebDssProxy(config WebDssConfig, lsttime int64, aclusers []string, isClientEdss bool) (oDssProxy, HDss, error) {
	impl := webDssImpl{isClientEdss: isClientEdss}
	var (
		dss HDss
		err error
	)
	if config.LibApi {
		impl.libApi = true
		if config.IsOlf {
			config.OlfCfg.DssBaseConfig.Encrypted = config.DssBaseConfig.Encrypted
			if dss, err = NewOlfDss(config.OlfCfg, lsttime, aclusers); err != nil {
				return nil, nil, err
			}
		} else if config.IsObs {
			config.OlfCfg.DssBaseConfig.Encrypted = config.DssBaseConfig.Encrypted
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
	proxy, libDss, err := newWebDssProxy(config, lsttime, aclusers, false)
	if err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	wdcc := webDssClientConfig{WebDssConfig: config, libDss: libDss}
	if err := proxy.initialize(proxy, wdcc, lsttime, aclusers); err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	if proxy.isRepoEncrypted() {
		return nil, fmt.Errorf("in NewWebDss: reposirory is encrypted")
	}
	return &ODss{proxy: proxy}, nil
}
