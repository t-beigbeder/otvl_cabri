//go:build unix

package cabridss

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func getSysAcl(fi os.FileInfo) []ACLEntry {
	st := fi.Sys().(*syscall.Stat_t)
	perm := fi.Mode().Perm()
	ael := []ACLEntry{
		{User: fmt.Sprintf("x-uid:%d", st.Uid), Rights: Rights{Read: perm&(1<<8) != 0, Write: perm&(1<<7) != 0, Execute: perm&(1<<6) != 0}},
		{User: fmt.Sprintf("x-gid:%d", st.Gid), Rights: Rights{Read: perm&(1<<5) != 0, Write: perm&(1<<4) != 0, Execute: perm&(1<<3) != 0}},
		{User: "x-other", Rights: Rights{Read: perm&(1<<2) != 0, Write: perm&(1<<1) != 0, Execute: perm&(1) != 0}},
	}
	return ael
}

func setSysAcl(path string, acl []ACLEntry) error {
	if len(acl) == 0 {
		return nil
	}
	var mode os.FileMode
	var ur, gr, or Rights
	var uid, gid int
	var err error
	for _, ae := range acl {
		if strings.HasPrefix(ae.User, "x-uid:") {
			ur = ae.Rights
			if uid, err = strconv.Atoi(ae.User[len("x-uid:"):]); err != nil {
				return fmt.Errorf("in setSysAcl: %v", err)
			}
		} else if strings.HasPrefix(ae.User, "x-gid:") {
			gr = ae.Rights
			if gid, err = strconv.Atoi(ae.User[len("x-uid:"):]); err != nil {
				return fmt.Errorf("in setSysAcl: %v", err)
			}
		} else if ae.User == "x-other" {
			or = ae.Rights
		}
	}
	if ur.Read {
		mode |= 1 << 8
	}
	if ur.Write {
		mode |= 1 << 7
	}
	if ur.Execute {
		mode |= 1 << 6
	}
	if gr.Read {
		mode |= 1 << 5
	}
	if gr.Write {
		mode |= 1 << 4
	}
	if gr.Execute {
		mode |= 1 << 3
	}
	if or.Read {
		mode |= 1 << 2
	}
	if or.Write {
		mode |= 1 << 1
	}
	if or.Execute {
		mode |= 1
	}
	if err = os.Chmod(path, mode); err != nil {
		return fmt.Errorf("in setSysAcl: %v", err)
	}
	if os.Geteuid() == 0 {
		if err = os.Chown(path, uid, gid); err != nil {
			return fmt.Errorf("in setSysAcl: %v", err)
		}
	}
	return nil
}
