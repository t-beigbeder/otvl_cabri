package cabridss

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"strings"
	"sync"
)

type ObsConfig struct {
	DssBaseConfig
	Endpoint     string            `json:"endpoint"`  // AWS S3 or Openstack Swift endpoint, eg "https://s3.gra.cloud.ovh.net"
	Region       string            `json:"region"`    // AWS S3  or Openstack Swift region, eg "GRA"
	AccessKey    string            `json:"accessKey"` // AWS S3 access key (Openstack Swift must generate it)
	SecretKey    string            `json:"secretKey"` // AWS S3 secret key (Openstack Swift must generate it)
	Container    string            `json:"container"` // AWS S3 bucket or Openstack Swift container
	GetS3Session func() IS3Session `json:"-"`         // if not nil enables to set a mock S3 implementation
}

type oDssObjImpl struct {
	oDssBaseImpl
	is3 IS3Session
}

func (odoi *oDssObjImpl) initialize(me oDssProxy, config interface{}, lsttime int64, aclusers []string) error {
	odoi.me = me
	odoi.lsttime = lsttime
	odoi.aclusers = aclusers
	obsConfig := config.(ObsConfig)
	if obsConfig.LocalPath != "" {
		var pc ObsConfig
		if err := LoadDssConfig(obsConfig.DssBaseConfig, &pc); err != nil {
			return fmt.Errorf("in Initialize: %w", err)
		}
		if obsConfig.Endpoint == "" {
			obsConfig.Endpoint = pc.Endpoint
		}
		if obsConfig.Region == "" {
			obsConfig.Region = pc.Region
		}
		if obsConfig.AccessKey == "" {
			obsConfig.AccessKey = pc.AccessKey
		}
		if obsConfig.SecretKey == "" {
			obsConfig.SecretKey = pc.SecretKey
		}
		if obsConfig.Container == "" {
			obsConfig.Container = pc.Container
		}
		odoi.repoId = pc.RepoId
		odoi.repoEncrypted = pc.Encrypted
		obsConfig.XImpl = pc.XImpl
	}
	if err := odoi.setIndex(obsConfig.DssBaseConfig, obsConfig.LocalPath); err != nil {
		return fmt.Errorf("in Initialize: %w", err)
	}
	if obsConfig.GetS3Session == nil {
		odoi.is3 = &s3Session{config: obsConfig}
	} else {
		odoi.is3 = obsConfig.GetS3Session()
	}
	return odoi.is3.Initialize()
}

func (odoi *oDssObjImpl) loadMeta(npath string, mTime int64) ([]byte, error) {
	return odoi.is3.Get(fmt.Sprintf("meta-%s.%s", internal.NameToHashStr32(npath), internal.Int64ToStr16(mTime)))
}

func (odoi *oDssObjImpl) queryMetaTimes(npath string) ([]int64, error) {
	mprefix := fmt.Sprintf("meta-%s", internal.NameToHashStr32(npath))
	mns, err := odoi.is3.List(mprefix)
	if err != nil {
		return nil, err
	}
	var times []int64
	for _, mn := range mns {
		suffix := ufpath.Ext(mn)
		scanned, err := internal.Str16ToInt64(suffix[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid entry %s (error %v)", mn, err)
		}
		times = append(times, scanned)
	}
	return times, nil
}

func (odoi *oDssObjImpl) storeMeta(npath string, time int64, bs []byte) error {
	return odoi.is3.Put(fmt.Sprintf("meta-%s.%s", internal.NameToHashStr32(npath), internal.Int64ToStr16(time)), bs)
}

func (odoi *oDssObjImpl) removeMeta(npath string, time int64) error {
	return odoi.is3.Delete(fmt.Sprintf("meta-%s.%s", internal.NameToHashStr32(npath), internal.Int64ToStr16(time)))
}

func (odoi *oDssObjImpl) xRemoveMeta(meta Meta) error {
	ipath := RemoveSlashIfNsIf(meta.Path, meta.IsNs)
	return odoi.index.removeMeta(ipath, meta.Itime)
}

func (odoi *oDssObjImpl) pushContent(size int64, ch string, mbs []byte, emid string, cf afero.File) error {
	cName := fmt.Sprintf("content-%s", ch)
	lr, _ := odoi.is3.List(cName)
	if len(lr) == 0 {
		r, err := os.Open(cf.Name())
		if err != nil {
			return fmt.Errorf("in pushContent: %w", err)
		}
		defer r.Close()
		if err = odoi.is3.Upload(cName, r); err != nil {
			return fmt.Errorf("in pushContent: %w", err)
		}
	}
	return nil
}

func (odoi *oDssObjImpl) spGetContentWriter(cwcbs contentWriterCbs, acl []ACLEntry) (io.WriteCloser, error) {
	return NewTempFileWriteCloserWithCb(odoi.getAfs(), "", "cw", func(err error, size int64, ch string, wcwc *WriteCloserWithCb) error {
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

func (odoi *oDssObjImpl) spGetContentReader(ch string) (io.ReadCloser, error) {
	return odoi.is3.Download(fmt.Sprintf("content-%s", ch))
}

func (odoi *oDssObjImpl) doGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	return odoi.spGetContentReader(meta.Ch)
}

func (odoi *oDssObjImpl) queryContent(ch string) (bool, error) {
	cn := fmt.Sprintf("content-%s", ch)
	lr, err := odoi.is3.List(cn)
	if err != nil {
		return false, err
	}
	if len(lr) != 1 || lr[0] != cn {
		return false, fmt.Errorf("in queryContent: %v", lr)
	}
	return true, nil
}

func (odoi *oDssObjImpl) removeContent(ch string) error {
	cn := fmt.Sprintf("content-%s", ch)
	if err := odoi.is3.Delete(cn); err != nil {
		return fmt.Errorf("in removeContent: %w", err)
	}
	return nil
}

func (odoi *oDssObjImpl) spClose() error { return nil }

func (odoi *oDssObjImpl) dumpIndex() string { return odoi.index.Dump() }

func (odoi *oDssObjImpl) setAfs(tfs afero.Fs) { panic("inconsistent") }

func (odoi *oDssObjImpl) getAfs() afero.Fs { return appFs }

func (odoi *oDssObjImpl) scanMetaObjs(sti StorageInfo, errs *ErrorCollector) {
	mx := sync.Mutex{}
	doScanMetaObj := func(mn string) ([]byte, error) {
		suffix := ufpath.Ext(mn)
		if len(suffix) == 0 {
			return nil, fmt.Errorf("no suffix")
		}
		t, err := internal.Str16ToInt64(suffix[1:])
		if err != nil {
			return nil, err
		}
		_ = t
		return odoi.is3.Get(mn)
	}
	pathErr := func(mn string, err error) {
		mx.Lock()
		defer mx.Unlock()
		sti.Path2Error[mn] = err
		errs.Collect(err)
	}
	pathOk := func(mn string, bs []byte) {
		mx.Lock()
		defer mx.Unlock()
		sti.Path2Meta[mn] = bs
		hne := mn[len("meta-"):]
		hn := strings.Split(hne, ".")[0]
		suffix := strings.Split(hne, ".")[1]
		t, _ := internal.Str16ToInt64(suffix)
		sti.Path2HnIt[mn] = SIHnIt{
			Hn: hn,
			It: t,
		}
	}
	mns, err := odoi.is3.List("meta-")
	if err != nil {
		pathErr("meta-", err)
		return
	}
	wg := sync.WaitGroup{}
	for _, mn := range mns {
		wg.Add(1)
		go func(mni string) {
			defer wg.Done()
			if odoi.reducer == nil {
				bs, err := doScanMetaObj(mni)
				if err != nil {
					pathErr(mni, err)
				} else {
					pathOk(mni, bs)
				}
			} else {
				odoi.reducer.Launch(fmt.Sprintf("doScanMetaObj-%s", mni), func() error {
					bs, err := doScanMetaObj(mni)
					if err != nil {
						pathErr(mni, err)
					} else {
						pathOk(mni, bs)
					}
					return nil
				})
			}
		}(mn)
	}
	wg.Wait()
}

func (odoi *oDssObjImpl) scanContentObjs(checksum bool, sti StorageInfo, errs *ErrorCollector) {
	pathErr := func(path string, err error) {
		sti.Path2Error[path] = err
		errs.Collect(err)
	}
	cns, err := odoi.is3.List("content-")
	if err != nil {
		pathErr("content-", err)
		return
	}
	for _, cn := range cns {
		sti.Path2Content[cn] = cn[len("content-"):]
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
	doScanContentObs := func(pcn string) {
		cr, err := odoi.is3.Download(pcn)
		if err != nil {
			lockPathErr(pcn, err)
			return
		}
		ch := ""
		ch, err = internal.ShaFrom(cr)
		if err != nil {
			cr.Close()
			lockPathErr(pcn, err)
			return
		}
		cr.Close()
		if ch != pcn[len("content-"):] {
			lockPathErr(pcn, fmt.Errorf("in doScanContentObs: content checksum %s differs from path %s", ch, pcn))
		}
	}
	for cn, _ := range sti.Path2Content {
		wg.Add(1)
		go func(pcn string) {
			defer wg.Done()
			if odoi.reducer == nil {
				doScanContentObs(pcn)
			} else {
				odoi.reducer.Launch(fmt.Sprintf("doScanContentObs-%s", cn[len("content-"):]), func() error {
					doScanContentObs(pcn)
					return nil
				})
			}
		}(cn)
	}
	wg.Wait()
	return
}

func (odoi *oDssObjImpl) scanPhysicalStorage(checksum bool, sti StorageInfo, errs *ErrorCollector) {
	odoi.scanMetaObjs(sti, errs)
	odoi.scanContentObjs(checksum, sti, errs)
}

func newObsProxy() oDssProxy {
	return &oDssObjImpl{}
}

// NewObsDss opens an "object-storage" DSS (data storage system)
// config provides the object store specification
// lsttime if not zero is the upper time of entries retrieved in it
// aclusers if not nil is a List of ACL users for access check
// returns a pointer to the ready to use DSS or an error if any occur
// If lsttime is not zero, access will be read-only
func NewObsDss(config ObsConfig, slsttime int64, aclusers []string) (HDss, error) {
	lsttime := slsttime * 1e9
	proxy := newObsProxy()
	if err := proxy.initialize(proxy, config, lsttime, aclusers); err != nil {
		return nil, fmt.Errorf("in NewObsDss: %w", err)
	}
	if proxy.isRepoEncrypted() != config.Encrypted {
		if proxy.isRepoEncrypted() {
			proxy.close()
			return nil, fmt.Errorf("in NewObsDss: repository is encrypted")
		} else {
			proxy.close()
			return nil, fmt.Errorf("in NewObsDss: repository is not encrypted")
		}
	}
	var red plumber.Reducer = nil
	if config.ReducerLimit != 0 {
		red = plumber.NewReducer(config.ReducerLimit, 0)
	}
	proxy.setReducer(red)
	return &ODss{proxy: proxy}, nil
}

// CreateObsDss creates an "object-storage" DSS (data storage system)
// config provides the object store specification
// returns a pointer to the ready to use DSS or an error if any occur
func CreateObsDss(config ObsConfig) (HDss, error) {
	if config.LocalPath != "" {
		config.RepoId = uuid.New().String()
		if err := SaveDssConfig(config.DssBaseConfig, config); err != nil {
			return nil, fmt.Errorf("in CreateObsDss: %w", err)
		}
		ix, err := config.DssBaseConfig.GetIndex(config.DssBaseConfig, config.LocalPath)
		if err != nil {
			return nil, fmt.Errorf("in CreateObsDss: %w", err)
		}
		if err := ix.Close(); err != nil {
			return nil, fmt.Errorf("in CreateObsDss: %w", err)
		}
	}
	return NewObsDss(config, 0, nil)
}

// CleanObsDss cleans an "object-storage" DSS (data storage system)
// config provides the object store specification
func CleanObsDss(config ObsConfig) error {
	ods := &ODss{proxy: newObsProxy()}
	if err := ods.proxy.initialize(ods.proxy, config, 0, nil); err != nil {
		return fmt.Errorf("in CleanObsDss: %w", err)
	}
	defer ods.proxy.close()
	odoi := ods.proxy.(*oDssObjImpl)
	return odoi.is3.DeleteAll("")
}
