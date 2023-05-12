package cabridss

import (
	"crypto/sha256"
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"hash"
	"io"
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
	root string
	afs  afero.Fs
}

func (fsy FsyDss) mkUpdateNs(npath string, mtime int64, children []string, existing []string, acl []ACLEntry) error {
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

func (fsy FsyDss) Mkns(npath string, mtime int64, children []string, acl []ACLEntry) error {
	existing, err := fsy.Lsns(npath)
	if err != nil {
		return fmt.Errorf("in Mkns: %w", err)
	}
	if len(existing) != 0 {
		return fmt.Errorf("Mkns cannot be used on a non empty directory, use Updatens instead")
	}
	return fsy.mkUpdateNs(npath, mtime, children, nil, acl)
}

func (fsy FsyDss) Updatens(npath string, mtime int64, children []string, acl []ACLEntry) error {
	existing, err := fsy.Lsns(npath)
	if err != nil {
		return fmt.Errorf("in Updatens: %w", err)
	}
	return fsy.mkUpdateNs(npath, mtime, children, existing, acl)
}

func (fsy FsyDss) Lsns(npath string) ([]string, error) {
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
		} else if fi.Mode().IsRegular() {
			children = append(children, fi.Name())
		}
	}
	return children, nil
}

func (fsy FsyDss) GetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error) {
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

func (fsy FsyDss) IsDuplicate(ch string) (bool, error) {
	return false, nil
}

func (fsy FsyDss) GetContentReader(npath string) (io.ReadCloser, error) {
	cpath := ufpath.Join(fsy.root, npath)
	f, err := fsy.GetAfs().Open(cpath)
	if err != nil {
		return nil, fmt.Errorf("in GetContentReader: %w", err)
	}
	return f, nil
}

func (fsy FsyDss) ctlStat(fp string, isNS bool) (os.FileInfo, error) {
	fi, err := fsy.GetAfs().Stat(fp)
	if err != nil {
		return nil, fmt.Errorf("in ctlStat: %w", err)
	}
	if isNS && !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", fp)
	}
	if !isNS && fi.IsDir() {
		return nil, fmt.Errorf("%s is a directory", fp)
	}
	return fi, nil
}

func (fsy FsyDss) Remove(npath string) error {
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

func (fsy *FsyDss) GetMeta(npath string, getCh bool) (IMeta, error) {
	isNS, ipath, err := checkNCpath(npath)
	if err != nil {
		return nil, err
	}
	fp := ufpath.Join(fsy.root, ipath)
	fi, err := fsy.ctlStat(fp, isNS)
	if err != nil {
		return nil, fmt.Errorf("in GetMeta: %w", err)
	}
	if isNS {
		children, err := fsy.Lsns(ipath)
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
	} else {
		ch := ""
		if getCh {
			cr, err := fsy.GetContentReader(npath)
			if err != nil {
				return nil, fmt.Errorf("in GetMeta: %w", err)
			}
			defer cr.Close()
			hw := sha256.New()
			_, err = io.Copy(hw, cr)
			if err != nil {
				return nil, fmt.Errorf("in GetMeta: %w", err)
			}
			ch = internal.Sha256ToStr32(hw.Sum(nil))
		}
		return Meta{
			Path: npath, Mtime: fi.ModTime().Unix(), Size: fi.Size(), Ch: ch,
			IsNs:  false,
			ACL:   getSysAcl(fi),
			Itime: fi.ModTime().UnixNano(),
		}, nil
	}
}

func (fsy FsyDss) SetCurrentTime(time int64) {}

func (dss *FsyDss) GetAfs() afero.Fs {
	if dss.afs != nil {
		return dss.afs
	}
	return appFs
}

func (fsy *FsyDss) Close() error { return nil }

func (fsy *FsyDss) GetRoot() string { return fsy.root }

func NewFsyDss(config FsyConfig, root string) (Dss, error) {
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("in NewFsyDss: not a directory: %s", root)
	}
	return &FsyDss{root: root}, nil
}
