package cabridss

import (
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"sort"
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

func (odoi *oDssObjImpl) initialize(config interface{}, lsttime int64, aclusers []string) error {
	odoi.me = odoi
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

func (odoi *oDssObjImpl) xStoreMeta(npath string, time int64, bs []byte) error {
	return odoi.index.storeMeta(npath, time, bs)
}

func (odoi *oDssObjImpl) removeMeta(npath string, time int64) error {
	return odoi.is3.Delete(fmt.Sprintf("meta-%s.%s", internal.NameToHashStr32(npath), internal.Int64ToStr16(time)))
}

func (odoi *oDssObjImpl) xRemoveMeta(npath string, time int64) error {
	return odoi.index.removeMeta(npath, time)
}

func (odoi *oDssObjImpl) onCloseContent(npath string, mtime int64, cf afero.File, size int64, sha256 []byte, acl []ACLEntry, smCb storeMetaCallback) error {
	defer os.Remove(cf.Name())
	cName := fmt.Sprintf("content-%s", internal.Sha256ToStr32(sha256))
	lr, _ := odoi.is3.List(cName)
	if len(lr) == 0 {
		r, err := os.Open(cf.Name())
		if err != nil {
			return fmt.Errorf("in onCloseContent: %w", err)
		}
		defer r.Close()
		if err = odoi.is3.Upload(cName, r); err != nil {
			return fmt.Errorf("in onCloseContent: %w", err)
		}
	}
	sort.Slice(acl, func(i, j int) bool {
		return acl[i].User < acl[j].User
	})
	meta := Meta{Path: npath, Mtime: mtime, Size: size, Ch: internal.Sha256ToStr32(sha256), ACL: acl}
	if err := odoi.metaSetter(meta, smCb); err != nil {
		return fmt.Errorf("in onCloseContent: %w", err)
	}
	return nil
}

func (odoi *oDssObjImpl) doGetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	cf, err := os.CreateTemp("", "cw")
	if err != nil {
		return nil, fmt.Errorf("in GetContentWriter: %w", err)
	}
	lcb := func(err error, size int64, sha256trunc []byte) {
		if err == nil {
			err = odoi.onCloseContent(npath, mtime, cf, size, sha256trunc, acl, nil)
		}
		if cb != nil {
			cb(err, size, sha256trunc)
		}
	}
	return &ContentHandle{cb: lcb, cf: cf, h: sha256.New()}, nil
}

func (odoi *oDssObjImpl) doGetContentReader(npath string, meta Meta) (io.ReadCloser, error) {
	return odoi.is3.Download(fmt.Sprintf("content-%s", meta.Ch))
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

func (odoi *oDssObjImpl) dumpIndex() string { return odoi.index.Dump() }

func (odoi *oDssObjImpl) setAfs(tfs afero.Fs) { panic("inconsistent") }

func (odoi *oDssObjImpl) getAfs() afero.Fs { panic("inconsistent") }

func (odoi *oDssObjImpl) scanMetaObjs(sti StorageInfo, errs *ErrorCollector) {
	pathErr := func(path string, err error) {
		sti.Path2Error[path] = err
		errs.Collect(err)
	}
	mns, err := odoi.is3.List("meta-")
	if err != nil {
		pathErr("meta-", err)
		return
	}
	for _, mn := range mns {
		suffix := ufpath.Ext(mn)
		if len(suffix) == 0 {
			pathErr(mn, fmt.Errorf("no suffix"))
			continue
		}
		t, err := internal.Str16ToInt64(suffix[1:])
		if err != nil {
			pathErr(mn, err)
		}
		_ = t
		bs, err := odoi.is3.Get(mn)
		if err != nil {
			pathErr(mn, err)
		}
		sti.Path2Meta[mn] = bs
	}
}

func (odoi *oDssObjImpl) scanContentObjs(sti StorageInfo, errs *ErrorCollector) {
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
}

func (odoi *oDssObjImpl) scanPhysicalStorage(sti StorageInfo, errs *ErrorCollector) {
	odoi.scanMetaObjs(sti, errs)
	odoi.scanContentObjs(sti, errs)
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
func NewObsDss(config ObsConfig, lsttime int64, aclusers []string) (HDss, error) {
	proxy := newObsProxy()
	if err := proxy.initialize(config, lsttime, aclusers); err != nil {
		return nil, fmt.Errorf("in NewObsDss: %w", err)
	}
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
	if err := ods.proxy.initialize(config, 0, nil); err != nil {
		return fmt.Errorf("in CleanObsDss: %w", err)
	}
	defer ods.proxy.close()
	odoi := ods.proxy.(*oDssObjImpl)
	return odoi.is3.DeleteAll("")
}
