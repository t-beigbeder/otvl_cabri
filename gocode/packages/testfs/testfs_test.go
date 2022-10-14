//go:build test_testfs

package testfs

import (
	"errors"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {
	tfs, err := CreateFs("testBasic", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	if tfs.Path() == "" {
		t.Fatal("path is empty")
	}
}

func TestCreateError(t *testing.T) {
	tfs, err := CreateFs("testCreateError/err", nil)
	if err == nil {
		defer tfs.Delete()
		t.Fatal("err is nil")
	}
}

func TestWithFunc(t *testing.T) {

	startup := func(f *Fs) (e error) {
		return
	}

	tfs, err := CreateFs("testWithFunc", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
}

func TestWithErrorFunc(t *testing.T) {

	startup := func(f *Fs) error {
		return errors.New("error from TestWithErrorFunc startup")
	}

	tfs, err := CreateFs("testWithFunc", startup)
	if err == nil || tfs != nil {
		t.Fatal("err is nil or tfs is not nil")
	}
}

func TestRand(t *testing.T) {

	startup := func(f *Fs) error {
		if w1 := f.RandomWord(); w1 != "normal" {
			t.Errorf("RandomWord w1 is %s", w1)
			return nil
		}
		if i1 := f.Rand().Int63(); i1 != 608747136543856411 {
			t.Errorf("Rand.Int63n i1 is %d", i1)
			return nil
		}
		return nil
	}

	tfs, err := CreateFs("testRand", startup)
	defer tfs.Delete()

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestRandTextFile(t *testing.T) {

	startup := func(tfs *Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
			return err
		}
		return nil
	}

	tfs, err := CreateFs("TestRandTextFile", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()

	sa, err := tfs.FileAsText("a.txt")
	if err != nil {
		t.Fatal(err.Error())
	}
	if ln := len(strings.Split(sa, "\n")); ln != 39 {
		t.Fatalf("count is %d", ln)
	}
	sb, err := tfs.FileAsText("d/b.txt")
	if err != nil {
		t.Fatal(err.Error())
	}
	if ln := len(strings.Split(sb, "\n")); ln != 9 {
		t.Fatalf("count is %d", ln)
	}
}

func TestOsErrors(t *testing.T) {
	var curLine int
	var curBase string
	tfs, err := CreateFs("TestOsErrors", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	tfs.SetOsCreate(func(path string) (*os.File, error) {
		fp, err := os.Create(path)
		curBase = ufpath.Base(path)
		if curBase == "a.txt" {
			return fp, errors.New("error from TestOsErrors osCreate")
		}
		return fp, err
	})
	tfs.SetOsFileWriteString(func(fp *os.File, s string) (int, error) {
		curLine++
		n, err := fp.WriteString(s)
		if curBase != "b.txt" || curLine < 2 {
			return n, err
		}
		return n, errors.New("error from TestOsErrors osFileWriteString")
	})
	tfs.SetIoutilReadFile(func(path string) (content []byte, err error) {
		content, err = ioutil.ReadFile(path)
		if ufpath.Base(path) == "c.txt" {
			return content, errors.New("error from TestOsErrors ioutilReadFile")
		}
		return content, err
	})
	if err := tfs.RandTextFile("a.txt", 61); err == nil || err.Error() != "error from TestOsErrors osCreate" {
		t.Fatal("no simulated error detected")
	}
	if err := tfs.RandTextFile("b.txt", 62); err == nil || err.Error() != "error from TestOsErrors osFileWriteString" {
		t.Fatal("no simulated error detected")
	}
	if err := tfs.RandTextFile("c.txt", 63); err != nil {
		t.Fatal(err.Error())
	}
	if _, err := tfs.FileAsText("c.txt"); err == nil || err.Error() != "error from TestOsErrors ioutilReadFile" {
		t.Fatal("no simulated error detected")
	}

}

func TestNonWritable(t *testing.T) {
	tfs, err := CreateFs("TestNonWritable", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	tfs.SetNoPlugin()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("no readonly panic")
			}
		}()
		tfs.SetOsCreate(os.Create)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("no readonly panic")
			}
		}()
		tfs.SetOsFileWriteString((*os.File).WriteString)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("no readonly panic")
			}
		}()
		tfs.SetIoutilReadFile(ioutil.ReadFile)
	}()

}
