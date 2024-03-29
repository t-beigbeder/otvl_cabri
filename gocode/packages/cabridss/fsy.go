package cabridss

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrifsu"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"hash"
	"io"
	"io/fs"
	"os"
	"sort"
	"time"
)

type ContentHandle struct {
	cb      WriteCloserCb
	cf      afero.File
	h       hash.Hash
	written int64
}

func (ch *ContentHandle) Write(p []byte) (n int, err error) {
	n, err = ch.cf.Write(p)
	if n > 0 {
		ch.written += int64(n)
		ch.h.Write(p[0:n])
	}
	return
}

func (ch *ContentHandle) Close() (err error) {
	err = ch.cf.Close()
	if err != nil {
		os.Remove(ch.cf.Name())
	}
	if ch.cb != nil {
		ch.cb(err, ch.written, internal.Sha256ToStr32(ch.h.Sum(nil)))
	}
	return err
}

type FsyConfig struct {
	DssBaseConfig
}

type FsyDss struct {
	root    string
	afs     afero.Fs
	reducer plumber.Reducer
	su      bool
}

func (fsy *FsyDss) doMkUpdateNs(npath string, mtime int64, children []string, existing []string, acl []ACLEntry) error {
	err := checkMknsArgs(npath, children, acl)
	if err != nil {
		return err
	}

	removeIfError := func(path string) {
		if err != nil {
			_ = fsy.GetAfs().Remove(path)
		}
	}

	created := make([]string, 0, len(children))
	for _, child := range children {
		found := false
		for _, ec := range existing {
			if ec == child {
				found = true
				break
			}
		}
		if !found {
			created = append(created, child)
		}
	}
	removed := make([]string, 0, len(existing))
	for _, ec := range existing {
		found := false
		for _, child := range children {
			if child == ec {
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, ec)
		}
	}

	for _, child := range created {
		cpath := ufpath.Join(fsy.root, npath, child)
		if child[len(child)-1] == '/' {
			if err = fsy.GetAfs().Mkdir(cpath, 0o777); err != nil {
				return fmt.Errorf("in Mkns/Updatens: %w", err)
			}
		} else {
			var f afero.File
			if f, err = fsy.GetAfs().Create(cpath); err != nil {
				return fmt.Errorf("in Mkns/Updatens: %w", err)
			}
			f.Close()
		}
		defer removeIfError(cpath)
	}

	for _, child := range removed {
		cpath := ufpath.Join(fsy.root, npath, child)
		if err = fsy.GetAfs().RemoveAll(cpath); err != nil {
			return fmt.Errorf("in Mkns/Updatens: %w", err)
		}
	}

	if err = fsy.GetAfs().Chtimes(ufpath.Join(fsy.root, npath), time.Now(), time.Unix(mtime, 0)); err != nil {
		return fmt.Errorf("in Mkns/Updatens: %w", err)
	}
	if err = setSysAcl(ufpath.Join(fsy.root, npath), acl); err != nil {
		return fmt.Errorf("in Mkns/Updatens: %w", err)
	}
	return nil
}

func (fsy *FsyDss) mkUpdateNs(npath string, mtime int64, children []string, existing []string, acl []ACLEntry) (err error) {
	if fsy.reducer == nil {
		return fsy.doMkUpdateNs(npath, mtime, children, existing, acl)
	}
	return fsy.reducer.Launch(
		fmt.Sprintf("mkUpdateNs %s", npath),
		func() error {
			return fsy.doMkUpdateNs(npath, mtime, children, existing, acl)
		})
}

func (fsy *FsyDss) Mkns(npath string, mtime int64, children []string, acl []ACLEntry) error {
	existing, err := fsy.Lsns(npath)
	if err != nil {
		return fmt.Errorf("in Mkns: %w", err)
	}
	if len(existing) != 0 {
		return fmt.Errorf("Mkns cannot be used on a non empty directory, use Updatens instead")
	}
	return fsy.mkUpdateNs(npath, mtime, children, nil, acl)
}

func (fsy *FsyDss) Updatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
	existing, err := fsy.Lsns(npath)
	if err != nil {
		return fmt.Errorf("in Updatens: %w", err)
	}
	return fsy.mkUpdateNs(npath, mtime, children, existing, acl)
}

func (fsy *FsyDss) doLsns(npath string) ([]string, error) {
	if err := checkNpath(npath); err != nil {
		return nil, err
	}
	cpath := ufpath.Join(fsy.root, npath)
	f, err := fsy.GetAfs().Open(cpath)
	if err != nil {
		return nil, fmt.Errorf("in Lsns: %w", err)
	}
	defer f.Close()
	tfi, err := f.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("in Lsns: %w", err)
	}
	var children []string
	for _, fi := range tfi {
		if fi.IsDir() {
			children = append(children, fi.Name()+"/")
		} else if fi.Mode().IsRegular() || fi.Mode().Type()&fs.ModeSymlink != 0 {
			children = append(children, fi.Name())
		}
	}
	return children, nil
}

func (fsy *FsyDss) Lsns(npath string) (children []string, err error) {
	if fsy.reducer == nil {
		children, err = fsy.doLsns(npath)
		return
	}
	if err = fsy.reducer.Launch(
		fmt.Sprintf("Lsns %s", npath),
		func() error {
			var iErr error
			if children, iErr = fsy.doLsns(npath); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (fsy *FsyDss) doGetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
	lcb := func(err error, size int64, ch string) {
		if err == nil {
			cpath := ufpath.Join(fsy.root, npath)
			err = fsy.GetAfs().Chtimes(cpath, time.Now(), time.Unix(mtime, 0))
			if err != nil {
				os.Remove(cpath)
			} else if err = setSysAcl(cpath, acl); err != nil {
				os.Remove(cpath)
			}
		}
		if cb != nil {
			cb(err, size, ch)
		}
	}
	err := checkMkcontentArgs(npath, acl)
	if err != nil {
		return nil, err
	}
	cpath := ufpath.Join(fsy.root, npath)
	var cf afero.File
	if cf, err = fsy.GetAfs().Create(cpath); err != nil {
		return nil, fmt.Errorf("in GetContentWriter: %w", err)
	}
	return &ContentHandle{cb: lcb, cf: cf, h: sha256.New()}, nil
}

func (fsy *FsyDss) GetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (wc io.WriteCloser, err error) {
	if fsy.reducer == nil {
		wc, err = fsy.doGetContentWriter(npath, mtime, acl, cb)
		return
	}
	if err = fsy.reducer.Launch(
		fmt.Sprintf("GetContentWriter %s", npath),
		func() error {
			var iErr error
			if wc, iErr = fsy.doGetContentWriter(npath, mtime, acl, cb); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (fsy *FsyDss) IsDuplicate(ch string) (bool, error) {
	return false, nil
}

func (fsy *FsyDss) doGetContentReader(npath string) (io.ReadCloser, error) {
	cpath := ufpath.Join(fsy.root, npath)
	f, err := fsy.GetAfs().Open(cpath)
	if err != nil {
		return nil, fmt.Errorf("in GetContentReader: %w", err)
	}
	return f, nil
}

func (fsy *FsyDss) GetContentReader(npath string) (rc io.ReadCloser, err error) {
	if fsy.reducer == nil {
		rc, err = fsy.doGetContentReader(npath)
		return
	}
	if err = fsy.reducer.Launch(
		fmt.Sprintf("GetContentReader %s", npath),
		func() error {
			var iErr error
			if rc, iErr = fsy.doGetContentReader(npath); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (fsy *FsyDss) doSymlink(npath string, tpath string, mtime int64, acl []ACLEntry) error {
	if err := checkNpath(npath); err != nil {
		return err
	}
	cpath := ufpath.Join(fsy.root, npath)
	if fi, err := os.Lstat(cpath); fi != nil && err == nil {
		ppath := ufpath.Join(fsy.root, ufpath.Dir(npath))
		uio, rw, err := cabrifsu.HasFileWriteAccess(ppath)
		if err != nil {
			return fmt.Errorf("in Symlink: %w", err)
		}
		if uio && !rw && fsy.su {
			if err = cabrifsu.EnableWrite(fsy.GetAfs(), ppath, false); err != nil {
				return fmt.Errorf("in Symlink: %w", err)
			}
		}
		if err := os.Remove(cpath); err != nil {
			return fmt.Errorf("in Symlink: %w", err)
		}
	}
	if err := os.Symlink(tpath, cpath); err != nil {
		return fmt.Errorf("in Symlink: %w", err)
	}
	if err := cabrifsu.Lutimes(cpath, mtime); err != nil {
		os.Remove(cpath)
	}
	return nil
}

func (fsy *FsyDss) Symlink(npath string, tpath string, mtime int64, acl []ACLEntry) error {
	if fsy.reducer == nil {
		return fsy.doSymlink(npath, tpath, mtime, acl)
	}
	return fsy.reducer.Launch(
		fmt.Sprintf("Symlink %s", npath),
		func() error {
			return fsy.doSymlink(npath, tpath, mtime, acl)
		})
}

func (fsy *FsyDss) ctlStat(fp string, isNS bool) (os.FileInfo, error) {
	checkLfi := func(fp string, err error) (os.FileInfo, bool, error) {
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, false, err
		}
		lfi, lerr := os.Lstat(fp)
		if lerr != nil {
			return nil, false, err
		}
		if lfi.Mode().Type()&fs.ModeSymlink != 0 {
			return lfi, true, nil
		}
		return nil, false, err
	}
	fi, err := fsy.GetAfs().Stat(fp)
	lfi, isSymLink, err := checkLfi(fp, err)
	if err != nil {
		return nil, fmt.Errorf("in ctlStat: %w", err)
	}
	if isSymLink {
		fi = lfi
	}
	if isNS && !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", fp)
	}
	if !isNS && fi.IsDir() {
		return nil, fmt.Errorf("%s is a directory", fp)
	}
	return fi, nil
}

func (fsy *FsyDss) doRemove(npath string) error {
	isNS, npath, err := checkNCpath(npath)
	if err != nil {
		return err
	}
	if npath == "" {
		return fmt.Errorf("cannot remove root")
	}
	fp := ufpath.Join(fsy.root, npath)
	_, err = fsy.ctlStat(fp, isNS)
	if err != nil {
		return fmt.Errorf("in Remove: %w", err)
	}
	if isNS {
		return fsy.GetAfs().RemoveAll(fp)
	} else {
		return fsy.GetAfs().Remove(fp)
	}
}

func (fsy *FsyDss) Remove(npath string) (err error) {
	if fsy.reducer == nil {
		return fsy.doRemove(npath)
	}
	return fsy.reducer.Launch(
		fmt.Sprintf("Remove %s", npath),
		func() error {
			return fsy.doRemove(npath)
		})
}

func (fsy *FsyDss) doGetMeta(npath string, getCh bool) (IMeta, error) {
	isNS, ipath, err := checkNCpath(npath)
	if err != nil {
		return nil, err
	}
	fp := ufpath.Join(fsy.root, ipath)
	fi, err := fsy.ctlStat(fp, isNS)
	if err != nil {
		return nil, fmt.Errorf("in GetMeta: %w", err)
	}
	isSymLink := fi.Mode().Type()&fs.ModeSymlink != 0
	if isNS {
		children, err := fsy.doLsns(ipath)
		sort.Strings(children)
		if err != nil {
			return nil, fmt.Errorf("in GetMeta: %w", err)
		}
		nsc, ch, _ := internal.Ns2Content(children, "")
		return Meta{
			Path: npath, Mtime: fi.ModTime().Unix(), Size: int64(len(nsc)), Ch: ch,
			IsNs: true, Children: children,
			ACL:   getSysAcl(fi),
			Itime: fi.ModTime().UnixNano(),
		}, nil
	} else if isSymLink {
		rl, err := os.Readlink(ufpath.Join(fsy.root, npath))
		if err != nil {
			return nil, fmt.Errorf("in GetMeta: %w", err)
		}
		cs := sha256.Sum256([]byte(rl))
		ch := internal.Sha256ToStr32(cs[:])
		return Meta{
			Path: npath, Mtime: fi.ModTime().Unix(), Size: fi.Size(), Ch: ch,
			IsNs:          false,
			IsSymLink:     true,
			SymLinkTarget: rl,
			ACL:           getSysAcl(fi),
			Itime:         fi.ModTime().UnixNano(),
		}, nil
	} else {
		ch := ""
		if getCh {
			cr, err := fsy.doGetContentReader(npath)
			if err != nil {
				return nil, fmt.Errorf("in GetMeta: %w", err)
			}
			defer cr.Close()
			ch, err = internal.ShaFrom(cr)
			if err != nil {
				return nil, fmt.Errorf("in GetMeta: %w", err)
			}
		}
		return Meta{
			Path: npath, Mtime: fi.ModTime().Unix(), Size: fi.Size(), Ch: ch,
			IsNs:  false,
			ACL:   getSysAcl(fi),
			Itime: fi.ModTime().UnixNano(),
		}, nil
	}
}

func (fsy *FsyDss) GetMeta(npath string, getCh bool) (meta IMeta, err error) {
	if fsy.reducer == nil {
		meta, err = fsy.doGetMeta(npath, getCh)
		return
	}
	if err = fsy.reducer.Launch(
		fmt.Sprintf("GetMeta %s", npath),
		func() error {
			var iErr error
			if meta, iErr = fsy.doGetMeta(npath, getCh); iErr != nil {
				return iErr
			}
			return nil
		}); err != nil {
		return
	}
	return
}

func (fsy *FsyDss) SetCurrentTime(time int64) {}

func (fsy *FsyDss) GetAfs() afero.Fs {
	if fsy.afs != nil {
		return fsy.afs
	}
	return appFs
}

func (fsy *FsyDss) Close() error {
	if fsy.reducer != nil {
		return fsy.reducer.Close()
	}
	return nil
}

func (fsy *FsyDss) SetSu() { fsy.su = true }

func (fsy *FsyDss) SuEnableWrite(npath string) error {
	if !fsy.su {
		return fmt.Errorf("in SuEnableWrite: not in su mode")
	}
	_, ipath, err := checkNCpath(npath)
	if err != nil {
		return err
	}
	return cabrifsu.EnableWrite(fsy.GetAfs(), ufpath.Join(fsy.root, ipath), false)
}

func (fsy *FsyDss) GetRoot() string { return fsy.root }

func NewFsyDss(config FsyConfig, root string) (Dss, error) {
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("in NewFsyDss: not a directory: %s", root)
	}
	var red plumber.Reducer = nil
	if config.ReducerLimit != 0 {
		red = plumber.NewReducer(config.ReducerLimit, 0)
	}
	return &FsyDss{root: root, reducer: red}, nil
}
