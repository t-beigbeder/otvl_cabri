//go:build !unix

package cabrifsu

import (
	"fmt"
	"runtime"
)

func HasFileWriteAccess(pathOrFi any) (bool, bool, error) {
	if runtime.GOOS != "windows" {
		panic(fmt.Sprintf("cabrifsu.HasFileWriteAccess not uxlike was only tested on windows"))
	}
	return hasFileWriteAccessNotUx(pathOrFi)
}

func Lutimes(path string, mtime int64) error {
	return nil
}
