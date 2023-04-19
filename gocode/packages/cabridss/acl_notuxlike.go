//go:build !(unix || linux)

package cabridss

import (
	"os"
)

func getSysAcl(fi os.FileInfo) []ACLEntry {
	return getSysAclNotUx(fi)
}

func setSysAcl(path string, acl []ACLEntry) error {
	return setSysAclNotUx(path, acl)
}
