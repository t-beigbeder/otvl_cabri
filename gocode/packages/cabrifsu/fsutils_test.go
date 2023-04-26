package cabrifsu

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"testing"
)

func TestEnableWrite(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestEnableWrite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	afs := afero.NewOsFs()

	if err = os.WriteFile(ufpath.Join(dir, "f1.txt"), []byte("content 1\n"), 0); err != nil {
		t.Fatal(err)
	}
	if err = EnableWrite(afs, ufpath.Join(dir, "f1.txt"), false); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(dir, "d1"), 0); err != nil {
		t.Fatal(err)
	}
	if err = EnableWrite(afs, ufpath.Join(dir, "d1"), false); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(dir, "d2"), 0777); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ufpath.Join(dir, "d2", "f2.txt"), []byte("content 2\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(dir, "d2", "d2a"), 0777); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ufpath.Join(dir, "d2", "d2a", "f2a1.txt"), []byte("content 2a1\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ufpath.Join(dir, "d2", "d2a", "f2a2.txt"), []byte("content 2a2\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(dir, "d2", "d2b"), 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(dir, "d2", "d2a", "d2az"), 0777); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ufpath.Join(dir, "d2", "d2a", "d2az", "f2az.txt"), []byte("content 2az\n"), 0666); err != nil {
		t.Fatal(err)
	}
	if err = DisableWrite(afs, dir, true); err != nil {
		t.Fatal(err)
	}
	if err = EnableWrite(afs, dir, true); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ufpath.Join(dir, "d3"), 0511); err != nil {
		t.Fatal(err)
	}
	var m os.FileMode
	var ro bool
	if m, ro, err = GetFileMode(afs, ufpath.Join(dir, "d2")); err != nil {
		t.Fatal(err)
	}
	s2 := fmt.Sprintf("%o %v %o", m, ro, m&0200)
	if m, ro, err = GetFileMode(afs, ufpath.Join(dir, "d3")); err != nil {
		t.Fatal(err)
	}
	s3 := fmt.Sprintf("%o %v %o", m, ro, m&0200)
	_, _ = s2, s3
}
