//go:build unix || linux || darwin

package cabrifsu

import (
	"os"
	"syscall"
)

func HasFileWriteAccess(pathOrFi any) (bool, bool, error) {
	var (
		st   *syscall.Stat_t
		path string
		fi   os.FileInfo
		ok   bool
		err  error
	)
	fi, ok = pathOrFi.(os.FileInfo)
	if !ok {
		path = pathOrFi.(string)
		fi, err = os.Stat(path)
		if err != nil {
			return false, false, err
		}
	}
	st = fi.Sys().(*syscall.Stat_t)
	if int(st.Uid) == os.Getuid() {
		return true, st.Mode&(1<<7) != 0, nil
	}
	gids, err := os.Getgroups()
	if err != nil {
		return false, false, err
	}
	for _, gid := range gids {
		if gid == int(st.Gid) {
			return false, st.Mode&(1<<4) != 0, nil
		}
	}
	return false, false, nil
}
