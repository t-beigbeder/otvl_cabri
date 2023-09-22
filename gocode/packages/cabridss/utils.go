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

func (sti StorageInfo) loadStoredInMemory() (metas map[string]map[int64][]byte) {
	metas = map[string]map[int64][]byte{}
	for _, bs := range sti.Path2Meta {
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

type PipeWithCb struct {
	rcs   chan struct{}
	size  int
	cb    func(err error, size int64, ch string, data interface{})
	hasCh bool
	h     hash.Hash
}

type PipeReaderWithCb struct {
	pr   *io.PipeReader
	pwcb *PipeWithCb
}

func (pr *PipeReaderWithCb) Read(data []byte) (n int, err error) {
	n, err = pr.pr.Read(data)
	return n, err
}

func (pr *PipeReaderWithCb) Close() error {
	return pr.CloseWithError(nil)
}

func (pr *PipeReaderWithCb) CloseWithError(err error) error {
	close(pr.pwcb.rcs)
	return pr.pr.CloseWithError(err)
}

type PipeWriterWithCb struct {
	pw     *io.PipeWriter
	pwcb   *PipeWithCb
	cbData interface{}
}

func (pw *PipeWriterWithCb) Write(data []byte) (n int, err error) {
	n, err = pw.pw.Write(data)
	if pw.pwcb.cb != nil {
		pw.pwcb.size += n
		if pw.pwcb.hasCh {
			pw.pwcb.h.Write(data[0:n])
		}
	}
	return
}
func (pw *PipeWriterWithCb) SetCbData(data interface{}) {
	pw.cbData = data
}

func (pw *PipeWriterWithCb) Close() (err error) {
	err = pw.pw.Close()
	<-pw.pwcb.rcs
	if pw.pwcb.cb != nil {
		ch := ""
		if pw.pwcb.hasCh {
			ch = internal.Sha256ToStr32(pw.pwcb.h.Sum(nil))
		}
		pw.pwcb.cb(err, int64(pw.pwcb.size), ch, pw.cbData)
	}
	return
}

func NewPipeWithCb(cb func(err error, size int64, ch string, data interface{}), hasCh bool) (*PipeReaderWithCb, *PipeWriterWithCb) {
	pr, pw := io.Pipe()
	pwcb := &PipeWithCb{cb: cb, hasCh: hasCh, h: sha256.New(), rcs: make(chan struct{})}
	return &PipeReaderWithCb{pr: pr, pwcb: pwcb}, &PipeWriterWithCb{pw: pw, pwcb: pwcb}
}
