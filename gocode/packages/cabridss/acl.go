package cabridss

import (
	"fmt"
	"os"
)

type Rights struct {
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Execute bool `json:"execute"`
}

type ACLEntry struct {
	// on unix-like fsy DSS: x-uid:<uid> or x-gid:<gid> will be honored
	User   string `json:"user"` // on encrypted DSS, any alias for an IdentityConfig whose secret is owned by the user will be honored
	Rights Rights `json:"rights"`
}

func (ace ACLEntry) GetUser() string {
	return ace.User
}

func (ace ACLEntry) GetRights() Rights {
	return ace.Rights
}

func Users(aes []ACLEntry) (users []string) {
	for _, ae := range aes {
		users = append(users, ae.User)
	}
	return
}

func GetUserRights(aes []ACLEntry, user string, defaultRights Rights) Rights {
	for _, ae := range aes {
		if ae.User == user {
			return ae.Rights
		}
	}
	return defaultRights
}

func getSysAclNotUx(fi os.FileInfo) []ACLEntry {
	perm := fi.Mode().Perm()
	ael := []ACLEntry{
		{User: "", Rights: Rights{Read: perm&(1<<8) != 0, Write: perm&(1<<7) != 0, Execute: perm&(1<<6) != 0}},
	}
	return ael
}

func setSysAclNotUx(path string, acl []ACLEntry) error {
	if len(acl) == 0 {
		return nil
	}
	var mode os.FileMode
	var err error
	ur := GetUserRights(acl, "", Rights{})
	if ur.Read {
		mode |= 1 << 8
	}
	if ur.Write {
		mode |= 1 << 7
	}
	if ur.Execute {
		mode |= 1 << 6
	}
	if err = os.Chmod(path, mode); err != nil {
		return fmt.Errorf("in setSysAclNotUx: %v", err)
	}
	return nil
}
