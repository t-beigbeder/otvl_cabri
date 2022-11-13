package cabridss

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/mockfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewFsyDssOk(t *testing.T) {
	_, err := NewFsyDss(FsyConfig{}, "/tmp")
	if err != nil {
		t.Fatalf("NewFsyDss failed for /tmp with error %v", err)
	}
}

func TestNewFsyDssErr(t *testing.T) {
	_, err := NewFsyDss(FsyConfig{}, "/NoSuchFileOrDirectory")
	if err == nil {
		t.Fatalf("NewFsyDss failed for /NoSuchFileOrDirectory")
	}
	_, err = NewFsyDss(FsyConfig{}, "/dev/null")
	if err == nil || !strings.Contains(err.Error(), "not a directory: ") {
		t.Fatalf("NewFsyDss failed for /dev/null with error %v", err)
	}
}

func TestNewFsyDssBase(t *testing.T) {

	startup := func(tfs *testfs.Fs) error {
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

	tfs, err := testfs.CreateFs("TestNewFsyDssBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatalf("TestNewFsyDssBase failed with error %v", err)
	}
	err = dss.Updatens("", time.Now().Unix(), []string{"d"}, nil)
	if err == nil {
		t.Fatalf("Mkns should fail with error mkdir file exists")
	}
	err = dss.Updatens("/d", time.Now().Unix(), []string{"d2"}, nil)
	if err == nil {
		t.Fatalf("Mkns should fail with error namespace / (leading)")
	}
	err = dss.Updatens("d/", time.Now().Unix(), []string{"d2"}, nil)
	if err == nil {
		t.Fatalf("Mkns should fail with error namespace / (trailing)")
	}
	err = dss.Updatens("", time.Now().Unix(), []string{"/d2/", "d2\n/f.txt", "", "f1", "f2", "f3", "f1", "f3"}, nil)
	if err == nil || !strings.Contains(err.Error(), "name(s) [/d2/ d2\n/f.txt  f1 f3] should") {
		t.Fatalf("Mkns should fail with name check errors")
	}
	err = dss.Updatens("", time.Now().Unix(), []string{"d2/"}, nil)
	if err != nil {
		t.Fatalf("TestNewFsyDssBase failed with error %v", err)
	}
	err = dss.Mkns("d2", time.Now().Unix(), []string{"d3/", "f4"}, nil)
	if err != nil {
		t.Fatalf("TestNewFsyDssBase failed with error %v", err)
	}
}

func TestNewFsyDssOsErrors(t *testing.T) {

	startup := func(tfs *testfs.Fs) error {
		return nil
	}

	mkdirErr := func(afs afero.Fs, name string, perm os.FileMode) error {
		if ufpath.Base(name) == "d_err" {
			return fmt.Errorf("mockfs mkdir %s error", name)
		}
		return afs.Mkdir(name, perm)
	}
	createErr := func(afs afero.Fs, name string) (afero.File, error) {
		if ufpath.Base(name) == "_err.txt" {
			return nil, fmt.Errorf("mockfs create %s error", name)
		}
		return afs.Create(name)
	}
	chtimesErr := func(afs afero.Fs, name string, atime time.Time, mtime time.Time) error {
		if ufpath.Base(name) == "dd" {
			return fmt.Errorf("mockfs chtimes %s error", name)
		}
		return afs.Chtimes(name, atime, mtime)
	}

	tfs, err := testfs.CreateFs("TestNewFsyDssOsErrors", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()

	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatalf("TestNewFsyDssOsErrors failed with error %v", err)
	}
	// Dss.SetTfs(tfs)
	cbs := mockfs.MockCbs{AfsMkdir: mkdirErr, AfsCreate: createErr, AfsChtimes: chtimesErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	err = dss.Mkns("", time.Now().Unix(), []string{"d/", "a.txt", "dd/"}, nil)
	if err != nil {
		t.Fatalf("TestNewFsyDssOsErrors failed with error %v", err)
	}
	err = dss.Mkns("d", time.Now().Unix(), []string{"d2/", "a2.txt", "b2.txt", "d_err/"}, nil)
	if err == nil {
		t.Fatalf("TestNewFsyDssOsErrors should fail with Mkdir error")
	}

	err = dss.Mkns("d", time.Now().Unix(), []string{"d3/", "a3.txt", "b3.txt", "_err.txt"}, nil)
	if err == nil {
		t.Fatalf("TestNewFsyDssOsErrors should fail with Create error")
	}

	err = dss.Mkns("dd", time.Now().Unix(), []string{}, nil)
	if err == nil {
		t.Fatalf("TestNewFsyDssOsErrors should fail with Create error")
	}
}

func TestFsyDssLsnsBase(t *testing.T) {
	tfs, err := testfs.CreateFs("TestFsyDssLsnsBase", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatalf("TestFsyDssLsnsBase failed with error %v", err)
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	err = dss.Mkns("", time.Now().Unix(), []string{"d2/"}, nil)
	if err != nil {
		t.Fatalf("TestFsyDssLsnsBase failed with error %v", err)
	}
	err = dss.Mkns("d2", time.Now().Unix(), []string{"d3/", "f4"}, nil)
	if err != nil {
		t.Fatalf("TestFsyDssLsnsBase failed with error %v", err)
	}
	err = dss.Mkns("d2/d3", time.Now().Unix(), []string{"d4a/", "f5", "d4b"}, nil)
	if err != nil {
		t.Fatalf("TestFsyDssLsnsBase failed with error %v", err)
	}
	children0, err := dss.Lsns("")
	if err != nil || len(children0) != 1 || children0[0] != "d2/" {
		t.Fatalf("TestFsyDssLsnsBase failed with error %v or children %v", err, children0)
	}
	children2, err := dss.Lsns("d2")
	if err != nil || len(children2) != 2 {
		t.Fatalf("TestFsyDssLsnsBase failed with error %v or children %v", err, children2)
	}
	children3, err := dss.Lsns("d2/d3")
	if err != nil || len(children3) != 3 {
		t.Fatalf("TestFsyDssLsnsBase failed with error %v or children %v", err, children3)
	}
}

func TestFsyDssLsnsErr(t *testing.T) {
	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		if ufpath.Base(name) == "d2" {
			return nil, fmt.Errorf("mockfs open %s error", name)
		}
		return afs.Open(name)
	}
	readdirErr := func(mfi afero.File, count int) ([]os.FileInfo, error) {
		if ufpath.Base(mfi.Name()) == "d3" {
			return nil, fmt.Errorf("mockfs readdir %s error", mfi.Name())
		}
		return mfi.Readdir(count)
	}

	tfs, err := testfs.CreateFs("TestFsyDssLsnsErr", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	_, err = NewFsyDss(FsyConfig{}, tfs.Path()+"nono")
	if err == nil {
		t.Fatalf("TestFsyDssLsnsErr should fail with no such file or dir error")
	}
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatalf("TestFsyDssLsnsErr failed with error %v", err)
	}
	err = dss.Mkns("", time.Now().Unix(), []string{"d2/"}, nil)
	if err != nil {
		t.Fatalf("TestFsyDssLsnsErr failed with error %v", err)
	}
	err = dss.Mkns("d2", time.Now().Unix(), []string{"d3/", "f4"}, nil)
	if err != nil {
		t.Fatalf("TestFsyDssLsnsErr failed with error %v", err)
	}
	err = dss.Mkns("d2/d3", time.Now().Unix(), []string{"d4a/", "f5", "d4b"}, nil)
	if err != nil {
		t.Fatalf("TestFsyDssLsnsErr failed with error %v", err)
	}

	cbs := mockfs.MockCbs{AfsOpen: openErr, AfiReaddir: readdirErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	_, err = dss.Lsns("d2")
	if err == nil {
		t.Fatalf("TestFsyDssLsnsErr should fail with open error")
	}
	_, err = dss.Lsns("d2/d3")
	if err == nil {
		t.Fatalf("TestFsyDssLsnsErr should fail with readdir error")
	}
}

func TestFsyDssGetContentWriterBase(t *testing.T) {
	startup := func(tfs *testfs.Fs) error {
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

	tfs, err := testfs.CreateFs("TestFsyDssGetContentWriterBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fi.Close()
	fo, err := dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, func(err error, size int64, ch string) {
		if err != nil {
			t.Fatal(err.Error())
		}
		if size != 241 {
			t.Fatalf("TestFsyDssGetContentWriterBase size %d != 241", size)
		}
		if ch != "484f617a695613aac4b346237aa01548" {
			t.Fatalf("TestFsyDssGetContentWriterBase hash %s != %s", ch, "484f617a695613aac4b346237aa01548")
		}
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fo.Close()
	io.Copy(fo, fi)
	_, err = dss.GetContentWriter("/no", time.Now().Unix(), nil, nil)
	if err == nil {
		t.Fatalf("TestFsyDssGetContentWriterBase should fail with err args")
	}
	if isDup, err := dss.IsDuplicate("484f617a695613aac4b346237aa01548"); isDup || err != nil {
		t.Fatalf("TestFsyDssGetContentWriterBase IsDuplicate failed %v %v", isDup, err)
	}
}

func TestFsyDssMtime(t *testing.T) {
	startup := func(tfs *testfs.Fs) error {
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

	tfs, err := testfs.CreateFs("TestFsyDssMtime", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fi.Close()
	tt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	if err = dss.Updatens("d", tt, []string{"a-copy.txt"}, nil); err != nil {
		t.Fatalf(err.Error())
	}
	dfi, err := os.Stat(ufpath.Join(tfs.Path(), "d"))
	if dfi.ModTime().Unix() != tt {
		t.Fatalf("TestFsyDssMtime mtime 'd' %d != %d", dfi.ModTime().Unix(), tt)
	}
	fo, err := dss.GetContentWriter("d/a-copy.txt", tt, nil, func(err error, size int64, ch string) {
		if err != nil {
			t.Fatal(err.Error())
		}
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	io.Copy(fo, fi)
	fo.Close()
	ffi, err := os.Stat(ufpath.Join(tfs.Path(), "d", "a-copy.txt"))
	if ffi.ModTime().Unix() != tt {
		t.Fatalf("TestFsyDssMtime mtime 'd/a-copy.txt' %d != %d", ffi.ModTime().Unix(), tt)
	}
}

func TestFsyDssGetContentReaderBase(t *testing.T) {
	startup := func(tfs *testfs.Fs) error {
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

	tfs, err := testfs.CreateFs("TestFsyDssGetContentReaderBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fi.Close()
	fo, err := dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fo.Close()
	io.Copy(fo, fi)
	fi2, err := dss.GetContentReader("a-copy.txt")
	defer fi2.Close()
	fo2, err := dss.GetContentWriter("a-copy-copy.txt", time.Now().Unix(), nil, func(err error, size int64, ch string) {
		if err != nil {
			t.Fatal(err.Error())
		}
		if size != 241 {
			t.Fatalf("TestFsyDssGetContentReaderBase size %d != 241", size)
		}
		if ch != "484f617a695613aac4b346237aa01548" {
			t.Fatalf("TestFsyDssGetContentReaderBase hash %s != %s", ch, "484f617a695613aac4b346237aa01548")
		}
	})
	io.Copy(fo2, fi2)
	fo2.Close()
}

func TestFsyDssOsErrors(t *testing.T) {
	step := ""
	startup := func(tfs *testfs.Fs) error {
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
	createErr := func(afs afero.Fs, name string) (afero.File, error) {
		if step == "ce2" {
			return nil, fmt.Errorf("mockfs create %s error %s", name, step)
		}
		return afs.Create(name)
	}
	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		if step == "oe1" {
			return nil, fmt.Errorf("mockfs open %s error %s", name, step)
		}
		return afs.Open(name)
	}
	closeErr := func(mfi afero.File) error {
		if step == "ce1" {
			return fmt.Errorf("mockfs Close %s error %s", mfi.Name(), step)
		}
		return mfi.Close()
	}
	chtimesErr := func(afs afero.Fs, name string, atime time.Time, mtime time.Time) error {
		if step == "cte1" {
			return fmt.Errorf("mockfs chtimes %s error", name)
		}
		return afs.Chtimes(name, atime, mtime)
	}

	tfs, err := testfs.CreateFs("TestFsyDssOsErrors", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsCreate: createErr, AfsOpen: openErr, AfsChtimes: chtimesErr, AfiClose: closeErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fi.Close()

	step = "ce1"
	fo, err := dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	io.Copy(fo, fi)
	err = fo.Close()
	if err == nil {
		t.Fatalf("TestFsyDssOsErrors should fail with Close error")
	}

	step = "cte1"
	fi2, err := dss.GetContentReader("a.txt")
	defer fi2.Close()
	fo2, err := dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, func(err error, size int64, ch string) {
		if err == nil {
			t.Fatalf("TestFsyDssOsErrors should fail with chtimes error")
		}
	})
	io.Copy(fo2, fi2)
	fo2.Close()

	step = "ce2"
	fi3, err := dss.GetContentReader("a.txt")
	defer fi3.Close()
	_, err = dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, nil)
	if err == nil {
		t.Fatalf("TestFsyDssOsErrors should fail with create error")
	}

	step = "oe1"
	_, err = dss.GetContentReader("a.txt")
	if err == nil {
		t.Fatalf("TestFsyDssOsErrors should fail with open error")
	}

}

func TestFsyDssRemoveBase(t *testing.T) {
	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
			return err
		}
		if err := os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("e/se/c1.txt", 20); err != nil {
			return err
		}
		if err := tfs.RandTextFile("e/se/c2.txt", 20); err != nil {
			return err
		}
		return nil
	}
	createErr := func(afs afero.Fs, name string) (afero.File, error) {
		return afs.Create(name)
	}
	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		return afs.Open(name)
	}
	closeErr := func(mfi afero.File) error {
		return mfi.Close()
	}
	chtimesErr := func(afs afero.Fs, name string, atime time.Time, mtime time.Time) error {
		return afs.Chtimes(name, atime, mtime)
	}

	tfs, err := testfs.CreateFs("TestFsyDssRemoveBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsCreate: createErr, AfsOpen: openErr, AfsChtimes: chtimesErr, AfiClose: closeErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	if err = dss.Remove("/z"); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("//"); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("/"); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("nosuchdir/"); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("nosuchfile"); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("e/se"); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with is a dir error")
	}
	if err = dss.Remove("e/se/c1.txt/"); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with is a file error")
	}
	if err = dss.Remove("e/se/c2.txt"); err != nil {
		t.Fatalf("TestFsyDssRemoveBase %v", err)
	}
	if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/se/c2.txt")); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with no such file e/se/c2.txt")
	}
	if err = dss.Remove("e/"); err != nil {
		t.Fatalf("TestFsyDssRemoveBase %v", err)
	}
	if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/se/c1.txt")); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with no such file e/se/c2.txt")
	}
	if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/se/")); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with no such dir e/se/")
	}
	if _, err = os.Stat(ufpath.Join(tfs.Path(), "e/")); err == nil {
		t.Fatalf("TestFsyDssRemoveBase should fail with no such dir e/")
	}
	_ = err
}

func TestFsyDssGetMetaBasic(t *testing.T) {
	step := ""
	subStep := 0

	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
			return err
		}
		if err := os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("e/se/c1.txt", 20); err != nil {
			return err
		}
		if err := tfs.RandTextFile("e/se/c2.txt", 20); err != nil {
			return err
		}
		return nil
	}
	createErr := func(afs afero.Fs, name string) (afero.File, error) {
		return afs.Create(name)
	}
	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		subStep++
		if step == "oe1" && subStep == 1 {
			return nil, fmt.Errorf("mockfs open %s error %s", name, step)
		}
		if step == "oe2" && subStep == 1 {
			return nil, fmt.Errorf("mockfs open %s error %s", name, step)
		}
		return afs.Open(name)
	}
	readErr := func(mfi afero.File, p []byte) (int, error) {
		subStep++
		if step == "oe3" && subStep == 3 {
			return 0, fmt.Errorf("mockfs read %s error %s", mfi.Name(), step)
		}
		return mfi.Read(p)
	}

	tfs, err := testfs.CreateFs("TestFsyDssGetMetaBasic", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsCreate: createErr, AfsOpen: openErr, AfiRead: readErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	if _, err = dss.GetMeta("/z", true); err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with params error")
	}
	if _, err = dss.GetMeta("//", true); err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with params error")
	}
	if _, err = dss.GetMeta("nosuchdir/", true); err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with params error")
	}
	if _, err = dss.GetMeta("nosuchfile", true); err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with params error")
	}
	if _, err = dss.GetMeta("e/se", true); err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with is a dir error")
	}
	if _, err = dss.GetMeta("e/se/c1.txt/", true); err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with is a file error")
	}

	step = "oe1"
	subStep = 0
	_, err = dss.GetMeta("d/", true)
	if err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with open dir error")
	}

	step = "oe2"
	subStep = 0
	_, err = dss.GetMeta("a.txt", true)
	if err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with open file error")
	}

	step = "oe3"
	subStep = 0
	_, err = dss.GetMeta("a.txt", true)
	if err == nil {
		t.Fatalf("TestFsyDssGetMetaBasic should fail with open file error")
	}

	step = ""
	meta1, err := dss.GetMeta("a.txt", true)
	if err != nil {
		t.Fatal(err.Error())
	}
	if meta1.GetSize() != 241 {
		t.Fatalf("TestFsyDssGetMetaBasic size a.txt %d != 241", meta1.GetSize())
	}
	if meta1.GetCh() != "484f617a695613aac4b346237aa01548" {
		t.Fatalf("TestFsyDssGetMetaBasic ch a.txt %s != 484f617a695613aac4b346237aa01548", meta1.GetCh())
	}

	meta2, err := dss.GetMeta("d/", true)
	if err != nil {
		t.Fatal(err.Error())
	}
	if meta2.GetSize() != 6 {
		t.Fatalf("TestFsyDssGetMetaBasic size d/ %d != 6", meta2.GetSize())
	}
	if meta2.GetCh() != "c880c199d0db1b5a2018f30227dacea8" {
		t.Fatalf("TestFsyDssGetMetaBasic ch d/ %s != c880c199d0db1b5a2018f30227dacea8", meta2.GetCh())
	}

	tt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	dss.Updatens("d", tt, []string{"a-copy.txt"}, nil)
	meta3, err := dss.GetMeta("d/", true)
	if meta3.GetMtime() != tt {
		t.Fatalf("TestFsyDssGetMetaBasic mtime d/ %d != %d", meta3.GetMtime(), tt)
	}
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi.Close()
	fo, err := dss.GetContentWriter("d/a-copy.txt", tt, nil, nil)
	io.Copy(fo, fi)
	fo.Close()
	ffi, err := dss.GetMeta("d/a-copy.txt", true)
	if ffi.GetMtime() != tt {
		t.Fatalf("TestFsyDssGetMetaBasic mtime 'd/a-copy.txt' %d != %d", ffi.GetMtime(), tt)
	}

	meta4, err := dss.GetMeta("d/a-copy.txt", false)
	if v := meta4.GetChUnsafe(); v != "" {
		t.Fatalf("TestFsyDssGetMetaBasic GetChUnsafe failed %s", v)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestFsyDssGetMetaBasic: meta.Ch() did not panic")
		}

	}()
	meta4.GetCh()
	if meta4.GetAcl() != nil {
		t.Fatalf("TestFsyDssGetMetaBasic acl 'd/a-copy.txt' non nil")
	}

}

func TestFsyDssUpdateNsBasic(t *testing.T) {
	step := ""
	subStep := 0

	mkns := func(dss Dss, npath string, tt int64, children []string) error {
		err := dss.Mkns(npath, tt, children, nil)
		if err != nil {
			t.Fatalf("TestFsyDssUpdateNsBasic: mkns failed with error %v", err)
		}
		return nil
	}
	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		subStep++
		if step == "o1" || step == "o2" {
			if subStep == 1 {
				return nil, fmt.Errorf("mockfs open %s error %s", name, step)
			}
			return afs.Open(name)
		}
		return afs.Open(name)
	}
	removeAllErr := func(afs afero.Fs, name string) error {
		subStep++
		if step == "r1" {
			if subStep == 2 {
				return fmt.Errorf("mockfs removeAll %s error", name)
			}
		}
		return afs.RemoveAll(name)
	}

	updns := func(dss Dss, npath string, tt int64, children []string) error {
		err := dss.Updatens(npath, tt, children, nil)
		if err != nil {
			t.Fatalf("TestFsyDssUpdateNsBasic: updatens failed with error %v", err)
		}
		return nil
	}

	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
			return err
		}
		if err := os.MkdirAll(ufpath.Join(tfs.Path(), "e", "se"), 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(ufpath.Join(tfs.Path(), "e", "sf"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("e/se/c1.txt", 20); err != nil {
			return err
		}
		if err := tfs.RandTextFile("e/se/c2.txt", 20); err != nil {
			return err
		}
		return nil
	}

	tfs, err := testfs.CreateFs("TestFsyDssUpdateNsBasic", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsOpen: openErr, AfsRemoveAll: removeAllErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()

	mkns(dss, "e/sf", ttt, []string{"g1.txt", "g2.txt"})
	mt1, err := dss.GetMeta("e/sf/", true)
	if err != nil || mt1.GetSize() != 14 {
		t.Fatalf("TestFsyDssUpdateNsBasic failed %v %v", err, mt1)
	}

	if err = dss.Mkns("e/sf", ttt, []string{"g2.txt", "g3b.txt"}, nil); err == nil {
		t.Fatalf("TestFsyDssUpdateNsBasic Mkns should fail dir exists")
	}
	step = "o1"
	subStep = 0
	if err = dss.Mkns("e/sf", ttt, []string{"g2.txt", "g3b.txt"}, nil); err == nil {
		t.Fatalf("TestFsyDssUpdateNsBasic Mkns should fail sys err")
	}
	step = "o2"
	subStep = 0
	if err = dss.Updatens("e/sf", ttt, []string{"g2.txt", "g3b.txt"}, nil); err == nil {
		t.Fatalf("TestFsyDssUpdateNsBasic Updatens should fail sys err")
	}
	step = ""

	updns(dss, "e/sf", ttt, []string{"g2.txt", "g3b.txt"})
	mt2, err := dss.GetMeta("e/sf/", true)
	if err != nil || mt2.GetSize() != 15 {
		t.Fatalf("TestFsyDssUpdateNsBasic failed %v %v", err, mt1)
	}

	updns(dss, "e/sf", ttt, []string{"g4.txt", "g5/"})
	updns(dss, "e/sf/g5", ttt, []string{"h6.txt", "h7/"})
	updns(dss, "e/sf", ttt, []string{"null"})

	updns(dss, "e/sf", ttt, []string{"g4.txt", "g5/"})
	updns(dss, "e/sf/g5", ttt, []string{"h6.txt", "h7/"})
	step = "r1"
	subStep = 0
	if err = dss.Updatens("e/sf", ttt, []string{"null"}, nil); err == nil {
		t.Fatalf("TestFsyDssUpdateNsBasic Updatens should fail sys err")
	}
	step = ""

}

func TestIrregular(t *testing.T) {

	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
			return err
		}
		if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
			return err
		}
		if err := os.Symlink(ufpath.Join(tfs.Path(), "a.txt"), ufpath.Join(tfs.Path(), "d", "link.a.txt")); err != nil {
			return err
		}
		return nil
	}

	tfs, err := testfs.CreateFs("TestIrregular", startup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err)
	}
	cs, err := dss.Lsns("d")
	if err != nil || len(cs) != 1 {
		t.Fatalf("%v %v", err, cs)
	}
}

func TestFsyStat(t *testing.T) {
	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		return nil
	}

	tfs, err := testfs.CreateFs("TestFsyStat", startup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	dss, err := NewFsyDss(FsyConfig{}, tfs.Path())
	if err != nil {
		t.Fatal(err)
	}
	if err = dss.Updatens("", time.Now().Unix(), []string{"a.txt", "d1/", "d2/"}, nil); err != nil {
		t.Fatal(err)
	}
	if err = dss.Mkns("d1", time.Now().Unix(), nil, []ACLEntry{{User: "x-uid:1000", Rights: Rights{Read: true, Execute: true}}}); err != nil {
		t.Fatal(err)
	}
	_, err = dss.GetMeta("d1/", true)
	if err != nil {
		t.Fatal(err)
	}
	if err = dss.Mkns("d2", time.Now().Unix(), []string{"f2.txt"}, []ACLEntry{
		{User: "x-uid:1000", Rights: Rights{Read: true, Write: true, Execute: true}},
		{User: "x-gid:1000", Rights: Rights{Read: true, Execute: true}},
		{User: "x-other", Rights: Rights{Execute: true}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = dss.GetMeta("d2/", true); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fi.Close()
	fo, err := dss.GetContentWriter("d2/f2.txt", time.Now().Unix(), []ACLEntry{
		{User: "x-uid:1000", Rights: Rights{Read: true, Write: true, Execute: true}},
		{User: "x-gid:1000", Rights: Rights{Read: true, Execute: true}},
		{User: "x-other", Rights: Rights{Execute: true}},
	}, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	io.Copy(fo, fi)
	fo.Close()
	m, err := dss.GetMeta("d2/f2.txt", true)
	if err != nil {
		t.Fatal(err)
	}
	m2 := Meta{Path: m.GetPath(), Mtime: m.GetMtime(), Size: m.GetSize(), Ch: m.GetCh(), ACL: []ACLEntry{
		{User: "x-other", Rights: Rights{Execute: true}},
		{User: "x-gid:1000", Rights: Rights{Read: true, Execute: true}},
		{User: "x-uid:1000", Rights: Rights{Read: true, Write: true, Execute: true}},
	}}
	if !m2.Equals(m, true) {
		t.Fatalf("IMeta Equals")
	}
}

func Test_setSysAcl(t *testing.T) {
	tfs, err := testfs.CreateFs("Test_setSysAcl", tfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	acl := []ACLEntry{
		{User: "x-uid:1000", Rights: Rights{Read: true, Write: true, Execute: true}},
		{User: "x-gid:1000", Rights: Rights{Read: true, Write: true, Execute: true}},
		{User: "x-other", Rights: Rights{Read: true, Write: true, Execute: true}},
	}
	if err = setSysAcl(ufpath.Join(tfs.Path(), "d/b.txt"), acl); err != nil {
		t.Error(err)
	}
}
