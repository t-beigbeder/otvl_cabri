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

func TestSetGid(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestSetgid")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	afs := afero.NewOsFs()
	checkFile := func(path string, ruio, rhfa, chm bool) error {
		uio, hfa, err := HasFileWriteAccess(fmt.Sprintf("%s/%s", dir, path))
		if err != nil || uio != ruio || hfa != rhfa {
			return fmt.Errorf("path %s err %v uio %v != %v hfa %v != %v", path, err, uio, ruio, hfa, rhfa)
		}
		fi, err := os.Stat(fmt.Sprintf("%s/%s", dir, path))
		if err != nil {
			return fmt.Errorf("os.Stat path %s err %v", path, err)
		}
		uio, hfa, err = HasFileWriteAccess(fi)
		if err != nil || uio != ruio || hfa != rhfa {
			return fmt.Errorf("path %s err %v uio %v != %v hfa %v != %v", path, err, uio, ruio, hfa, rhfa)
		}
		if err = EnableWrite(afs, fmt.Sprintf("%s/%s", dir, path), false); err != nil {
			if chm {
				return fmt.Errorf("EnableWrite path %s err %v", path, err)
			}
		}
		return nil
	}

	so, se, err := runCommand(fmt.Sprintf("sudo chown root %s", dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	defer func() {
		runCommand(fmt.Sprintf("sudo chown -R %d %s", os.Getgid(), dir))
	}()
	so, se, err = runCommand(fmt.Sprintf("sudo chgrp %d %s", os.Getgid(), dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	so, se, err = runCommand(fmt.Sprintf("sudo chmod g+srwx %s", dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	so, se, err = runCommand(fmt.Sprintf("touch %s/f", dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	so, se, err = runCommand(fmt.Sprintf("mkdir %s/d", dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	so, se, err = runCommand(fmt.Sprintf("sudo touch %s/fr %s/fr2", dir, dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	so, se, err = runCommand(fmt.Sprintf("sudo chmod g+w %s/fr2", dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	so, se, err = runCommand(fmt.Sprintf("sudo mkdir %s/dr %s/dr2", dir, dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	so, se, err = runCommand(fmt.Sprintf("sudo chmod g+wX %s/dr2", dir))
	if err != nil {
		t.Fatal(string(so), string(se), err)
	}
	if err = checkFile("f", true, true, true); err != nil {
		t.Fatal(err)
	}
	if err = checkFile("d", true, true, true); err != nil {
		t.Fatal(err)
	}
	if err = checkFile("fr", false, false, false); err != nil {
		t.Fatal(err)
	}
	if err = checkFile("dr", false, false, false); err != nil {
		t.Fatal(err)
	}
	if err = checkFile("fr2", false, true, true); err != nil {
		t.Fatal(err)
	}
	if err = checkFile("dr2", false, true, true); err != nil {
		t.Fatal(err)
	}
}

func TestSimuWin(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestSimuWin")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	afs := afero.NewOsFs()
	defer EnableWrite(afs, dir, true)
	if err = os.WriteFile(ufpath.Join(dir, "f1.txt"), []byte("content 1\n"), 0); err != nil {
		t.Fatal(err)
	}
	if io, wa, le := hasFileWriteAccessNotUx(ufpath.Join(dir, "f1.txt")); io != true || wa != true || le != nil {
		t.Fatal(io, wa, err)
	}
	if err = DisableWrite(afs, ufpath.Join(dir, "f1.txt"), false); err != nil {
		t.Fatal(err)
	}
	if io, wa, le := hasFileWriteAccessNotUx(ufpath.Join(dir, "f1.txt")); io != true || wa != false || le != nil {
		t.Fatal(io, wa, err)
	}
	if err := os.Mkdir(ufpath.Join(dir, "d1"), 0); err != nil {
		t.Fatal(err)
	}
	if io, wa, le := hasFileWriteAccessNotUx(ufpath.Join(dir, "d1")); io != true || wa != true || le != nil {
		t.Fatal(io, wa, err)
	}
	if err = DisableWrite(afs, ufpath.Join(dir, "d1"), false); err != nil {
		t.Fatal(err)
	}
	if io, wa, le := hasFileWriteAccessNotUx(ufpath.Join(dir, "d1")); io != true || wa != false || le != nil {
		t.Fatal(io, wa, err)
	}
}
