//go:build !(unix || linux)

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
