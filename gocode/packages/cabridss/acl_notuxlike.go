//go:build !(unix || linux)

package cabridss

import (
	"os"
)

func getSysAcl(fi os.FileInfo) []ACLEntry         { return nil }
func setSysAcl(path string, acl []ACLEntry) error { return nil }
