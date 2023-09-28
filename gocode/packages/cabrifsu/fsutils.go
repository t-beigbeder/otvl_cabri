package cabrifsu

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
)

func hasFileWriteAccessNotUx(pathOrFi any) (bool, bool, error) {
	var (
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
	return true, fi.Mode()&(1<<7) != 0, nil
}

func doEnableWrite(afs afero.Fs, path string, fi os.FileInfo, recursive bool) error {
	uio, rw, err := HasFileWriteAccess(fi)
	if err != nil {
		return fmt.Errorf("in doEnableWrite: %v", err)
	}
	if !rw && !uio {
		return fmt.Errorf("%s is read-only", path)
	}
	if !rw {
		var mode os.FileMode
		if uio {
			mode = fi.Mode() | 1<<8 | 1<<7
			if fi.IsDir() {
				mode |= 1 << 6
			}
		} else {
			mode = fi.Mode() | 1<<5 | 1<<4
			if fi.IsDir() {
				mode |= 1 << 3
			}
		}
		if err := afs.Chmod(path, mode); err != nil {
			return fmt.Errorf("in doEnableWrite: %v", err)
		}
	}
	if !fi.IsDir() || !recursive {
		return nil
	}
	f, err := afs.Open(path)
	if err != nil {
		return fmt.Errorf("in doEnableWrite: %w", err)
	}
	defer f.Close()
	cfis, err := f.Readdir(0)
	if err != nil {
		return fmt.Errorf("in doEnableWrite: %w", err)
	}
	for _, cfi := range cfis {
		err = doEnableWrite(afs, ufpath.Join(path, cfi.Name()), cfi, true)
		if err != nil {
			return fmt.Errorf("in doEnableWrite: %w", err)
		}
	}
	return nil
}

func EnableWrite(afs afero.Fs, path string, recursive bool) error {
	fi, err := afs.Stat(path)
	if err != nil {
		return fmt.Errorf("in EnableWrite: %v", err)
	}
	return doEnableWrite(afs, path, fi, recursive)
}
func doDisableWrite(afs afero.Fs, path string, fi os.FileInfo, recursive bool) error {
	if fi.IsDir() && recursive {
		f, err := afs.Open(path)
		if err != nil {
			return fmt.Errorf("in doDisableWrite: %w", err)
		}
		cfis, err := f.Readdir(0)
		if err != nil {
			f.Close()
			return fmt.Errorf("in doDisableWrite: %w", err)
		}
		f.Close()
		for _, cfi := range cfis {
			err = doDisableWrite(afs, ufpath.Join(path, cfi.Name()), cfi, true)
			if err != nil {
				return fmt.Errorf("in doDisableWrite: %w", err)
			}
		}
	}
	mode := fi.Mode() & 0555
	if err := afs.Chmod(path, mode); err != nil {
		return fmt.Errorf("in doDisableWrite: %v", err)
	}
	return nil
}

func DisableWrite(afs afero.Fs, path string, recursive bool) error {
	fi, err := afs.Stat(path)
	if err != nil {
		return fmt.Errorf("in EnableWrite %s: %v", path, err)
	}
	return doDisableWrite(afs, path, fi, recursive)
}

func GetFileMode(afs afero.Fs, path string) (mode os.FileMode, ro bool, err error) {
	fi, err := afs.Stat(path)
	if err != nil {
		return 0, false, fmt.Errorf("in GetFileMode: %v", err)
	}
	return fi.Mode(), fi.Mode()&0200 == 0, nil
}
