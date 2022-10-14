package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
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

func (edi *eDssImpl) apiMetaArgs(npath string, time int64, bs []byte, acl []ACLEntry) (anpath string, atime int64, abs []byte, err error) {
	anpath = uuid.New().String()
	atime = 0
	if abs, err = EncryptMsg(string(bs), Users(acl)...); err != nil {
		err = fmt.Errorf("in apiMetaArgs: %w", err)
		return
	}
	return
}

func (edi *eDssImpl) apiGetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	snpath := uuid.New().String()
	ewc, err := edi.me.doGetContentWriter(snpath, mtime, acl, func(err error, size int64, sha256trunc []byte) {
		cb(err, size, sha256trunc)
	})
	if err != nil {
		return nil, fmt.Errorf("in apiGetContentWriter: %w", err)
	}
	wc, err := Encrypt(ewc, Users(acl)...)
	if err != nil {
		return nil, fmt.Errorf("in apiGetContentWriter: %w", err)
	}
	return NewWriteCloserWithCb(wc, func(err error) error {
		if err == nil {
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
