//go:build !unix

package cabridss

import (
	"fmt"
	"os"
	"runtime"
)

func getSysAcl(fi os.FileInfo) []ACLEntry {
	if runtime.GOOS != "windows" {
		panic(fmt.Sprintf("cabridss.getSysAcl not uxlike was only tested on windows"))
	}
	return getSysAclNotUx(fi)
}

func setSysAcl(path string, acl []ACLEntry) error {
	if runtime.GOOS != "windows" {
		panic(fmt.Sprintf("cabridss.setSysAcl not uxlike was only tested on windows"))
	}
	return setSysAclNotUx(path, acl)
}
