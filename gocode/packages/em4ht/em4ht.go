package em4ht

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type FS struct {
	sfs    fs.FS
	gMtime time.Time
}

type file struct {
	ef fs.File
	fs *FS
}

type fileInfo struct {
	fi fs.FileInfo
	fs *FS
}

func (f fileInfo) Name() string {
	return f.fi.Name()
}

func (f fileInfo) Size() int64 {
	return f.fi.Size()
}

func (f fileInfo) Mode() fs.FileMode {
	return f.fi.Mode()
}

func (f fileInfo) ModTime() time.Time {
	return f.fs.gMtime
}

func (f fileInfo) IsDir() bool {
	return f.fi.IsDir()
}

func (f fileInfo) Sys() interface{} {
	return f.fi.Sys()
}

func (fs FS) Open(name string) (fs.File, error) {
	ef, err := fs.sfs.Open(name)
	if err != nil {
		return nil, err
	}
	f := file{ef, &fs}
	return f, nil
}

func (f file) Stat() (fs.FileInfo, error) {
	fi, err := f.ef.Stat()
	if err != nil {
		return nil, err
	}
	return fileInfo{fi, f.fs}, nil
}

func (f file) Read(bytes []byte) (int, error) {
	return f.ef.Read(bytes)
}

func (f file) Close() error {
	return f.ef.Close()
}

func NewEm4htFS(efs embed.FS, rootDir string) (*FS, error) {
	efn, err := os.Executable()
	if err != nil {
		return nil, err
	}
	ef, err := os.Open(efn)
	if err != nil {
		return nil, err
	}
	st, err := ef.Stat()
	if err != nil {
		return nil, err
	}
	sfs, err := fs.Sub(efs, rootDir)
	return &FS{sfs, st.ModTime()}, nil
}

type spaHttpFileSystem struct {
	fs     http.FileSystem
	prefix string
	favico string
}

func (sfs spaHttpFileSystem) Open(path string) (http.File, error) {
	log.Println("access", path)
	if path != "/" && path != sfs.favico && !strings.HasPrefix(path, "/"+sfs.prefix+"/") {
		path = "/index.html"
	}
	f, err := sfs.fs.Open(path)
	if err != nil {
		log.Printf("Error opening %s %v", path, err)
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := "index.html"
		if _, err := sfs.fs.Open(index); err != nil {
			log.Printf("Error opening %s %v", index, err)
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}
	}
	return f, nil
}

func NewSpaFileSystem(efs embed.FS, rootDir string, prefix string, favico string) (http.FileSystem, error) {
	e4hfs, err := NewEm4htFS(efs, rootDir)
	if err != nil {
		return nil, err
	}
	return spaHttpFileSystem{http.FS(e4hfs), prefix, favico}, nil
}
