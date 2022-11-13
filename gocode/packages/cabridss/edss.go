package cabridss

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"time"
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

func (edi *eDssImpl) initialize(config interface{}, lsttime int64, aclusers []string) error {
	edc := config.(eDssClientConfig)
	if err := edi.webDssImpl.initialize(edc.webDssClientConfig, lsttime, aclusers); err != nil {
		return fmt.Errorf("in eDssImpl.initialize: %w", err)
	}
	if !edi.repoEncrypted {
		return fmt.Errorf("in eDssImpl.initialize: the repository is not encrypted")
	}
	edi.me = edi
	return nil
}

func (edi *eDssImpl) isDuplicate(ch string) (bool, error) {
	return false, nil // encrypted content is never the same
}

func (edi *eDssImpl) isEncrypted() bool { return true }

func (edi *eDssImpl) defaultAcl(acl []ACLEntry) []ACLEntry {
	if acl != nil {
		return acl
	}
	for _, id := range edi.apc.GetConfig().(webDssClientConfig).identities {
		if id.Alias == "" {
			return []ACLEntry{{User: id.PKey, Rights: Rights{Write: true}}}
		}
	}
	return nil
}

func (edi *eDssImpl) secrets(acl []ACLEntry) (res []string) {
	for _, user := range Users(acl) {
		for _, id := range edi.apc.GetConfig().(webDssClientConfig).identities {
			if id.PKey == user {
				res = append(res, id.Secret)
			}
		}
	}
	return
}

func (edi *eDssImpl) doGetMetaTimesFor(npath string) ([]int64, error) {
	return nil, nil // encrypted meta is only retrieved from local index
}

func (edi *eDssImpl) doGetMetaAt(npath string, time int64) (Meta, error) {
	bs, err, ok := edi.index.loadMeta(npath, time)
	if err != nil || !ok {
		return Meta{}, err
	}
	var meta Meta
	if edi.metamockcbs != nil && edi.metamockcbs.MockUnmarshal != nil {
		err = edi.metamockcbs.MockUnmarshal(bs, &meta)
	} else {
		err = json.Unmarshal(bs, &meta)
	}
	if err != nil {
		return Meta{}, err
	}
	return meta, nil
}

func (edi *eDssImpl) xStoreMeta(anpath string, atime int64, abs []byte, acl []ACLEntry) error {
	if err := cXStoreMeta(edi.apc, anpath, atime, abs, acl); err != nil {
		return fmt.Errorf("in xStoreMeta: %v", err)
	}
	sbs, err := DecryptMsg(abs, edi.secrets(acl)...)
	if err != nil {
		return fmt.Errorf("in xStoreMeta: %v", err)
	}
	var meta Meta
	if err := json.Unmarshal([]byte(sbs), &meta); err != nil {
		return fmt.Errorf("in xStoreMeta: %v", err)
	}
	npath := meta.Path
	if meta.IsNs {
		npath = RemoveSlashIf(meta.Path)
	}
	return edi.index.storeMeta(npath, meta.Itime, []byte(sbs))
}

func (edi *eDssImpl) getEncodedContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	cf, err := os.CreateTemp("", "ccw")
	if err != nil {
		return nil, fmt.Errorf("in getEncodedContentWriter: %w", err)
	}
	lcb := func(err error, size int64, ch string) {
		panic("FIXME: to be migrated")
		//if err == nil {
		//	err = edi.onCloseContent(npath, mtime, cf, size, ch, acl, func(npath string, time int64, bs []byte) error {
		//		if err = edi.me.xStoreMeta(npath, time, bs, acl); err != nil {
		//			return fmt.Errorf("in getEncodedContentWriter: %w", err)
		//		}
		//		return nil
		//	})
		//}
		//if cb != nil {
		//	cb(err, size, ch)
		//}
	}
	return &ContentHandle{cb: lcb, cf: cf, h: sha256.New()}, nil
}

func (edi *eDssImpl) apiGetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	// cf oDssBaseImpl.metaSetter
	//meta := Meta{Path: npath, Mtime: mtime, Size: size, Ch: css, ACL: acl}
	//time := time.Now().Unix()
	//_, anpath, atime, abs, err := odbi.me.apiMetaArgs(npath, time, bs, meta.ACL)
	anpath := uuid.New().String()
	var (
		cSize int64
		cCh   string
	)
	iTime := time.Now().Unix()
	if edi.mockct != 0 {
		iTime = edi.mockct
		edi.mockct += 1
	}
	cMeta := Meta{
		Path: npath, Mtime: mtime, Size: cSize, Ch: cCh, ACL: acl,
		Itime: iTime, Empath: uuid.New().String(), Ecpath: anpath,
	}
	_ = cMeta
	ewc, err := edi.getEncodedContentWriter(anpath, time.Now().Unix(), nil, func(err error, size int64, ch string) {
		// server, then user stuff
		cb(err, size, ch)
	})
	if err != nil {
		return nil, fmt.Errorf("in apiGetContentWriter: %w", err)
	}
	wc, err := Encrypt(ewc, Users(acl)...)
	if err != nil {
		return nil, fmt.Errorf("in apiGetContentWriter: %w", err)
	}
	return NewWriteCloserWithCb(wc, func(err error, size int64, ch string, wcwc *WriteCloserWithCb) error {
		if err == nil {
			cSize = size
			cCh = ch
			if err = ewc.Close(); err == nil {
				return nil
			}
		} else {
			ewc.Close()
		}
		return fmt.Errorf("in apiGetContentWriter close CB: %v", err)
	}), nil
}

func (edi *eDssImpl) openSession(aclusers []string) error {
	if err := cOpenSession(edi.apc, aclusers); err != nil {
		return fmt.Errorf("in openSession: %w", err)
	}
	return nil
}

func newEDssProxy(config EDssConfig, lsttime int64, aclusers []string) (oDssProxy, HDss, error) {
	wdp, dss, err := newWebDssProxy(config.WebDssConfig, lsttime, aclusers)
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
func NewEDss(config EDssConfig, lsttime int64, aclusers []string) (HDss, error) {
	proxy, libDss, err := newEDssProxy(config, lsttime, aclusers)
	if err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	wdcc := eDssClientConfig{webDssClientConfig{WebDssConfig: config.WebDssConfig, libDss: libDss}}
	if err := proxy.initialize(wdcc, lsttime, aclusers); err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	edi := proxy.(*eDssImpl)
	if err := edi.openSession(aclusers); err != nil {
		return nil, fmt.Errorf("in NewWebDss: %w", err)
	}
	return &ODss{proxy: proxy}, nil
}
