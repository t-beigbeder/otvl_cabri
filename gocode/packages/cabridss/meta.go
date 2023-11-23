package cabridss

type Meta struct {
	Path          string     `json:"path"`                    // full path for data content
	Mtime         int64      `json:"mtime"`                   // last modification POSIX time
	Size          int64      `json:"size"`                    // content size
	Ch            string     `json:"ch"`                      // truncated SHA256 checksum of the content
	IsNs          bool       `json:"isNs"`                    // is it a namespace, if true has children
	Children      []string   `json:"children"`                // namespace children, sorted by name
	IsSymLink     bool       `json:"isSymLink,omitempty"`     // is it a symbolic link, if true has SymLinkTarget
	SymLinkTarget string     `json:"symLinkTarget,omitempty"` // target of the symbolic link
	ACL           []ACLEntry `json:"acl"`                     // access control List, sorted by user
	Itime         int64      `json:"itime"`                   // index time
	ECh           string     `json:"ech"`                     // truncated SHA256 checksum of the encrypted content if encrypted else empty
	EMId          string     `json:"emid"`                    // encrypted meta-data unique identifier if encrypted else empty
}

type IMeta interface {
	GetPath() string                     // path in the DSS
	GetMtime() int64                     // last modification POSIX time
	GetSize() int64                      // content size
	GetCh() string                       // content truncated SHA256 checksum (panic if DSS does not enable)
	GetChUnsafe() string                 // content truncated SHA256 checksum or empty if DSS does not enable
	GetIsNs() bool                       // is it a namespace, if true has children
	GetChildren() []string               // namespace children, sorted by name
	GetIsSymLink() bool                  // is it a symbolic link, if true has SymLinkTarget
	GetSymLinkTarget() string            // target of the symbolic link
	GetAcl() []ACLEntry                  // access control List, sorted by user
	GetItime() int64                     // index time
	Equals(other IMeta, chacl bool) bool // checks equality (does not compare Ch if one end is unavailable) compare ACL if chacl true
}

type MetaMockCbs struct {
	MockMarshal   func(v interface{}) ([]byte, error)
	MockUnmarshal func(data []byte, v interface{}) error
}

func (m Meta) GetPath() string { return m.Path }

func (m Meta) GetMtime() int64 { return m.Mtime }

func (m Meta) GetSize() int64 { return m.Size }

func (m Meta) GetCh() string {
	if m.Ch == "" {
		panic("GetMeta didn't request getCh")
	}
	return m.Ch
}

func (m Meta) GetChUnsafe() string { return m.Ch }

func (m Meta) GetIsNs() bool { return m.IsNs }

func (m Meta) GetChildren() []string { return m.Children }

func (m Meta) GetIsSymLink() bool { return m.IsSymLink }

func (m Meta) GetSymLinkTarget() string { return m.SymLinkTarget }

func (m Meta) GetAcl() []ACLEntry { return m.ACL }

func (m Meta) GetItime() int64 { return m.Itime }

func CmpAcl(acl1, acl2 []ACLEntry) bool {
	if len(acl1) != len(acl2) {
		return false
	}
	for _, ace := range acl1 {
		found := false
		for _, oace := range acl2 {
			if ace.User != oace.User {
				continue
			}
			if ace.Rights.Read != oace.Rights.Read || ace.Rights.Write != oace.Rights.Write || ace.Rights.Execute != oace.Rights.Execute {
				return false
			}
			found = true
			break
		}
		if !found {
			return false
		}
	}
	return true
}

func (m Meta) Equals(om IMeta, chacl bool) bool {
	if om == nil {
		return false
	}
	if m.Size != om.GetSize() || m.Mtime != om.GetMtime() || (m.Ch != "" && om.GetChUnsafe() != "" && m.Ch != om.GetCh()) {
		return false
	}
	if chacl && !CmpAcl(m.ACL, om.GetAcl()) {
		return false
	}
	return true
}
