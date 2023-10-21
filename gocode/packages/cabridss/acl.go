package cabridss

import (
	"fmt"
	"os"
	"strings"
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
		{User: fmt.Sprintf("x-uid:%d", os.Getuid()), Rights: Rights{Read: perm&(1<<8) != 0, Write: perm&(1<<7) != 0, Execute: perm&(1<<6) != 0}},
	}
	return ael
}

func setSysAclNotUx(path string, acl []ACLEntry) error {
	if len(acl) == 0 {
		return nil
	}
	var mode os.FileMode
	var err error
	ur := GetUserRights(acl, fmt.Sprintf("x-uid:%d", os.Getuid()), Rights{})
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

// CheckUiACL convert a list of <user:rights> strings into actual ACL
func CheckUiACL(sacl []string) (acl []ACLEntry, err error) {
	for _, sac := range sacl {
		var (
			u, rights string
		)
		sacsubs := strings.Split(sac, ":")
		if strings.HasPrefix(sac, "x-uid") || strings.HasPrefix(sac, "x-gid") {
			if len(sacsubs) != 3 {
				return nil, fmt.Errorf("invalid ACL string %s, not <user:rights>", sac)
			}
			u, rights = sacsubs[0]+":"+sacsubs[1], sacsubs[2]
		} else {
			if len(sacsubs) != 2 {
				return nil, fmt.Errorf("invalid ACL string %s, not <user:rights>", sac)
			}
			u, rights = sacsubs[0], sacsubs[1]
		}
		ur := Rights{}
		for _, char := range rights {
			if char == 'r' {
				ur.Read = true
			} else if char == 'w' {
				ur.Write = true
			} else if char == 'x' {
				ur.Execute = true
			} else {
				return nil, fmt.Errorf("invalid character %c for access right (not in 'rwx')", char)
			}
		}
		if rights == "" {
			ur = Rights{Read: true, Write: true, Execute: true}
		}
		acl = append(acl, ACLEntry{User: u, Rights: ur})
	}
	return
}
