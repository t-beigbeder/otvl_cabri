package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"io"
	"sort"
	"sync"
)

type EDssConfig struct {
	WebDssConfig
}

type eDssBaseImpl struct {
	oDssBaseImpl
}

type eDssImpl struct {
	webDssImpl
}

func (edi *eDssImpl) initialize(me oDssProxy, config interface{}, lsttime int64, aclusers []string) error {
	edc := config.(eDssClientConfig)
	if err := edi.webDssImpl.initialize(me, edc.webDssClientConfig, lsttime, aclusers); err != nil {
		return fmt.Errorf("in eDssImpl.initialize: %w", err)
	}
	if !edi.repoEncrypted {
		return fmt.Errorf("in eDssImpl.initialize: the repository is not encrypted")
	}
	return nil
}

func (edi *eDssImpl) isDuplicate(ch string) (bool, error) {
	return false, nil // encrypted content is never the same
}

func (edi *eDssImpl) isEncrypted() bool { return true }

func (edi *eDssImpl) defaultUser() string {
	for _, id := range edi.apc.GetConfig().(webDssClientConfig).identities {
		if id.Alias == "" {
			return id.PKey
		}
	}
	return ""
}

func (edi *eDssImpl) defaultAcl(acl []ACLEntry) []ACLEntry {
	if acl != nil {
		return acl
	}
	if edi.defaultUser() == "" {
		return nil
	}
	return []ACLEntry{{User: edi.defaultUser(), Rights: Rights{Read: true, Write: true}}}
}

func (edi *eDssImpl) secrets(users []string) (res []string) {
	if len(users) == 0 {
		users = []string{edi.defaultUser()}
	}
	for _, user := range users {
		for _, id := range edi.apc.GetConfig().(webDssClientConfig).identities {
			if id.PKey == user && id.Secret != "" {
				res = append(res, id.Secret)
			}
		}
	}
	return
}

func (edi *eDssImpl) pkeys(users []string) (res []string) {
	if len(users) == 0 {
		if edi.defaultUser() == "" {
			return
		}
		return []string{edi.defaultUser()}
	}
	for _, user := range users {
		for _, id := range edi.apc.GetConfig().(webDssClientConfig).identities {
			if id.PKey == user {
				res = append(res, user)
			}
		}
	}
	return
}

func (edi *eDssImpl) doUpdatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
	content, css, _ := internal.Ns2Content(children, "")
	sort.Strings(children)
	meta := Meta{
		Path:     npath + "/",
		Mtime:    mtime,
		Size:     int64(len(content)),
		Ch:       css,
		IsNs:     true,
		Children: children,
		ACL:      acl,
	}
	meta.EMId = uuid.New().String()
	mbs, itime, err := edi.getMetaBytes(meta)
	if err != nil {
		return fmt.Errorf("in doUpdatens: %w", err)
	}
	embs, err := EncryptMsg(string(mbs), edi.pkeys(Users(acl))...)
	if err != nil {
		return fmt.Errorf("in doUpdatens: %w", err)
	}
	if err := edi.storeMeta(meta.EMId, MIN_TIME, embs); err != nil {
		return fmt.Errorf("in doUpdatens: %w", err)
	}
	if err := edi.index.storeMeta(RemoveSlashIf(meta.Path), itime, mbs); err != nil {
		return fmt.Errorf("in doUpdatens: %w", err)
	}
	return nil
}

func (edi *eDssImpl) doGetMetaTimesFor(npath string) ([]int64, error) {
	return nil, nil // encrypted meta is only retrieved from local index
}

func (edi *eDssImpl) doGetMetaAt(npath string, time int64) (Meta, error) {
	mbs, err, ok := edi.index.loadMeta(npath, time)
	if err != nil || !ok {
		return Meta{}, err
	}
	var meta Meta
	if edi.metamockcbs != nil && edi.metamockcbs.MockUnmarshal != nil {
		err = edi.metamockcbs.MockUnmarshal(mbs, &meta)
	} else {
		err = json.Unmarshal(mbs, &meta)
	}
	if err != nil {
		return Meta{}, err
	}
	return meta, nil
}

func (edi *eDssImpl) spGetContentWriter(cwcbs contentWriterCbs, acl []ACLEntry) (io.WriteCloser, error) {
	var (
		eWcwc *WriteCloserWithCb
		eErr  error
		eSize int64
		eCh   string
		cErr  error
		cSize int64
		cCh   string
	)

	ecw, err := NewTempFileWriteCloserWithCb(edi.getAfs(), "", "ecw", func(err error, size int64, ch string, me *WriteCloserWithCb) error {
		eWcwc, eErr, eSize, eCh = me, err, size, ch
		_, _, _, _, _, _ = eErr, eSize, eCh, cErr, cSize, cCh
		outError := err
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter %w", err)
			return outError
		}
		mbs, emid, err := cwcbs.getMetaBytes(err, size, ch)
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter %w", err)
			return outError
		}
		meta, err := edi.decodeMeta(mbs)
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter %w", err)
			return outError
		}
		meta.Size = cSize
		meta.Ch = cCh
		meta.ECh = eCh
		mbs, itime, err := edi.getMetaBytes(meta)
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter %w", err)
			return outError
		}
		embs, err := EncryptMsg(string(mbs), edi.pkeys(Users(acl))...)
		if err != nil {
			outError = fmt.Errorf("in spGetContentWriter %w", err)
			return outError
		}
		cf := eWcwc.Underlying.(afero.File)
		if err := edi.pushContent(size, ch, embs, emid, cf); err != nil {
			outError = fmt.Errorf("in spGetContentWriter: %w", err)
			return outError
		}

		if err := edi.index.storeMeta(meta.Path, itime, mbs); err != nil {
			outError = fmt.Errorf("in spGetContentWriter: %w", err)
			return outError
		}

		return eErr
	})
	if err != nil {
		return nil, fmt.Errorf("in spGetContentWriter: %w", err)
	}
	wc, err := Encrypt(ecw, edi.pkeys(Users(acl))...)
	if err != nil {
		return nil, fmt.Errorf("in spGetContentWriter: %w", err)
	}
	return NewWriteCloserWithCb(wc, func(err error, size int64, ch string, me *WriteCloserWithCb) error {
		outError := err
		defer func() {
			if cwcbs.closeCb != nil {
				cwcbs.closeCb(outError, size, ch)
			}
		}()

		cErr, cSize, cCh = err, size, ch
		if cErr != nil {
			outError = cErr
			ecw.Close()
			return outError
		}
		if err := ecw.Close(); err != nil {
			outError = err
			return err
		}
		return nil
	}), nil
}

func (edi *eDssImpl) doGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	erc, err := edi.spGetContentReader(meta.ECh)
	if err != nil {
		return nil, fmt.Errorf("in doGetContentReader: %w", err)
	}
	crc, err := Decrypt(erc, edi.secrets(Users(meta.ACL))...)
	return NewReadCloserWithCb(crc, func() error {
		return erc.Close()
	})
}

func (edi *eDssImpl) xRemoveMeta(meta Meta) error {
	if err := cXRemoveMeta(edi.apc, meta.EMId, MIN_TIME); err != nil {
		return fmt.Errorf("in xRemoveMeta: %v", err)
	}
	ipath := RemoveSlashIfNsIf(meta.Path, meta.IsNs)
	return edi.index.removeMeta(ipath, meta.Itime)
}

func (edi *eDssImpl) spUpdateClient(cix Index, eud UpdatedData, isFull bool) error {
	udd := UpdatedData{Changed: map[string][]TimedMeta{}, Deleted: map[string]bool{}}
	for _, etms := range eud.Changed {
		for _, etm := range etms {
			smbs, err := DecryptMsg([]byte(etm.Bytes), edi.secrets(edi.aclusers)...)
			if err != nil {
				return fmt.Errorf("in spUpdateClient: %w", err)
			}
			meta, err := edi.decodeMeta([]byte(smbs))
			if err != nil {
				return fmt.Errorf("in spUpdateClient: %w", err)
			}
			mbs, err := json.Marshal(meta)
			if err != nil {
				return fmt.Errorf("in spUpdateClient: %w", err)
			}
			nph := internal.NameToHashStr32(RemoveSlashIfNsIf(meta.Path, meta.IsNs))
			tms, _ := udd.Changed[nph]
			tms = append(tms, TimedMeta{Time: meta.Itime, Bytes: string(mbs)})
			udd.Changed[nph] = tms
		}
	}
	return cix.updateData(udd, isFull)
}

func (edi *eDssImpl) decryptScannedStorage(checksum bool, sts *mSPS, sti StorageInfo, errs *ErrorCollector) {
	pathErr := func(path string, err error) {
		sti.Path2Error[path] = err
		errs.Collect(err)
	}

	eSti := sts.Sti
	mc2scan := map[string]Meta{}
	for epath, ebs := range eSti.Path2Meta {
		smbs, err := DecryptMsg(ebs, edi.secrets(edi.aclusers)...)
		if err != nil {
			pathErr(epath, err)
			continue
		}
		sti.Path2Meta[epath] = []byte(smbs)
		var meta Meta
		if err := json.Unmarshal([]byte(smbs), &meta); err != nil {
			pathErr(epath, err)
			continue
		}
		if meta.IsNs {
			continue
		}
		sti.ExistingCs[meta.Ch] = true
		sti.ExistingEcs[meta.ECh] = true
		mc2scan[epath] = meta
	}
	if !checksum {
		return
	}
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	lockPathErr := func(mn string, err error) {
		mx.Lock()
		defer mx.Unlock()
		pathErr(mn, err)
	}
	lockPath2CContent := func(pep, ch string) {
		mx.Lock()
		defer mx.Unlock()
		sti.Path2CContent[pep] = ch
	}
	doDecryptContent := func(pep string, pMeta Meta) {
		cr, err := edi.me.doGetContentReader(pMeta.Path, pMeta)
		if err != nil {
			lockPathErr(pep, err)
			return
		}
		defer cr.Close()
		ch, err := internal.ShaFrom(cr)
		if err != nil {
			lockPathErr(pep, err)
		}
		lockPath2CContent(pep, ch)
		if ch != pMeta.Ch {
			lockPathErr(pep, fmt.Errorf("%s (meta %s) cs %s loaded %s", pep, pMeta.Path, pMeta.Ch, ch))
			return
		}
	}
	for epath, meta := range mc2scan {
		wg.Add(1)
		go func(pep string, pMeta Meta) {
			defer wg.Done()
			if edi.reducer == nil {
				doDecryptContent(pep, pMeta)
			} else {
				edi.reducer.Launch(fmt.Sprintf("doDecryptContent-%s", pep), func() error {
					doDecryptContent(pep, pMeta)
					return nil
				})
			}
		}(epath, meta)
	}
	wg.Wait()
}

func (edi *eDssImpl) spScanPhysicalStorageClient(checksum bool, sts *mSPS, sti StorageInfo, errs *ErrorCollector) {
	copyMap(sti.Path2Error, sts.Sti.Path2Error)
	copyMap(sti.Path2Content, sts.Sti.Path2Content)
	errs = &sts.Errs
	edi.decryptScannedStorage(checksum, sts, sti, errs)
}

func (edi *eDssImpl) spLoadRemoteIndex(mai map[string][]AuditIndexInfo) (map[string]map[int64][]byte, error) {
	appMai := func(k string, aii AuditIndexInfo) {
		if _, ok := mai[k]; !ok {
			mai[k] = []AuditIndexInfo{aii}
		}
		mai[k] = append(mai[k], aii)
	}
	cmetas := map[string]map[int64][]byte{}
	remx, err := cLoadIndex(edi.apc)
	if err != nil {
		return cmetas, err
	}

	for k, emm := range remx.Metas {
		for t, em := range emm {
			sm, err := DecryptMsg(em, edi.secrets(edi.aclusers)...)
			if err != nil {
				appMai(k, AuditIndexInfo{"RemoteDecodeError", err, t, em})
				continue
			}
			m := []byte(sm)
			var meta Meta
			if err = json.Unmarshal(m, &meta); err != nil {
				appMai(k, AuditIndexInfo{"RemoteInconsistent", err, t, m})
				continue
			}
			npath := RemoveSlashIfNsIf(meta.Path, meta.IsNs)
			h := internal.NameToHashStr32(npath)
			_, ok := cmetas[h]
			if !ok {
				cmetas[h] = map[int64][]byte{}
			}
			cmetas[h][meta.Itime] = m
		}
	}
	return cmetas, nil
}

func (edi *eDssImpl) spReindex() (StorageInfo, *ErrorCollector) {
	sti := StorageInfo{
		Path2Meta:     map[string][]byte{},
		Path2Content:  map[string]string{},
		Path2CContent: map[string]string{},
		ExistingCs:    map[string]bool{},
		ExistingEcs:   map[string]bool{},
		Path2Error:    map[string]error{},
	}
	errs := &ErrorCollector{}
	if !edi.libApi {
		errs.Collect(fmt.Errorf("in reindex: cannot reindex remotely"))
		return sti, errs
	}
	wdc := edi.apc.GetConfig().(webDssClientConfig)
	sti, errs = wdc.libDss.Reindex()
	return sti, nil
}

func (edi *eDssImpl) decodeMetaPath(mp string) (hn string, itime int64) {
	panic("inconsistent")
}

func (edi *eDssImpl) openSession(aclusers []string) error {
	if err := cOpenSession(edi.apc, aclusers); err != nil {
		return fmt.Errorf("in openSession: %w", err)
	}
	return nil
}

func newEDssProxy(config EDssConfig, lsttime int64, aclusers []string) (oDssProxy, HDss, error) {
	wdp, dss, err := newWebDssProxy(config.WebDssConfig, lsttime, aclusers, true)
	if err != nil {
		return nil, nil, fmt.Errorf("in newEDssProxy: %w", err)
	}
	impl := eDssImpl{webDssImpl: *wdp.(*webDssImpl)}
	return &impl, dss, nil
}

// NewEDss opens a web or direct api client for an "object-storage" encrypted DSS (data storage system)
// config provides the object store specification
// lsttime if not zero is the upper time of entries retrieved in it
// aclusers if not nil is a List of ACL users for access check and for decryption
// returns a pointer to the ready to use DSS or an error if any occur
// If lsttime is not zero, access will be read-only
func NewEDss(config EDssConfig, slsttime int64, aclusers []string) (HDss, error) {
	lsttime := slsttime * 1e9
	config.Encrypted = true
	proxy, libDss, err := newEDssProxy(config, lsttime, aclusers)
	if err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	wdcc := eDssClientConfig{webDssClientConfig{WebDssConfig: config.WebDssConfig, libDss: libDss}}
	if err := proxy.initialize(proxy, wdcc, lsttime, aclusers); err != nil {
		proxy.close()
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	edi := proxy.(*eDssImpl)
	if err := edi.openSession(aclusers); err != nil {
		proxy.close()
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	var red plumber.Reducer = nil
	if config.ReducerLimit != 0 {
		red = plumber.NewReducer(config.ReducerLimit, 0)
	}
	proxy.setReducer(red)
	return &ODss{proxy: proxy}, nil
}
