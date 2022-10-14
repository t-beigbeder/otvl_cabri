package mockfs

import (
	"github.com/spf13/afero"
	"os"
	"time"
)

type File struct {
	base afero.File
	cbs  *MockCbs
}

func (f File) Close() error {
	if f.cbs != nil && f.cbs.AfiClose != nil {
		return f.cbs.AfiClose(f.base)
	}
	return f.base.Close()
}

func (f File) Read(p []byte) (n int, err error) {
	if f.cbs != nil && f.cbs.AfiRead != nil {
		return f.cbs.AfiRead(f.base, p)
	}
	return f.base.Read(p)
}

func (f File) ReadAt(p []byte, off int64) (n int, err error) {
	return f.base.ReadAt(p, off)
}

func (f File) Seek(offset int64, whence int) (int64, error) {
	return f.base.Seek(offset, whence)
}

func (f File) Write(p []byte) (n int, err error) {
	if f.cbs != nil && f.cbs.AfiWrite != nil {
		return f.cbs.AfiWrite(f.base, p)
	}
	return f.base.Write(p)
}

func (f File) WriteAt(p []byte, off int64) (n int, err error) {
	return f.base.WriteAt(p, off)
}

func (f File) Name() string {
	return f.base.Name()
}

func (f File) Readdir(count int) ([]os.FileInfo, error) {
	if f.cbs != nil && f.cbs.AfiReaddir != nil {
		return f.cbs.AfiReaddir(f.base, count)
	}
	return f.base.Readdir(count)
}

func (f File) Readdirnames(n int) ([]string, error) {
	return f.base.Readdirnames(n)
}

func (f File) Stat() (os.FileInfo, error) {
	return f.base.Stat()
}

func (f File) Sync() error {
	return f.base.Sync()
}

func (f File) Truncate(size int64) error {
	return f.base.Truncate(size)
}

func (f File) WriteString(s string) (ret int, err error) {
	return f.base.WriteString(s)
}

type MockCbs struct {
	AfsMkdir     func(mfs afero.Fs, name string, perm os.FileMode) error
	AfsMkdirAll  func(mfs afero.Fs, path string, perm os.FileMode) error
	AfsCreate    func(mfs afero.Fs, name string) (afero.File, error)
	AfsStat      func(mfs afero.Fs, name string) (os.FileInfo, error)
	AfsOpen      func(mfs afero.Fs, name string) (afero.File, error)
	AfsRemoveAll func(mfs afero.Fs, name string) error
	AfsChtimes   func(mfs afero.Fs, name string, atime time.Time, mtime time.Time) error
	AfsRename    func(mfs afero.Fs, oldname, newname string) error
	AfsOpenFile  func(mfs afero.Fs, name string, flag int, perm os.FileMode) (afero.File, error)
	AfiWrite     func(mfi afero.File, p []byte) (n int, err error)
	AfiReaddir   func(mfi afero.File, count int) ([]os.FileInfo, error)
	AfiRead      func(mfi afero.File, p []byte) (n int, err error)
	AfiClose     func(mfi afero.File) error
}

type MockFs struct {
	base afero.Fs
	cbs  *MockCbs
}

func (m MockFs) GetBase() afero.Fs {
	return m.base
}

func (m MockFs) Create(name string) (afero.File, error) {
	if m.cbs != nil && m.cbs.AfsCreate != nil {
		base, err := m.cbs.AfsCreate(m.base, name)
		if err != nil {
			return nil, err
		}
		return &File{base: base, cbs: m.cbs}, nil
	}
	base, err := m.base.Create(name)
	if err != nil {
		return nil, err
	}
	return &File{base: base}, nil
}

func (m MockFs) Mkdir(name string, perm os.FileMode) error {
	if m.cbs != nil && m.cbs.AfsMkdir != nil {
		return m.cbs.AfsMkdir(m.base, name, perm)
	}
	return m.base.Mkdir(name, perm)
}

func (m MockFs) MkdirAll(path string, perm os.FileMode) error {
	if m.cbs != nil && m.cbs.AfsMkdirAll != nil {
		return m.cbs.AfsMkdirAll(m.base, path, perm)
	}
	return m.base.MkdirAll(path, perm)
}

func (m MockFs) Open(name string) (afero.File, error) {
	if m.cbs != nil && m.cbs.AfsOpen != nil {
		base, err := m.cbs.AfsOpen(m.base, name)
		if err != nil {
			return nil, err
		}
		return &File{base: base, cbs: m.cbs}, nil
	}
	base, err := m.base.Open(name)
	if err != nil {
		return nil, err
	}
	return &File{base: base}, nil
}

func (m MockFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if m.cbs != nil && m.cbs.AfsOpenFile != nil {
		base, err := m.cbs.AfsOpenFile(m.base, name, flag, perm)
		if err != nil {
			return nil, err
		}
		return &File{base: base, cbs: m.cbs}, nil
	}
	base, err := m.base.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &File{base: base}, nil
}

func (m MockFs) Remove(name string) error {
	return m.base.Remove(name)
}

func (m MockFs) RemoveAll(name string) error {
	if m.cbs != nil && m.cbs.AfsRemoveAll != nil {
		return m.cbs.AfsRemoveAll(m.base, name)
	}
	return m.base.RemoveAll(name)
}

func (m MockFs) Rename(oldname, newname string) error {
	if m.cbs != nil && m.cbs.AfsRename != nil {
		return m.cbs.AfsRename(m.base, oldname, newname)
	}
	return m.base.Rename(oldname, newname)
}

func (m MockFs) Stat(name string) (os.FileInfo, error) {
	if m.cbs != nil && m.cbs.AfsStat != nil {
		return m.cbs.AfsStat(m.base, name)
	}
	return m.base.Stat(name)
}

func (m MockFs) Name() string {
	return m.base.Name()
}

func (m MockFs) Chmod(name string, mode os.FileMode) error {
	return m.base.Chmod(name, mode)
}

func (m MockFs) Chown(name string, uid, gid int) error {
	return m.base.Chown(name, uid, gid)
}

func (m MockFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if m.cbs != nil && m.cbs.AfsChtimes != nil {
		return m.cbs.AfsChtimes(m.base, name, atime, mtime)
	}
	return m.base.Chtimes(name, atime, mtime)
}

func New(base afero.Fs, cbs *MockCbs) afero.Fs {
	fs := MockFs{base: base, cbs: cbs}
	return fs
}
