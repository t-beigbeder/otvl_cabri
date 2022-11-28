package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"sort"
	"time"
)

type oDssBaseProxy interface {
	// internal functions directly mapped from Dss interface ones
	mkns(npath string, mtime int64, children []string, acl []ACLEntry) error
	updatens(npath string, mtime int64, children []string, acl []ACLEntry) error
	lsns(npath string) (children []string, err error)
	isDuplicate(ch string) (bool, error)
	getContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error)
	getContentReader(npath string) (io.ReadCloser, error)
	remove(npath string) error
	getMeta(npath string, getCh bool) (IMeta, error)
	getHistory(npath string, recursive bool) (map[string][]HistoryInfo, error)
	removeHistory(npath string, recursive, evaluate bool, start, end int64) (map[string][]HistoryInfo, error)
	setCurrentTime(time int64)
	setMetaMockCbs(cbs *MetaMockCbs)
	close() error
	getIndex() Index
	getRepoId() string
	isEncrypted() bool
	auditIndex() (map[string][]AuditIndexInfo, error)
	scanStorage() (StorageInfo, *ErrorCollector)
	// other
	doUpdatens(npath string, mtime int64, children []string, acl []ACLEntry) error
	setIndex(config DssBaseConfig, localPath string) error // to be called by oDssSpecificProxy.initialize
	isRepoEncrypted() bool
	defaultAcl(acl []ACLEntry) []ACLEntry
	doGetMetaTimesFor(npath string) ([]int64, error)
	decodeMeta(mbs []byte) (Meta, error)
	doGetMetaAt(npath string, time int64) (Meta, error)
	storeAndIndexMeta(npath string, time int64, bs []byte) error
	spUpdateClient(cix Index, data UpdatedData, isFull bool) error
	spScanPhysicalStorageClient(sts *mSPS, sti StorageInfo, errs *ErrorCollector)
}

type contentWriterCbs struct {
	closeCb      WriteCloserCb
	getMetaBytes func(iErr error, size int64, ch string) (mbs []byte, oErr error)
}

type oDssSpecificProxy interface {
	initialize(me oDssProxy, config interface{}, lsttime int64, aclusers []string) error // called on implementation instantiation (NewXxxDss)
	loadMeta(npath string, time int64) ([]byte, error)
	queryMetaTimes(npath string) (times []int64, err error)
	storeMeta(npath string, time int64, bs []byte) error
	removeMeta(npath string, time int64) error
	xRemoveMeta(npath string, time int64) error
	pushContent(size int64, ch string, mbs []byte, cf afero.File) error
	spGetContentWriter(cwcbs contentWriterCbs, acl []ACLEntry) (io.WriteCloser, error)
	spGetContentReader(ch string) (io.ReadCloser, error)
	doGetContentReader(npath string, meta Meta) (io.ReadCloser, error)
	queryContent(ch string) (exist bool, err error)
	spClose() error
	dumpIndex() string
	scanPhysicalStorage(sti StorageInfo, errs *ErrorCollector)
	// internal functions directly mapped from Dss interface ones
	setAfs(tfs afero.Fs)
	getAfs() afero.Fs
}

type oDssProxy interface {
	oDssBaseProxy
	oDssSpecificProxy
}

type ODss struct {
	proxy oDssProxy
}

func (ods ODss) Mkns(npath string, mtime int64, children []string, acl []ACLEntry) error {
	return ods.proxy.mkns(npath, mtime, children, ods.proxy.defaultAcl(acl))
}

func (ods ODss) Updatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
	return ods.proxy.updatens(npath, mtime, children, ods.proxy.defaultAcl(acl))
}

func (ods ODss) Lsns(npath string) (children []string, err error) {
	return ods.proxy.lsns(npath)
}

func (ods ODss) IsDuplicate(ch string) (bool, error) {
	return ods.proxy.isDuplicate(ch)
}

func (ods ODss) GetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	return ods.proxy.getContentWriter(npath, mtime, ods.proxy.defaultAcl(acl), cb)
}

func (ods ODss) GetContentReader(npath string) (io.ReadCloser, error) {
	return ods.proxy.getContentReader(npath)
}

func (ods ODss) Remove(npath string) error {
	return ods.proxy.remove(npath)
}

func (ods ODss) GetMeta(npath string, getCh bool) (IMeta, error) {
	return ods.proxy.getMeta(npath, getCh)
}

func (ods ODss) GetHistory(npath string, recursive bool) (map[string][]HistoryInfo, error) {
	return ods.proxy.getHistory(npath, recursive)
}

func (ods ODss) RemoveHistory(npath string, recursive, evaluate bool, start, end int64) (map[string][]HistoryInfo, error) {
	return ods.proxy.removeHistory(npath, recursive, evaluate, start, end)
}

func (ods ODss) SetCurrentTime(time int64) {
	ods.proxy.setCurrentTime(time)
}

func (ods ODss) SetMetaMockCbs(cbs *MetaMockCbs) {
	ods.proxy.setMetaMockCbs(cbs)
}

func (ods ODss) SetAfs(tfs afero.Fs) {
	ods.proxy.setAfs(tfs)
}

func (ods ODss) GetAfs() afero.Fs {
	return ods.proxy.getAfs()
}

func (ods ODss) Close() error { return ods.proxy.close() }

func (ods ODss) GetIndex() Index { return ods.proxy.getIndex() }

func (ods ODss) DumpIndex() string { return ods.proxy.dumpIndex() }

func (ods ODss) GetRepoId() string { return ods.proxy.getRepoId() }

func (ods ODss) IsEncrypted() bool { return ods.proxy.isEncrypted() }

func (ods ODss) IsRepoEncrypted() bool { return ods.proxy.isRepoEncrypted() }

func (ods ODss) AuditIndex() (map[string][]AuditIndexInfo, error) { return ods.proxy.auditIndex() }

func (ods ODss) ScanStorage() (StorageInfo, *ErrorCollector) { return ods.proxy.scanStorage() }

type oDssBaseImpl struct {
	me            oDssProxy
	lsttime       int64        // if not zero is the upper time of entries retrieved in it
	aclusers      []string     // if not nil List of ACL users to check access
	mockct        int64        // if not zero mock current time
	metamockcbs   *MetaMockCbs // if not nil callbacks for json marshal/unmarshal
	index         Index        // the DSS index, possibly nIndex which is a noop index
	repoId        string       // the DSS repoId or ""
	repoEncrypted bool         // repository is encrypted
}

func (odbi *oDssBaseImpl) metaTimesFor(npath string, allTimes bool) ([]int64, error) {
	times, err, ok := odbi.index.queryMetaTimes(npath)
	if err != nil {
		return nil, err
	}
	if !ok {
		if times, err = odbi.me.doGetMetaTimesFor(npath); err != nil {
			return nil, err
		}
	}

	if allTimes {
		sort.Slice(times, func(i, j int) bool {
			return times[i] < times[j]
		})
		return times, nil
	}
	var found = MIN_TIME
	for _, time := range times {
		if odbi.lsttime != 0 && time > odbi.lsttime {
			continue
		}
		if time > found {
			found = time
		}
	}
	if found == MIN_TIME {
		return nil, fmt.Errorf("no such entry: %s", npath)
	}
	return []int64{found}, err
}

func (odbi *oDssBaseImpl) metaTimeFor(npath string) (int64, error) {
	var times []int64
	var err error
	if times, err = odbi.metaTimesFor(npath, false); err != nil {
		return 0, err
	}
	return times[0], nil
}

func (odbi *oDssBaseImpl) doGetMeta(npath string) (Meta, error) {
	time, err := odbi.metaTimeFor(npath)
	if err != nil {
		return Meta{}, err
	}
	return odbi.me.doGetMetaAt(npath, time)
}

func (odbi *oDssBaseImpl) hasReadAcl(meta Meta) bool {
	readable := len(odbi.aclusers) == 0
	for _, user := range odbi.aclusers {
		for _, ace := range meta.GetAcl() {
			if ace.User != user {
				continue
			}
			if ace.Rights.Read {
				readable = true
				break
			}
		}
	}
	return readable
}

func (odbi *oDssBaseImpl) hasWriteAcl(meta Meta) bool {
	writable := len(odbi.aclusers) == 0
	for _, user := range odbi.aclusers {
		for _, ace := range meta.GetAcl() {
			if ace.User != user {
				continue
			}
			if ace.Rights.Write {
				writable = true
				break
			}
		}
	}
	return writable
}

func (odbi *oDssBaseImpl) getNsMeta(npath string) (Meta, error) {
	meta, err := odbi.doGetMeta(npath)
	if err != nil {
		return Meta{}, err
	}
	if !odbi.hasReadAcl(meta) {
		return Meta{}, fmt.Errorf("getNsMeta: %s access denied", npath)
	}
	return meta, nil
}

func (odbi *oDssBaseImpl) hasParent(npath string, isDir bool) (bool, error) {
	if npath == "" {
		return true, nil
	}
	parent := ufpath.Dir(npath)
	if parent == "." {
		parent = ""
	}
	ok, err := odbi.hasParent(parent, true)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	meta, err := odbi.getNsMeta(parent)
	if err != nil {
		return false, err
	}
	return isNpathIn(npath, isDir, meta.Children), nil
}

func (odbi *oDssBaseImpl) mkupns(npath string, mtime int64, children []string, acl []ACLEntry) error {
	if odbi.lsttime != 0 {
		return fmt.Errorf("read-only DSS")
	}
	err := checkMknsArgs(npath, children, acl)
	if err != nil {
		return err
	}
	ok, err := odbi.hasParent(npath, true)
	if err != nil {
		return fmt.Errorf("in Mkns/Updatens: %v", err)
	}
	if !ok {
		return fmt.Errorf("no such entry: %s", npath)
	}
	meta, err := odbi.doGetMeta(npath)
	if err == nil && !odbi.hasWriteAcl(meta) {
		return fmt.Errorf("in Mkns/Updatens: %s read-only", npath)
	}
	return odbi.me.doUpdatens(npath, mtime, children, acl)
}

func (odbi *oDssBaseImpl) mkns(npath string, mtime int64, children []string, acl []ACLEntry) error {
	return odbi.mkupns(npath, mtime, children, acl)
}

func (odbi *oDssBaseImpl) updatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
	return odbi.mkupns(npath, mtime, children, acl)
}

func (odbi *oDssBaseImpl) lsns(npath string) (children []string, err error) {
	if err := checkNpath(npath); err != nil {
		return nil, err
	}
	ok, err := odbi.hasParent(npath, true)
	if err != nil {
		return nil, fmt.Errorf("in Lsns: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("no such entry: %s", npath)
	}
	meta, err := odbi.doGetMeta(npath)
	if err != nil {
		return nil, fmt.Errorf("in Lsns: %w", err)
	}
	if err == nil && !odbi.hasReadAcl(meta) {
		return nil, fmt.Errorf("in Lsns: %s access denied", npath)
	}
	return meta.Children, nil
}

func (odbi *oDssBaseImpl) isDuplicate(ch string) (bool, error) {
	if odbi.me.isEncrypted() {
		panic("isEncrypted")
	}
	return odbi.me.queryContent(ch)
}

func (odbi *oDssBaseImpl) getContentWriter(npath string, mtime int64, acl []ACLEntry, closeCb WriteCloserCb) (io.WriteCloser, error) {
	if odbi.lsttime != 0 {
		return nil, fmt.Errorf("read-only DSS")
	}
	err := checkMkcontentArgs(npath, acl)
	if err != nil {
		return nil, err
	}
	ok, err := odbi.hasParent(npath, false)
	if err != nil {
		return nil, fmt.Errorf("in GetContentWriter: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("no such entry: %s", npath)
	}
	meta, err := odbi.doGetMeta(npath)
	if err == nil && !odbi.hasWriteAcl(meta) {
		return nil, fmt.Errorf("in GetContentWriter: %s read-only", npath)
	}
	return odbi.me.spGetContentWriter(contentWriterCbs{
		closeCb: closeCb,
		getMetaBytes: func(iErr error, size int64, ch string) (mbs []byte, oErr error) {
			if iErr == nil {
				meta := Meta{Path: npath, Mtime: mtime, Size: size, Ch: ch, ACL: acl}
				mbs, _, err := odbi.getMetaBytes(meta)
				if err != nil {
					return nil, fmt.Errorf("in getMetaBytes: %w", err)
				}
				return mbs, nil
			}
			return nil, fmt.Errorf("in getMetaBytes: %w", iErr)
		},
	}, acl)
}

func (odbi *oDssBaseImpl) getContentReader(npath string) (io.ReadCloser, error) {
	err := checkNpath(npath)
	if err != nil {
		return nil, err
	}
	ok, err := odbi.hasParent(npath, false)
	if err != nil {
		return nil, fmt.Errorf("in GetContentReader: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("no such entry: %s", npath)
	}
	meta, err := odbi.doGetMeta(npath)
	if err != nil {
		return nil, fmt.Errorf("in GetContentReader: %v", err)
	}
	if !odbi.hasReadAcl(meta) {
		return nil, fmt.Errorf("in GetContentReader: %s access denied", npath)
	}
	return odbi.me.doGetContentReader(npath, meta)
}

func (odbi *oDssBaseImpl) remove(npath string) error {
	if odbi.lsttime != 0 {
		return fmt.Errorf("read-only DSS")
	}
	isNS, ipath, err := checkNCpath(npath)
	if err != nil {
		return err
	}
	if ipath == "" {
		return fmt.Errorf("cannot remove root")
	}
	ok, err := odbi.hasParent(ipath, isNS)
	if err != nil {
		return fmt.Errorf("in Remove: %v", err)
	}
	if !ok {
		return fmt.Errorf("no such entry: %s", npath)
	}
	metac, err := odbi.doGetMeta(ipath)
	if err == nil && !odbi.hasWriteAcl(metac) {
		return fmt.Errorf("in Remove: %s read-only", npath)
	}

	parent := ufpath.Dir(ipath)
	if parent == "." {
		parent = ""
	}
	meta, err := odbi.doGetMeta(parent)
	if err == nil && !odbi.hasWriteAcl(meta) {
		return fmt.Errorf("in Remove: %s read-only", npath)
	}
	me := ufpath.Base(ipath)
	if isNS {
		me += "/"
	}
	uchildren := []string{}
	for _, child := range meta.Children {
		if child != me {
			uchildren = append(uchildren, child)
		}
	}
	return odbi.me.doUpdatens(parent, time.Now().Unix(), uchildren, meta.ACL)
}

func (odbi *oDssBaseImpl) getMeta(npath string, getCh bool) (IMeta, error) {
	isDir, ipath, err := checkNCpath(npath)
	if err != nil {
		return nil, err
	}
	ok, err := odbi.hasParent(ipath, isDir)
	if err != nil {
		return nil, fmt.Errorf("in GetMeta: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("no such entry: %s", npath)
	}
	meta, err := odbi.doGetMeta(ipath)
	if err == nil && !odbi.hasReadAcl(meta) {
		return nil, fmt.Errorf("in GetMeta: %s access denied", npath)
	}
	return meta, nil
}

type historyEntry struct {
	start int64
	end   int64
	meta  Meta
}

func (odbi *oDssBaseImpl) doGetRawHistory(npath string) ([]historyEntry, error) {
	mts, err := odbi.metaTimesFor(npath, true)
	if err != nil {
		return nil, fmt.Errorf("in doGetRawHistory: %v", err)
	}
	hes := make([]historyEntry, len(mts))
	for i, mt := range mts {
		hes[i] = historyEntry{start: mt, end: MAX_TIME}
		if i > 0 {
			hes[i-1].end = mt - 1
		}
		if hes[i].meta, err = odbi.me.doGetMetaAt(npath, mt); err != nil {
			return nil, fmt.Errorf("in doGetRawHistory: %v", err)
		}
	}
	return hes, nil
}

func (odbi *oDssBaseImpl) doGetRootHistory(recursive bool, res map[string][]historyEntry) error {
	hes, err := odbi.doGetRawHistory("")
	if err != nil {
		return fmt.Errorf("in doGetRootHistory: %v", err)
	}
	res[""] = hes
	children := map[string]bool{}
	for _, he := range hes {
		if recursive {
			for _, child := range he.meta.Children {
				children[child] = true
			}
		}
	}
	sort.Slice(hes, func(i, j int) bool {
		return hes[i].start < hes[j].start
	})
	for child, _ := range children {
		cIsNs, cIPath, _ := checkNCpath(child)
		if err := odbi.doGetHistory(cIPath, cIsNs, true, res); err != nil {
			return fmt.Errorf("in doGetRootHistory: %v", err)
		}
	}
	return nil
}

func (odbi *oDssBaseImpl) doGetHistory(npath string, isDir, recursive bool, res map[string][]historyEntry) error {
	if npath == "" {
		return odbi.doGetRootHistory(recursive, res)
	}
	hes, err := odbi.doGetRawHistory(npath)
	if err != nil {
		return fmt.Errorf("in doGetHistory: %v", err)
	}
	parent := ufpath.Dir(npath)
	if parent == "." {
		parent = ""
	}
	pRes := map[string][]historyEntry{}
	if err := odbi.doGetHistory(parent, true, false, pRes); err != nil {
		return fmt.Errorf("in doGetHistory: %v", err)
	}
	fhesM := map[string]historyEntry{}
	fhesEM := map[string]historyEntry{}
	for _, pHe := range pRes[AppendSlashIf(parent)] {
		if !isNpathIn(npath, isDir, pHe.meta.Children) {
			continue
		}
		for _, he := range hes {
			fhe := he
			if fhe.start < pHe.start {
				fhe.start = pHe.start
			}
			if pHe.end != MAX_TIME && (fhe.end == MAX_TIME || fhe.end > pHe.end) {
				fhe.end = pHe.end
			}
			if fhe.end != MAX_TIME && fhe.end <= fhe.start {
				continue
			}
			prevFhe, ok := fhesEM[fmt.Sprintf("%d", fhe.start-1)]
			if ok && prevFhe.meta.Equals(fhe.meta, true) {
				delete(fhesM, fmt.Sprintf("%d-%d", prevFhe.start, prevFhe.end))
				delete(fhesEM, fmt.Sprintf("%d", prevFhe.end))
				fhe.start = prevFhe.start
			}
			fhesM[fmt.Sprintf("%d-%d", fhe.start, fhe.end)] = fhe
			fhesEM[fmt.Sprintf("%d", fhe.end)] = fhe
		}
	}
	var fhes []historyEntry
	children := map[string]bool{}
	for _, v := range fhesM {
		fhes = append(fhes, v)
		if recursive {
			for _, child := range v.meta.Children {
				children[child] = true
			}
		}
	}
	sort.Slice(fhes, func(i, j int) bool {
		return fhes[i].start < fhes[j].start
	})
	me := npath
	if isDir {
		me = AppendSlashIf(npath)
	}
	res[me] = fhes
	for child, _ := range children {
		cIsNs, cIPath, _ := checkNCpath(ufpath.Join(npath, child))
		if err := odbi.doGetHistory(cIPath, cIsNs, true, res); err != nil {
			return fmt.Errorf("in doGetHistory: %v", err)
		}
	}
	return nil
}

func (odbi *oDssBaseImpl) getHistory(npath string, recursive bool) (map[string][]HistoryInfo, error) {
	isDir, ipath, err := checkNCpath(npath)
	if err != nil {
		return nil, err
	}
	iRes := map[string][]historyEntry{}
	if err := odbi.doGetHistory(ipath, isDir, recursive, iRes); err != nil {
		return nil, fmt.Errorf("in GetHistory: %v", err)
	}
	eRes := map[string][]HistoryInfo{}
	for np, hes := range iRes {
		his := make([]HistoryInfo, len(hes))
		for i, he := range hes {
			his[i] = HistoryInfo{Start: he.start, End: he.end, HMeta: he.meta}
		}
		eRes[np] = his
	}
	return eRes, err
}

func (odbi *oDssBaseImpl) doRemoveHistory(ipath string, isDir bool, recursive bool, evaluate bool, start int64, end int64, oRes map[string][]historyEntry) error {
	iRes := map[string][]historyEntry{}
	if err := odbi.doGetHistory(ipath, isDir, recursive, iRes); err != nil {
		return fmt.Errorf("in RemoveHistory: %v", err)
	}
	if start == 0 {
		start = MIN_TIME
	}
	if end == 0 {
		end = MAX_TIME
	}
	for path, hes := range iRes {
		oRes[path] = []historyEntry{}
		for _, he := range hes {
			if he.start > end || he.start < start || he.end < start || he.end > end {
				continue
			}
			oRes[path] = append(oRes[path], he)
			if !evaluate {
				if err := odbi.me.xRemoveMeta(ipath, he.meta.Itime); err != nil {
					return fmt.Errorf("in doRemoveHistory: %v", err)
				}
				if odbi.me.isEncrypted() {
					return fmt.Errorf("in doRemoveHistory: not yet implemented") // FIXME
				}
				if err := odbi.me.removeMeta(ipath, he.meta.Itime); err != nil {
					return fmt.Errorf("in doRemoveHistory: %v", err)
				}
			}
		}
		if len(oRes[path]) == 0 {
			delete(oRes, path)
		}
	}
	return nil
}

func (odbi *oDssBaseImpl) removeHistory(npath string, recursive, evaluate bool, start, end int64) (map[string][]HistoryInfo, error) {
	isDir, ipath, err := checkNCpath(npath)
	if err != nil {
		return nil, err
	}
	oRes := map[string][]historyEntry{}
	if err = odbi.doRemoveHistory(ipath, isDir, recursive, evaluate, start, end, oRes); err != nil {
		return nil, fmt.Errorf("in RemoveHistory: %v", err)
	}
	eRes := map[string][]HistoryInfo{}
	for np, hes := range oRes {
		his := make([]HistoryInfo, len(hes))
		for i, he := range hes {
			his[i] = HistoryInfo{Start: he.start, End: he.end, HMeta: he.meta}
		}
		eRes[np] = his
	}
	return eRes, err
}

func (odbi *oDssBaseImpl) setCurrentTime(time int64) { odbi.mockct = time }

func (odbi *oDssBaseImpl) setMetaMockCbs(cbs *MetaMockCbs) { odbi.metamockcbs = cbs }

func (odbi *oDssBaseImpl) close() error {
	err := odbi.me.spClose()
	if err != nil {
		odbi.index.Close()
		return err
	}
	return odbi.index.Close()
}

func (odbi *oDssBaseImpl) getIndex() Index { return odbi.index }

func (odbi *oDssBaseImpl) setIndex(baseConfig DssBaseConfig, localPath string) (err error) {
	if baseConfig.GetIndex == nil {
		odbi.index = NewNIndex()
	} else {
		odbi.index, err = baseConfig.GetIndex(baseConfig, localPath)
	}
	return
}

func (odbi *oDssBaseImpl) getRepoId() string { return odbi.repoId }

func (odbi *oDssBaseImpl) isEncrypted() bool { return false }

func (odbi *oDssBaseImpl) doUpdatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
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
	mbs, itime, err := odbi.getMetaBytes(meta)
	if err != nil {
		return fmt.Errorf("in doUpdatens: %w", err)
	}
	return odbi.storeAndIndexMeta(RemoveSlashIf(meta.Path), itime, mbs)
}

func (odbi *oDssBaseImpl) doAuditIndexFromStorage(sti StorageInfo, mai map[string][]AuditIndexInfo) error {
	appMai := func(k string, aii AuditIndexInfo) {
		if aii.Error == "" {
			aii.Error = "IndexMissing"
		}
		if _, ok := mai[k]; !ok {
			mai[k] = []AuditIndexInfo{aii}
		}
		mai[k] = append(mai[k], aii)
	}
	for path, bs := range sti.Path2Meta {
		dst := ufpath.Ext(path)
		if len(dst) == 0 {
			appMai(path, AuditIndexInfo{"Inconsistent", fmt.Errorf("bad file extension"), MIN_TIME, bs})
			continue
		}
		t, err := internal.Str16ToInt64(dst[1:])
		if err != nil {
			appMai(path, AuditIndexInfo{"Inconsistent", fmt.Errorf("bad file extension %v", err), MIN_TIME, bs})
			continue
		}
		var meta Meta
		if err := json.Unmarshal(bs, &meta); err != nil {
			appMai(path, AuditIndexInfo{"Inconsistent", err, t, bs})
			continue
		}
		ipath := RemoveSlashIfNsIf(meta.Path, meta.IsNs)
		bs2, err, ok := odbi.index.loadMeta(ipath, meta.Itime)
		if err != nil {
			appMai(path, AuditIndexInfo{"", err, t, bs})
			continue
		}
		if !ok {
			appMai(path, AuditIndexInfo{"", fmt.Errorf("no error"), t, bs})
			continue
		}
		var meta2 Meta
		if err := json.Unmarshal(bs2, &meta2); err != nil {
			appMai(path, AuditIndexInfo{"Inconsistent", err, t, bs2})
			continue
		}
		if !meta2.Equals(meta, true) {
			appMai(path, AuditIndexInfo{"", fmt.Errorf("%s (meta %s) meta %v loaded %v", path, ipath, meta, meta2), t, bs})
			continue
		}
	}
	return nil
}

func (odbi *oDssBaseImpl) doAuditIndexFromIndex(sti StorageInfo, mai map[string][]AuditIndexInfo) error {
	appMai := func(k string, aii AuditIndexInfo) {
		if aii.Error == "" {
			aii.Error = "StorageMissing"
		}
		if _, ok := mai[k]; !ok {
			mai[k] = []AuditIndexInfo{aii}
		}
		mai[k] = append(mai[k], aii)
	}

	_, metas, _, err := odbi.getIndex().(*pIndex).loadInMemory()
	smetas := sti.loadStoredInMemory()
	if err != nil {
		return fmt.Errorf("in doAuditIndexFromIndex: %v", err)
	}

	for k, mm := range metas {
		for t, m := range mm {
			var meta Meta
			if err = json.Unmarshal(m, &meta); err != nil {
				appMai(k, AuditIndexInfo{"Inconsistent", err, t, m})
				continue
			}
			if _, ok := smetas[k]; !ok {
				appMai(k, AuditIndexInfo{"", err, t, m})
				continue
			}
			if _, ok := smetas[k][t]; !ok {
				appMai(k, AuditIndexInfo{"", err, t, m})
				continue
			}
			var meta2 Meta
			if err := json.Unmarshal(smetas[k][t], &meta2); err != nil {
				appMai(k, AuditIndexInfo{"Inconsistent", err, t, smetas[k][t]})
				continue
			}
			if !meta2.Equals(meta, true) {
				appMai(k, AuditIndexInfo{"", fmt.Errorf("%s (meta %s) meta %v loaded %v", k, RemoveSlashIfNsIf(meta.Path, meta.IsNs), meta, meta2), t, m})
				continue
			}
		}
	}
	return nil
}

func (odbi *oDssBaseImpl) auditIndex() (map[string][]AuditIndexInfo, error) {
	if !odbi.getIndex().IsPersistent() {
		return nil, fmt.Errorf("in AuditIndex: not persistent")
	}
	mai, err := odbi.getIndex().(*pIndex).pRepair()
	if err != nil {
		return nil, fmt.Errorf("in AuditIndex: index analysis error %v", err)
	}
	if len(mai) > 0 {
		return mai, nil
	}
	sti, errs := odbi.scanStorage()
	if errs != nil {
		return nil, fmt.Errorf("in doAuditIndexFromStorage: %v", errs)
	}
	res := map[string][]AuditIndexInfo{}
	if err = odbi.doAuditIndexFromStorage(sti, res); err != nil {
		if err != nil {
			return nil, fmt.Errorf("in AuditIndex: %v", err)
		}
	}
	if err = odbi.doAuditIndexFromIndex(sti, res); err != nil {
		if err != nil {
			return nil, fmt.Errorf("in AuditIndex: %v", err)
		}
	}
	return res, nil
}

func (odbi *oDssBaseImpl) scanStorage() (StorageInfo, *ErrorCollector) {
	sti := StorageInfo{
		Path2Meta:    map[string][]byte{},
		Path2Content: map[string]string{},
		ExistingCs:   map[string]bool{},
		Path2Error:   map[string]error{},
	}
	errs := &ErrorCollector{}
	odbi.me.scanPhysicalStorage(sti, errs)
	pathErr := func(path string, err error) {
		sti.Path2Error[path] = err
		errs.Collect(err)
	}
	for path, bs := range sti.Path2Meta {
		if odbi.isRepoEncrypted() {
			// will check the following after decryption on client side
			// and also maybe we are there
			continue
		}
		var meta Meta
		if err := json.Unmarshal(bs, &meta); err != nil {
			pathErr(path, err)
			continue
		}
		ipath := meta.Path
		if meta.IsNs {
			ipath = RemoveSlashIf(ipath)
		}
		bs2, err := odbi.me.loadMeta(ipath, meta.Itime)
		if err != nil {
			pathErr(path, err)
			continue
		}
		if internal.BytesToSha256Str(bs) != internal.BytesToSha256Str(bs2) {
			pathErr(path, fmt.Errorf("%s (meta %s) sha %s loaded %s", path, ipath, internal.BytesToSha256Str(bs), internal.BytesToSha256Str(bs2)))
			continue
		}
		if meta.IsNs {
			continue
		}
		sti.ExistingCs[meta.Ch] = true
		cr, err := odbi.me.doGetContentReader(ipath, meta)
		if err != nil {
			pathErr(path, err)
			continue
		}
		ch := internal.ShaFrom(cr)
		cr.Close()
		if ch != meta.Ch {
			pathErr(path, fmt.Errorf("%s (meta %s) cs %s loaded %s", path, ipath, meta.Ch, ch))
			continue
		}
	}
	for path, ccs := range sti.Path2Content {
		if odbi.isRepoEncrypted() {
			// will check the following after decryption on client side
			// and also maybe we are there
			continue
		}
		_, ok := sti.ExistingCs[ccs]
		if !ok {
			pathErr(path, fmt.Errorf("%s (ch %s) is not used anymore", path, ccs))
			continue
		}
	}
	if len(*errs) > 0 {
		return StorageInfo{}, errs
	}
	return sti, nil
}

func (odbi *oDssBaseImpl) isRepoEncrypted() bool { return odbi.repoEncrypted }

func (odbi *oDssBaseImpl) defaultAcl(acl []ACLEntry) []ACLEntry { return acl }

func (odbi *oDssBaseImpl) doGetMetaTimesFor(npath string) (times []int64, err error) {
	if times, err = odbi.me.queryMetaTimes(npath); err != nil {
		return
	}
	if times != nil {
		if err = odbi.index.storeMetaTimes(npath, times); err != nil {
			return
		}
	}
	return
}

func (odbi *oDssBaseImpl) decodeMeta(mbs []byte) (meta Meta, err error) {
	if odbi.metamockcbs != nil && odbi.metamockcbs.MockUnmarshal != nil {
		err = odbi.metamockcbs.MockUnmarshal(mbs, &meta)
	} else {
		err = json.Unmarshal(mbs, &meta)
	}
	if err != nil {
		return
	}
	return
}

func (odbi *oDssBaseImpl) doGetMetaAt(npath string, time int64) (Meta, error) {
	bs, err, ok := odbi.index.loadMeta(npath, time)
	if err != nil {
		return Meta{}, err
	}
	if !ok {
		bs, err = odbi.me.loadMeta(npath, time)
		if err != nil {
			return Meta{}, err
		}
	}
	meta, err := odbi.decodeMeta(bs)
	if err != nil {
		return Meta{}, err
	}
	if err = odbi.index.storeMeta(npath, time, bs); err != nil {
		return Meta{}, err
	}
	return meta, nil
}

func (odbi *oDssBaseImpl) getMetaBytes(meta Meta) ([]byte, int64, error) {
	sort.Slice(meta.ACL, func(i, j int) bool {
		return meta.ACL[i].User < meta.ACL[j].User
	})
	time := time.Now().Unix()
	if odbi.mockct != 0 {
		time = odbi.mockct
		odbi.mockct += 1
	}
	meta.Itime = time
	var bs []byte
	var err error
	if odbi.metamockcbs != nil && odbi.metamockcbs.MockMarshal != nil {
		bs, err = odbi.metamockcbs.MockMarshal(meta)
	} else {
		bs, err = json.Marshal(meta)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("in getMetaBytes: %w", err)
	}
	return bs, time, nil
}

func (odbi *oDssBaseImpl) storeAndIndexMeta(npath string, time int64, bs []byte) error {
	if err := odbi.index.storeMeta(npath, time, bs); err != nil {
		return fmt.Errorf("in storeAndIndexMeta: %w", err)
	}
	if err := odbi.me.storeMeta(npath, time, bs); err != nil {
		return fmt.Errorf("in storeAndIndexMeta: %w", err)
	}
	return nil
}

func (odbi *oDssBaseImpl) spUpdateClient(cix Index, data UpdatedData, isFull bool) error {
	return cix.updateData(data, isFull)
}

func (odbi *oDssBaseImpl) spScanPhysicalStorageClient(sts *mSPS, sti StorageInfo, errs *ErrorCollector) {
	panic("inconsistent")
}
