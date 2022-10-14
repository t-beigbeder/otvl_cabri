package cabridss

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
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

type WriteCloserWithCb struct {
	wc      io.WriteCloser
	closeCb func(err error) error
}

func (wcwc *WriteCloserWithCb) Write(p []byte) (n int, err error) {
	return wcwc.wc.Write(p)
}

func (wcwc *WriteCloserWithCb) Close() error {
	return wcwc.closeCb(wcwc.wc.Close())
}

func NewWriteCloserWithCb(underlying io.WriteCloser, closeCb func(err error) error) io.WriteCloser {
	return &WriteCloserWithCb{underlying, closeCb}
}
