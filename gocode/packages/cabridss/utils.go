package cabridss

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"hash"
	"io"
	"os"
)

func doMkallNs(dss Dss, npath string, mtime int64) error {
	m, _ := dss.GetMeta(npath+"/", false)
	if m != nil {
		return nil
	}
	pnpath, me := ufpath.Split(npath)
	pnpath = RemoveSlashIf(pnpath)
	if npath != "" {
		if err := doMkallNs(dss, pnpath, mtime); err != nil {
			return err
		}
	} else {
		return dss.Mkns("", mtime, nil, nil)
	}

	others, err := dss.Lsns(pnpath)
	if err != nil {
		return err
	}
	if !(internal.NpType(me + "/")).ExistIn(others) {
		others = append(others, me+"/")
	}
	if err = dss.Updatens(pnpath, mtime, others, nil); err != nil {
		return err
	}
	return dss.Mkns(npath, mtime, nil, nil)
}

func MkallNs(dss Dss, npath string, mtime int64) error {
	if err := checkNpath(npath); err != nil {
		return err
	}
	if npath == "" {
		return nil
	}
	return doMkallNs(dss, npath, mtime)
}

func MkallContent(dss Dss, cpath string, mtime int64) error {
	ns, c := ufpath.Split(cpath)
	if ns != "" {
		ns = RemoveSlashIf(ns)
		if err := MkallNs(dss, ns, mtime); err != nil {
			return err
		}
	}
	cs, err := dss.Lsns(ns)
	if err != nil {
		return err
	}
	if !(internal.NpType(c)).ExistIn(cs) {
		cs = append(cs, c)
	}
	return dss.Updatens(ns, mtime, cs, nil)
}

func Parent(npath string) (parent string) {
	if npath == "" {
		panic("no parent")
	}
	if npath[len(npath)-1] == '/' {
		npath = npath[:len(npath)-1]
	}
	parent = ufpath.Dir(npath)
	if parent == "." {
		parent = ""
	}
	return
}

func AppendSlashIf(path string) string {
	if path == "" {
		return path
	}
	return path + "/"
}

func RemoveSlashIf(path string) string {
	if path == "" {
		return path
	}
	if path[len(path)-1] != '/' {
		panic("not a nspath " + path)
	}
	return path[0 : len(path)-1]
}

func RemoveSlashIfNsIf(path string, isNs bool) string {
	if isNs {
		return RemoveSlashIf(path)
	}
	return path
}

func isNpathIn(npath string, isDir bool, parentCs []string) (res bool) {
	parent := ufpath.Dir(npath)
	me := ufpath.Base(npath)
	if isDir {
		me += "/"
	}
	if parent == "." {
		parent = ""
	}
	for _, pc := range parentCs {
		if pc == me {
			return true
		}
	}
	return
}

type ErrorCollector []error

func (c *ErrorCollector) Collect(e error) { *c = append(*c, e) }

func (c *ErrorCollector) Error() (err string) {
	err = "Collected errors:\n"
	for i, e := range *c {
		err += fmt.Sprintf("\tError %d: %s\n", i, e.Error())
	}

	return err
}

func (c *ErrorCollector) Any() bool { return len(*c) != 0 }

type WriteCloserErrCb func(err error, size int64, ch string, me *WriteCloserWithCb) error

type WriteCloserWithCb struct {
	Underlying io.WriteCloser
	h          hash.Hash
	written    int64
	closeCb    WriteCloserErrCb
	tempFile   afero.File
}

func (wcwc *WriteCloserWithCb) Write(p []byte) (n int, err error) {
	if n, err = wcwc.Underlying.Write(p); err == nil && n > 0 {
		wcwc.written += int64(n)
		wcwc.h.Write(p[0:n])
	}
	return
}

func (wcwc *WriteCloserWithCb) Close() error {
	err := wcwc.closeCb(wcwc.Underlying.Close(), wcwc.written, internal.Sha256ToStr32(wcwc.h.Sum(nil)), wcwc)
	if wcwc.tempFile != nil {
		_ = os.Remove(wcwc.tempFile.Name())
	}
	return err
}

func NewWriteCloserWithCb(underlying io.WriteCloser, closeCb WriteCloserErrCb) io.WriteCloser {
	tempFile, _ := underlying.(afero.File)
	return &WriteCloserWithCb{underlying, sha256.New(), 0, closeCb, tempFile}
}

type TempFileWriteCloserWithCb struct {
	WriteCloserWithCb
	tmpFile afero.File
}

func NewTempFileWriteCloserWithCb(fs afero.Fs, dir, pattern string, closeCb WriteCloserErrCb) (io.WriteCloser, error) {
	tempFile, err := afero.TempFile(fs, dir, pattern)
	if err != nil {
		return nil, fmt.Errorf("in NewTempFileWriteCloserWithCb: %w", err)
	}
	return NewWriteCloserWithCb(tempFile, closeCb), nil
}

type ReadCloserWithCb struct {
	underlying io.Reader
	closeCb    func() error
}

func (rcwc *ReadCloserWithCb) Read(p []byte) (n int, err error) {
	return rcwc.underlying.Read(p)
}

func (rcwc *ReadCloserWithCb) Close() error {
	return rcwc.closeCb()
}

func NewReadCloserWithCb(underlying io.Reader, closeCb func() error) (io.ReadCloser, error) {
	return &ReadCloserWithCb{underlying: underlying, closeCb: closeCb}, nil
}

func (sti StorageInfo) loadStoredInMemory(encrypted bool) (metas map[string]map[int64][]byte) {
	metas = map[string]map[int64][]byte{}
	p2m := sti.Path2Meta
	if encrypted {
		p2m = sti.Path2CMeta
	}
	for _, bs := range p2m {
		var meta Meta
		if err := json.Unmarshal(bs, &meta); err != nil {
			continue
		}
		hn := internal.NameToHashStr32(RemoveSlashIfNsIf(meta.Path, meta.IsNs))
		if _, ok := metas[hn]; !ok {
			metas[hn] = map[int64][]byte{}
		}
		metas[hn][meta.Itime] = bs
	}
	return
}

func getInitStorageInfo() StorageInfo {
	return StorageInfo{
		Path2Meta:     map[string][]byte{},
		Path2CMeta:    map[string][]byte{},
		Path2HnIt:     map[string]SIHnIt{},
		ExistingCs:    map[string]bool{},
		ExistingEcs:   map[string]bool{},
		Path2Content:  map[string]string{},
		Path2CContent: map[string]string{},
		Path2Error:    map[string]error{},
	}
}
