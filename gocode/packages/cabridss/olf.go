package cabridss

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
)

type OlfConfig struct {
	DssBaseConfig
	Root string // filesystem root for the OLF DSS
	Size string // DSS size may be small, medium or large ("s", "m" or "l")
	// which enables storage of typically 200k, 4M or a huge number of files, with the additional cost of indexing storage
}

type oDssOlfImpl struct {
	oDssBaseImpl
	root string   // filesystem root for the OLF DSS
	size string   // DSS size "s", "m" or "l"
	afs  afero.Fs // if not nil mock filesystem
}

func (odoi *oDssOlfImpl) initialize(me oDssProxy, config interface{}, lsttime int64, aclusers []string) error {
	odoi.me = me
	odoi.lsttime = lsttime
	odoi.aclusers = aclusers
	olfConfig := config.(OlfConfig)
	var pc OlfConfig
	if err := LoadDssConfig(olfConfig.DssBaseConfig, &pc); err != nil {
		return fmt.Errorf("in Initialize: %w", err)
	}
	odoi.repoId = pc.RepoId
	odoi.repoEncrypted = pc.Encrypted
	odoi.root = olfConfig.Root
	odoi.size = pc.Size
	olfConfig.XImpl = pc.XImpl
	if err := odoi.setIndex(olfConfig.DssBaseConfig, ""); err != nil {
		return fmt.Errorf("in Initialize: %w", err)
	}
	return nil
}

func (odoi *oDssOlfImpl) loadMeta(npath string, mTime int64) ([]byte, error) {
	ht := sha256.Sum256([]byte(npath))
	mpath := fmt.Sprintf("%s.%s",
		ufpath.Join(odoi.root, "meta", internal.Sha256ToPath(ht[:], odoi.size)),
		internal.Int64ToStr16(mTime))
	mf, err := odoi.getAfs().Open(mpath)
	if err != nil {
		return nil, err
	}
	defer mf.Close()
	var b bytes.Buffer
	_, err = io.Copy(&b, mf)
	if err != nil {
		return nil, fmt.Errorf("read %s: %v", mpath, err)
	}
	return b.Bytes(), nil
}

func (odoi *oDssOlfImpl) queryMetaTimes(npath string) ([]int64, error) {
	ht := sha256.Sum256([]byte(npath))
	mprefix := ufpath.Join(odoi.root, "meta", internal.Sha256ToPath(ht[:], odoi.size))
	mdir := ufpath.Dir(mprefix)
	mname := ufpath.Base(mprefix)
	di, err := odoi.getAfs().Stat(mdir)
	if err != nil || !di.IsDir() {
		return nil, fmt.Errorf("no such directory: %s (err %v)", mdir, err)
	}
	df, err := odoi.getAfs().Open(mdir)
	if err != nil {
		return nil, err
	}
	defer df.Close()
	fil, err := df.Readdir(0)
	if err != nil {
		return nil, err
	}
	var times []int64
	for _, fi := range fil {
		if !strings.HasPrefix(fi.Name(), mname+".") {
			continue
		}
		suffix := ufpath.Ext(fi.Name())
		scanned, err := internal.Str16ToInt64(suffix[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid entry %s in %s (error %v)", fi.Name(), mdir, err)
		}
		times = append(times, scanned)
	}
	return times, nil
}

func (odoi *oDssOlfImpl) storeMeta(npath string, time int64, bs []byte) error {
	ht := sha256.Sum256([]byte(npath))
	mpath := fmt.Sprintf("%s.%s",
		ufpath.Join(odoi.root, "meta", internal.Sha256ToPath(ht[:], odoi.size)),
		internal.Int64ToStr16(time))
	if err := odoi.getAfs().MkdirAll(ufpath.Dir(mpath), 0o777); err != nil {
		return fmt.Errorf("in storeMeta: %w", err)
	}
	mf, err := odoi.getAfs().Create(mpath)
	if err != nil {
		return fmt.Errorf("in storeMeta: %w", err)
	}
	defer mf.Close()
	n, err := mf.Write(bs)
	if n != len(bs) || err != nil {
		_ = odoi.getAfs().Remove(mpath)
		return fmt.Errorf("in storeMeta: Write %s %d < %d error %v", mpath, n, len(bs), err)
	}
	return mf.Close()
}

func (odoi *oDssOlfImpl) removeMeta(npath string, time int64) error {
	ht := sha256.Sum256([]byte(npath))
	mpath := fmt.Sprintf("%s.%s",
		ufpath.Join(odoi.root, "meta", internal.Sha256ToPath(ht[:], odoi.size)),
		internal.Int64ToStr16(time))
	if err := odoi.getAfs().Remove(mpath); err != nil {
		return fmt.Errorf("in removeMeta: %w", err)
	}
	return nil
}

func (odoi *oDssOlfImpl) xRemoveMeta(meta Meta) error {
	ipath := RemoveSlashIfNsIf(meta.Path, meta.IsNs)
	return odoi.index.removeMeta(ipath, meta.Itime)
}

func (odoi *oDssOlfImpl) pushContent(size int64, ch string, mbs []byte, emid string, cf afero.File) error {
	cpath := ufpath.Join(odoi.root, "content", internal.Str32ToPath(ch, odoi.size))
	fi, err := odoi.getAfs().Stat(cpath)
	if fi != nil && fi.IsDir() {
		return fmt.Errorf("in pushContent: content path %s is a directory", cpath)
	}
	if err != nil && !os.IsExist(err) {
		if err = odoi.getAfs().MkdirAll(ufpath.Dir(cpath), 0o777); err != nil {
			return fmt.Errorf("in pushContent: %w", err)
		}
		if err = odoi.getAfs().Rename(cf.Name(), cpath); err != nil {
			return fmt.Errorf("in pushContent: %w", err)
		}
	}
	return nil
}

func (odoi *oDssOlfImpl) spGetContentWriter(cwcbs contentWriterCbs, acl []ACLEntry) (io.WriteCloser, error) {
	return NewTempFileWriteCloserWithCb(odoi.getAfs(), ufpath.Join(odoi.root, "tmp"), "cw", func(err error, size int64, ch string, wcwc *WriteCloserWithCb) error {
		if err != nil {
			return fmt.Errorf("in spGetContentWriter %w", err)
		}
		mbs, emid, err := cwcbs.getMetaBytes(err, size, ch)
		if err != nil {
			return fmt.Errorf("in spGetContentWriter %w", err)
		}
		if err = odoi.pushContent(size, ch, mbs, emid, wcwc.Underlying.(afero.File)); err != nil {
			return fmt.Errorf("in spGetContentWriter %w", err)
		}
		var (
			itime int64
			npath string
		)
		if !odoi.isRepoEncrypted() {
			meta, err := odoi.decodeMeta(mbs)
			if err != nil {
				return fmt.Errorf("in spGetContentWriter %w", err)
			}
			itime = meta.Itime
			npath = RemoveSlashIfNsIf(meta.Path, meta.IsNs)
		} else {
			itime = MIN_TIME
			npath = emid
		}
		if err = odoi.me.storeAndIndexMeta(npath, itime, mbs); err != nil {
			return fmt.Errorf("in spGetContentWriter %w", err)
		}
		return nil
	})
}

func (odoi *oDssOlfImpl) spGetContentReader(ch string) (io.ReadCloser, error) {
	cpath := ufpath.Join(odoi.root, "content", internal.Str32ToPath(ch, odoi.size))
	cf, err := odoi.getAfs().Open(cpath)
	if err != nil {
		return nil, fmt.Errorf("in GetContentReader: %w", err)
	}
	return cf, nil
}

func (odoi *oDssOlfImpl) doGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	return odoi.spGetContentReader(meta.Ch)
}

func (odoi *oDssOlfImpl) queryContent(ch string) (bool, error) {
	cpath := ufpath.Join(odoi.root, "content", internal.Str32ToPath(ch, odoi.size))
	_, err := odoi.getAfs().Stat(cpath)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (odoi *oDssOlfImpl) removeContent(ch string) error {
	cpath := ufpath.Join(odoi.root, "content", internal.Str32ToPath(ch, odoi.size))
	if err := odoi.getAfs().Remove(cpath); err != nil {
		return fmt.Errorf("in removeContent: %w", err)
	}
	return nil
}

func (odoi *oDssOlfImpl) spClose() error { return nil }

func (odoi *oDssOlfImpl) dumpIndex() string { return odoi.index.Dump() }

func (odoi *oDssOlfImpl) setAfs(tfs afero.Fs) { odoi.afs = tfs }

func (odoi *oDssOlfImpl) getAfs() afero.Fs {
	if odoi.afs != nil {
		return odoi.afs
	}
	return appFs
}

func (odoi *oDssOlfImpl) scanMetaDir(path string, sti StorageInfo, errs *ErrorCollector) {
	pathErr := func(path string, err error) {
		sti.Path2Error[path] = err
		errs.Collect(err)
	}
	df, err := odoi.getAfs().Open(path)
	if err != nil {
		pathErr(path, err)
		return
	}
	defer df.Close()
	fil, err := df.Readdir(0)
	if err != nil {
		pathErr(path, err)
		return
	}
	for _, fi := range fil {
		cPath := ufpath.Join(path, fi.Name())
		if fi.IsDir() {
			odoi.scanMetaDir(cPath, sti, errs)
			continue
		}
		if fi.Size() >= MAX_META_SIZE {
			pathErr(cPath, fmt.Errorf("%s size %d", cPath, fi.Size()))
			continue
		}
		bs := make([]byte, fi.Size())
		cdf, err := odoi.getAfs().Open(cPath)
		if err != nil {
			pathErr(cPath, err)
			return
		}
		n, err := cdf.Read(bs)
		cdf.Close()
		if n != int(fi.Size()) || err != nil {
			pathErr(cPath, fmt.Errorf("%s read %d err %v", cPath, n, err))
			continue
		}
		sti.Path2Meta[cPath] = bs
		hn, _ := internal.Path2Str32(cPath, odoi.size)
		suffix := ufpath.Ext(cPath)
		t, _ := internal.Str16ToInt64(suffix[1:])
		sti.Path2HnIt[cPath] = SIHnIt{
			Hn: hn,
			It: t,
		}
	}
	return
}

func (odoi *oDssOlfImpl) scanContentDir(path string, checksum bool, sti StorageInfo, errs *ErrorCollector) {
	pathErr := func(path string, err error) {
		sti.Path2Error[path] = err
		errs.Collect(err)
	}
	df, err := odoi.getAfs().Open(path)
	if err != nil {
		pathErr(path, err)
		return
	}
	defer df.Close()
	fil, err := df.Readdir(0)
	if err != nil {
		pathErr(path, err)
		return
	}
	for _, fi := range fil {
		cPath := ufpath.Join(path, fi.Name())
		if fi.IsDir() {
			odoi.scanContentDir(cPath, false, sti, errs)
			continue
		}
		relPath := cPath[strings.LastIndex(cPath, "/content/")+len("/content") : len(cPath)]
		cch := strings.Join(strings.Split(relPath, "/"), "")
		sti.Path2Content[cPath] = cch
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
	doScanContentOlf := func(pcp, pcch string) {
		cr, err := odoi.getAfs().Open(pcp)
		if err != nil {
			lockPathErr(path, err)
			return
		}
		ch := ""
		ch, err = internal.ShaFrom(cr)
		if err != nil {
			cr.Close()
			lockPathErr(path, err)
			return
		}
		cr.Close()
		if ch != pcch {
			lockPathErr(path, fmt.Errorf("in doScanContentOlf: content checksum %s differs from path %s", ch, pcp))
		}
	}
	for cPath, cch := range sti.Path2Content {
		wg.Add(1)
		go func(pcp, pcch string) {
			defer wg.Done()
			if odoi.reducer == nil {
				doScanContentOlf(pcp, pcch)
			} else {
				odoi.reducer.Launch(fmt.Sprintf("doScanContentOlf-%s", pcch), func() error {
					doScanContentOlf(pcp, pcch)
					return nil
				})
			}
		}(cPath, cch)
	}
	wg.Wait()
	return
}

func (odoi *oDssOlfImpl) scanPhysicalStorage(checksum bool, sti StorageInfo, errs *ErrorCollector) {
	odoi.scanMetaDir(ufpath.Join(odoi.root, "meta"), sti, errs)
	odoi.scanContentDir(ufpath.Join(odoi.root, "content"), checksum, sti, errs)
}

func newOlfProxy() oDssProxy {
	return &oDssOlfImpl{}
}

// NewOlfDss opens an "object-storage-like files" DSS (data storage system)
// config provides the object store specification
// lsttime if not zero is the upper time of entries retrieved in it
// aclusers if not nil is a List of ACL users for access check
// returns a pointer to the ready to use DSS or an error if any occur
// If lsttime is not zero, access will be read-only
func NewOlfDss(config OlfConfig, slsttime int64, aclusers []string) (HDss, error) {
	lsttime := slsttime * 1e9
	err := checkDir(config.Root)
	if err != nil {
		return nil, fmt.Errorf("in NewOlfDss: %w", err)
	}
	err = checkDir(ufpath.Join(config.Root, "meta"))
	if err != nil {
		return nil, fmt.Errorf("in NewOlfDss: %w", err)
	}
	err = checkDir(ufpath.Join(config.Root, "content"))
	if err != nil {
		return nil, fmt.Errorf("in NewOlfDss: %w", err)
	}
	proxy := newOlfProxy()
	if err := proxy.initialize(proxy, config, lsttime, aclusers); err != nil {
		return nil, fmt.Errorf("in NewOlfDss: %w", err)
	}
	if proxy.isRepoEncrypted() != config.Encrypted {
		if proxy.isRepoEncrypted() {
			proxy.close()
			return nil, fmt.Errorf("in NewOlfDss: repository is encrypted")
		} else {
			proxy.close()
			return nil, fmt.Errorf("in NewOlfDss: repository is not encrypted")
		}
	}
	var red plumber.Reducer = nil
	if config.ReducerLimit != 0 {
		red = plumber.NewReducer(config.ReducerLimit, 0)
	}
	proxy.setReducer(red)
	return &ODss{proxy: proxy}, nil
}

// CreateOlfDss creates an "object-storage-like files" DSS data storage system
// config provides the object store specification
// returns a pointer to the ready to use DSS or an error if any occur
func CreateOlfDss(config OlfConfig) (HDss, error) {
	root := config.Root
	size := config.Size
	err := checkDir(root)
	if err != nil {
		return nil, fmt.Errorf("in CreateOlfDss: %w", err)
	}
	if size != "s" && size != "m" && size != "l" {
		return nil, fmt.Errorf("in CreateOlfDss: incorrect size type %s", size)
	}
	if config.LocalPath == "" {
		return nil, fmt.Errorf("in CreateOlfDss: please provide a LocalPath")
	}
	config.RepoId = uuid.New().String()
	if err := SaveDssConfig(config.DssBaseConfig, config); err != nil {
		return nil, fmt.Errorf("in CreateObsDss: %w", err)
	}

	err = os.Mkdir(ufpath.Join(root, "meta"), 0o777)
	if err != nil {
		return nil, fmt.Errorf("in CreateOlfDss: %w", err)
	}
	err = os.Mkdir(ufpath.Join(root, "content"), 0o777)
	if err != nil {
		return nil, fmt.Errorf("in CreateOlfDss: %w", err)
	}
	err = os.Mkdir(ufpath.Join(root, "tmp"), 0o777)
	if err != nil {
		return nil, fmt.Errorf("in CreateOlfDss: %w", err)
	}
	return NewOlfDss(config, 0, nil)
}
