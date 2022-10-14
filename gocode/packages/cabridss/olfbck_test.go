//go:build nomore

package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/mockfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func noMoreTestCreateOlfDssOk(t *testing.T) {
	tfs, err := testfs.CreateFs("TestCreateOlfDssOk", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	_, err = CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatalf("TestCreateOlfDssOk failed with error %v", err)
	}
}

func noMoreTestCreateOlfDssErr(t *testing.T) {
	tfs, err := testfs.CreateFs("TestCreateOlfDssErr", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	_, err = CreateOlfDss(ufpath.Join(tfs.Path(), "no"), "l")
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with no such file or directory error")
	}
	_, err = CreateOlfDss("/dev/null", "l")
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with not a diretory error")
	}
	_ = os.Mkdir(ufpath.Join(tfs.Path(), "meta"), 0o777)
	_, err = CreateOlfDss(tfs.Path(), "l")
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with cannot create error")
	}
	_ = os.Remove(ufpath.Join(tfs.Path(), "meta"))
	_ = os.Mkdir(ufpath.Join(tfs.Path(), "content"), 0o777)
	_, err = CreateOlfDss(tfs.Path(), "l")
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with cannot create error")
	}
	_ = os.Remove(ufpath.Join(tfs.Path(), "meta"))
	_ = os.Remove(ufpath.Join(tfs.Path(), "content"))
	_ = os.Mkdir(ufpath.Join(tfs.Path(), "tmp"), 0o777)
	_, err = CreateOlfDss(tfs.Path(), "l")
	if err == nil {
		t.Fatalf("TestCreateOlfDssErr should fail with cannot create error")
	}
}

func noMoreTestNewOlfDssOk(t *testing.T) {
	tfs, err := testfs.CreateFs("TestNewOlfDssOk", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	_, err = CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatalf("TestNewOlfDssOk failed with error %v", err)
	}
	_, err = NewOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path()}, 0, nil)
	if err != nil {
		t.Fatalf("TestNewOlfDssOk failed with error %v", err)
	}
}

func noMoreTestNewOlfDssErr1(t *testing.T) {
	tfs, err := testfs.CreateFs("TestNewOlfDssErr1", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	_, err = CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatalf("TestNewOlfDssErr1 failed with error %v", err)
	}
	_, err = NewOlfDss(OlfConfig{Root: ufpath.Join(tfs.Path(), "no")}, 0, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr1 should fail with error no such file or directory")
	}
	_ = os.Remove(ufpath.Join(tfs.Path(), "content"))
	_, err = NewOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path()}, 0, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr1 should fail with error no such file or directory")
	}
	_ = os.Remove(ufpath.Join(tfs.Path(), "meta"))
	_, err = NewOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path()}, 0, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr1 should fail with error no such file or directory")
	}
}

func noMoreTestNewOlfDssErr2(t *testing.T) {
	mkdirAllErr := func(afs afero.Fs, path string, perm os.FileMode) error {
		if ufpath.Base(path) == "b94" {
			return fmt.Errorf("mockfs mkdirAll %s error", path)
		}
		return afs.MkdirAll(path, perm)
	}
	createErr := func(afs afero.Fs, name string) (afero.File, error) {
		if ufpath.Base(name) == "dc35c34c14daabb84fe2fe4cc5" {
			return nil, fmt.Errorf("mockfs create %s error", name)
		}
		if strings.HasPrefix(ufpath.Base(name), "4298fc1c149afbf4c8996fb924.0") {
			return nil, fmt.Errorf("mockfs create %s error", name)
		}
		return afs.Create(name)
	}
	writeErr := func(mfi afero.File, p []byte) (n int, err error) {
		if ufpath.Base(mfi.Name()) == "6cc0e38cf3fdb17a607986855d" {
			return 0, fmt.Errorf("mockfs write %s error", mfi.Name())
		}
		return mfi.Write(p)
	}

	tfs, err := testfs.CreateFs("TestNewOlfDssErr2", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsMkdirAll: mkdirAllErr, AfsCreate: createErr, AfiWrite: writeErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	err = dss.Mkns("/d", time.Now().Unix(), []string{"d2"}, nil)
	if err == nil {
		t.Fatalf("Mkns should fail with error namespace / (leading)")
	}
	err = dss.Mkns("d/", time.Now().Unix(), []string{"d2"}, nil)
	if err == nil {
		t.Fatalf("Mkns should fail with error namespace / (trailing)")
	}
	err = dss.Mkns("", time.Now().Unix(), []string{"/d2/", "d2\n/f.txt", "", "f1", "f2", "f3", "f1", "f3"}, nil)
	if err == nil || !strings.Contains(err.Error(), "name(s) [/d2/ d2\n/f.txt  f1 f3] should") {
		t.Fatalf("Mkns should fail with name check errors")
	}

	csp := "c2a/a1c/5cef47a2ec0c4f6c3ea59d148c"
	_ = os.MkdirAll(ufpath.Join(tfs.Path(), "content", csp), 0o777)
	err = dss.Mkns("", time.Now().Unix(), []string{"d/"}, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr2 should fail with error is a directory")
	}
	_ = os.Remove(ufpath.Join(tfs.Path(), "content", csp))

	// csp now is "9d8/b94/7a7f373db59e8715ec6e348994"
	err = dss.Mkns("", time.Now().Unix(), []string{"dd/"}, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr2 should fail with error MkdirAll")
	}

	// csp now is "1ff/dfe/dc35c34c14daabb84fe2fe4cc5"
	err = dss.Mkns("", time.Now().Unix(), []string{"ddd/"}, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr2 should fail with error Create")
	}

	// csp now is "705/abb/6cc0e38cf3fdb17a607986855d"
	err = dss.Mkns("", time.Now().Unix(), []string{"dddd/"}, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr2 should fail with error Write")
	}

}

func noMoreTestNewOlfDssErr3(t *testing.T) {
	step := ""
	createErr := func(afs afero.Fs, name string) (afero.File, error) {
		return afs.Create(name)
	}
	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		if step == "" && strings.HasPrefix(ufpath.Base(name), "1749c611ba0fa0e16c61831db4.") {
			return nil, fmt.Errorf("mockfs open %s error", name)
		}
		return afs.Open(name)
	}
	writeErr := func(mfi afero.File, p []byte) (n int, err error) {
		if strings.HasPrefix(ufpath.Base(mfi.Name()), "7343f016890c510e93f9352611.0") {
			return 0, fmt.Errorf("mockfs write %s error", mfi.Name())
		}
		return mfi.Write(p)
	}
	marshalErr := func(v interface{}) ([]byte, error) {
		mv, ok := v.(Meta)
		if !ok || mv.Path == "d2/" {
			return nil, fmt.Errorf("marshalErr %v", v)
		}
		return json.Marshal(v)
	}
	closeErr := func(mfi afero.File) error {
		//if step == "ce1" && ufpath.Base(mfi.Name()) == "4298fc1c149afbf4c8996fb924" {
		//	return fmt.Errorf("mockfs Close %s %s error", mfi.Name(), step)
		//}
		return mfi.Close()
	}

	tfs, err := testfs.CreateFs("TestNewOlfDssErr3", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsCreate: createErr, AfsOpen: openErr, AfiWrite: writeErr, AfiClose: closeErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	dss.SetMetaMockCbs(&MetaMockCbs{MockMarshal: marshalErr})
	err = dss.Mkns("", time.Now().Unix(), []string{"d/", "d2/", "d3/"}, nil)
	if err != nil {
		t.Fatalf("TestNewOlfDssErr3 failed with error %v", err)
	}
	// mpath is "/path/to/meta/18a/c3e/7343f016890c510e93f9352611.0*"
	err = dss.Mkns("d", time.Now().Unix(), []string{"d2/"}, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr3 should fail with error Write")
	}
	err = dss.Mkns("d2", time.Now().Unix(), []string{"d2a/"}, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr3 should fail with error Marshal")
	}
	err = dss.Mkns("d3", time.Now().Unix(), []string{"d3a/", "d3b/"}, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	// "d3" => /meta/f45/1a6/1749c611ba0fa0e16c61831db4.0000000061ddc607
	err = dss.Mkns("d3/d3a", time.Now().Unix(), []string{}, nil)
	if err == nil {
		t.Fatalf("TestNewOlfDssErr3 should fail with error open")
	}
	//step = "ce1"
	//err = Dss.Mkns("d3/d3b", time.Now().Unix(), []string{}, nil)
	//if err == nil {
	//	t.Fatalf("TestNewOlfDssErr3 should fail with error Close")
	//}
}

func noMoreTestNewOlfDssBase(t *testing.T) {
	tfs, err := testfs.CreateFs("TestNewOlfDssBase", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	err = dss.Mkns("", time.Now().Unix(), []string{}, nil)
	if err != nil {
		t.Fatalf("TestNewOlfDssBase failed with error %v", err)
	}
	if _, err = os.Stat(ufpath.Join(tfs.Path(), "meta", "e3b", "0c4")); err != nil {
		t.Fatal(err)
	}
	err = dss.Updatens("", time.Now().Unix(), []string{"dd/", "ddd/"}, nil)
	if err != nil {
		t.Fatalf("TestNewOlfDssBase failed with error %v", err)
	}

	err = dss.Mkns("dd", time.Now().Unix(), []string{}, nil)
	if err != nil {
		t.Fatalf("TestNewOlfDssBase failed with error %v", err)
	}
	if _, err = os.Stat(ufpath.Join(tfs.Path(), "meta", "9b7", "ecc")); err != nil {
		t.Fatal(err)
	}

	err = dss.Updatens("dd", time.Now().Unix(), []string{"d2/", "d1", "éh", "eh", "ab", "àbc", "abc", "ab/"}, nil)
	if err != nil {
		t.Fatalf("TestNewOlfDssBase failed with error %v", err)
	}

	dss.SetCurrentTime(int64(0x000000006055b7df))
	err = dss.Mkns("ddd", time.Now().Unix(), []string{"d2/", "d1", "éh", "eh", "ab", "àbc", "abc", "ab/"}, nil)
	if err != nil {
		t.Fatalf("TestNewOlfDssBase failed with error %v", err)
	}
	if _, err = os.Stat(ufpath.Join(tfs.Path(), "meta", "730", "f75", "dafd73e047b86acb2dbd74e75d.000000006055b7df")); err != nil {
		t.Fatal(err)
	}
}

func noMoreTestOlfDssGetContentWriterBase(t *testing.T) {

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
	statErr := func(mfs afero.Fs, name string) (os.FileInfo, error) {
		if ufpath.Base(name) == "7a695613aac4b346237aa01547" {
			return nil, fmt.Errorf("mockfs stat %s error", name)
		}
		return mfs.Stat(name)
	}
	tfs, err := testfs.CreateFs("TestOlfDssGetContentWriterBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsStat: statErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fi.Close()
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	err = dss.Mkns("", ttt, []string{"a.txt"}, nil)
	if err != nil {
		t.Fatalf("TestOlfDssGetContentWriterBase: mkns failed with error %v", err)
	}

	fo, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err != nil {
			t.Fatal(err.Error())
		}
		if size != 241 {
			t.Fatalf("TestOlfDssGetContentWriterBase size %d != 241", size)
		}
		hs := internal.Sha256ToStr32(sha256trunc)
		if hs != "484f617a695613aac4b346237aa01548" {
			t.Fatalf("TestOlfDssGetContentWriterBase %s != %s", hs, "484f617a695613aac4b346237aa01548")
		}
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	io.Copy(fo, fi)
	fo.Close()

	_, err = dss.GetContentWriter("/no", time.Now().Unix(), nil, nil)
	if err == nil {
		t.Fatalf("TestOlfDssGetContentWriterBase should fail with err args")
	}

	_, err = dss.GetContentWriter("no/parent", time.Now().Unix(), nil, nil)
	if err == nil {
		t.Fatalf("TestOlfDssGetContentWriterBase should fail with err no parent")
	}

	if isDup1, err := dss.IsDuplicate("484f617a695613aac4b346237aa01548"); !isDup1 || err != nil {
		t.Fatalf("TestOlfDssGetContentWriterBase IsDuplicate failed %v %v", isDup1, err)
	}

	if isDup2, err := dss.IsDuplicate("484f617a695613aac4b346237aa01549"); isDup2 || err != nil {
		t.Fatalf("TestOlfDssGetContentWriterBase IsDuplicate failed %v %v", isDup2, err)
	}

	if isDup3, err := dss.IsDuplicate("484f617a695613aac4b346237aa01547"); isDup3 || err == nil {
		t.Fatalf("TestOlfDssGetContentWriterBase IsDuplicate failed %v %v", isDup3, err)
	}
}

func noMoreTestOlfDssGetContentReaderBase(t *testing.T) {

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

	tfs, err := testfs.CreateFs("TestOlfDssGetContentReaderBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
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
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	err = dss.Mkns("", ttt, []string{"a.txt", "a-copy.txt"}, nil)
	if err != nil {
		t.Fatalf("TestOlfDssGetContentReaderBase: mkns failed with error %v", err)
	}

	fo, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	io.Copy(fo, fi)
	fo.Close()

	fi2, err := dss.GetContentReader("a.txt")
	defer fi2.Close()
	fo2, err := dss.GetContentWriter("a-copy.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err != nil {
			t.Fatal(err.Error())
		}
		if size != 241 {
			t.Fatalf("TestOlfDssGetContentReaderBase size %d != 241", size)
		}
		hs := internal.Sha256ToStr32(sha256trunc)
		if hs != "484f617a695613aac4b346237aa01548" {
			t.Fatalf("TestOlfDssGetContentReaderBase hash %s != %s", hs, "484f617a695613aac4b346237aa01548")
		}
	})
	io.Copy(fo2, fi2)
	fo2.Close()

	_, err = dss.GetContentReader("/no")
	if err == nil {
		t.Fatalf("TestOlfDssGetContentReaderBase should fail with err args")
	}

	_, err = dss.GetContentReader("no/parent")
	if err == nil {
		t.Fatalf("TestOlfDssGetContentReaderBase should fail with err no parent")
	}

}

func noMoreTestOlfDssOSErrors(t *testing.T) {
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
		return nil
	}
	statErr := func(mfs afero.Fs, name string) (os.FileInfo, error) {
		if step == "se1" && strings.Contains(name, "content") {
			os.MkdirAll(name, 0o777)
			fi, err := mfs.Stat(name)
			os.Remove(name)
			os.Remove(ufpath.Dir(name))
			return fi, err
		}
		if step == "se2" {
			return nil, fmt.Errorf("mockfs stat %s error %s", name, step)
		}
		if step == "se3" {
			return nil, fmt.Errorf("mockfs stat %s error %s", name, step)
		}
		return mfs.Stat(name)
	}
	mkdirallErr := func(afs afero.Fs, name string, perm os.FileMode) error {
		if step == "mda1" {
			return fmt.Errorf("mockfs mkdirall %s error %s", name, step)
		}
		return afs.MkdirAll(name, perm)
	}
	renameErr := func(afs afero.Fs, oldname, newname string) error {
		if step == "rne1" {
			return fmt.Errorf("mockfs rename %s %s error %s", oldname, newname, step)
		}
		return afs.Rename(oldname, newname)
	}
	createErr := func(afs afero.Fs, name string) (afero.File, error) {
		if step == "crm1" {
			return nil, fmt.Errorf("mockfs create %s error %s", name, step)
		}
		return afs.Create(name)
	}
	openFileErr := func(afs afero.Fs, name string, flag int, perm os.FileMode) (afero.File, error) {
		if step == "ofe1" {
			return nil, fmt.Errorf("mockfs openFile %s error %s", name, step)
		}
		return afs.OpenFile(name, flag, perm)
	}
	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		if len(ufpath.Base(name)) > 3 {
			if step == "ope1" {
				subStep++
				if subStep == 3 {
					return nil, fmt.Errorf("mockfs open %s %s %d error", name, step, subStep)
				}
			}
			if step == "ope2" {
				subStep++
				if subStep == 3 {
					return nil, fmt.Errorf("mockfs open %s %s %d error", name, step, subStep)
				}
			}
		}
		return afs.Open(name)
	}

	tfs, err := testfs.CreateFs("TestOlfDssOSErrors", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{
		AfsStat: statErr, AfsMkdirAll: mkdirallErr,
		AfsRename: renameErr, AfsCreate: createErr, AfsOpenFile: openFileErr,
		AfsOpen: openErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fi.Close()
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	err = dss.Mkns("", ttt, []string{"a.txt", "a-copy.txt"}, nil)
	if err != nil {
		t.Fatalf("TestOlfDssOSErrors: mkns failed with error %v", err)
	}

	step = "se1"
	fo, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err == nil {
			t.Fatalf("TestOlfDssOSErrors should fail with stat isdir error")
		}
	})
	io.Copy(fo, fi)
	fo.Close()

	step = "mda1"
	fi1b, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi1b.Close()
	fo1b, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err == nil {
			t.Fatalf("TestOlfDssOSErrors should fail with mkdirall error")
		}
	})
	io.Copy(fo1b, fi1b)
	fo1b.Close()

	step = "rne1"
	fi1c, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi1c.Close()
	fo1c, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err == nil {
			t.Fatalf("TestOlfDssOSErrors should fail with rename error")
		}
	})
	io.Copy(fo1c, fi1c)
	fo1c.Close()

	step = "crm1"
	fi1d, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi1d.Close()
	fo1d, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err == nil {
			t.Fatalf("TestOlfDssOSErrors should fail with create meta error")
		}
	})
	io.Copy(fo1d, fi1d)
	fo1d.Close()

	step = "se2"
	fi1e, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi1e.Close()
	_, err = dss.GetContentWriter("a.txt", time.Now().Unix(), nil, nil)
	if err == nil {
		t.Fatalf("TestOlfDssOSErrors should fail with stat has parent error")
	}

	step = "ofe1"
	fi1f, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi1f.Close()
	_, err = dss.GetContentWriter("a.txt", time.Now().Unix(), nil, nil)
	if err == nil {
		t.Fatalf("TestOlfDssOSErrors should fail with open file temp error")
	}

	step = ""
	fi1g, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi1g.Close()
	fo1g, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err != nil {
			t.Fatalf("TestOlfDssOSErrors should fail with rename error")
		}
	})
	io.Copy(fo1g, fi1g)
	fo1g.Close()

	step = ""
	fi1h, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	defer fi1h.Close()
	fo1h, err := dss.GetContentWriter("a.txt", time.Now().Unix(), nil, func(err error, size int64, sha256trunc []byte) {
		if err != nil {
			t.Fatalf("TestOlfDssOSErrors should fail with rename error")
		}
	})
	io.Copy(fo1h, fi1h)
	fo1h.Close()

	step = "se3"
	_, err = dss.GetContentReader("a.txt")
	if err == nil {
		t.Fatalf("TestOlfDssOSErrors should fail with stat has parent error")
	}

	step = "ope1"
	_, err = dss.GetContentReader("a.txt")
	if err == nil {
		t.Fatalf("TestOlfDssOSErrors should fail with open meta error")
	}

	step = "ope2"
	subStep = 0
	_, err = dss.GetContentReader("a.txt")
	if err == nil {
		t.Fatalf("TestOlfDssOSErrors should fail with open content error")
	}
}

func noMoreTestOlfDssRemoveBase(t *testing.T) {
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
		return nil
	}

	now := time.Now().Unix()
	mkns := func(dss Dss, npath string, children []string) error {
		err := dss.Mkns(npath, now, children, nil)
		if err != nil {
			t.Fatalf("TestOlfDssRemoveBase: mkns failed with error %v", err)
		}
		return nil
	}
	mkcont := func(dss Dss, npath string, tfs *testfs.Fs, oripath string) error {
		fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
		if err != nil {
			t.Fatal(err.Error())
		}
		defer fi.Close()
		fo, err := dss.GetContentWriter(npath, now, nil, nil)
		if err != nil {
			t.Fatalf("TestOlfDssRemoveBase: mkcont failed with error %v", err)
		}
		io.Copy(fo, fi)
		err = fo.Close()
		if err != nil {
			t.Fatalf("TestOlfDssRemoveBase: mkcont failed with error %v", err)
		}

		return nil
	}

	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		if step == "f4err" {
			subStep++
			if subStep == 1 {
				return nil, fmt.Errorf("mockfs open %s %s %d error", name, step, subStep)
			}
		}
		return afs.Open(name)
	}
	statErr := func(mfs afero.Fs, name string) (os.FileInfo, error) {
		if step == "f5err" {
			subStep++
			if subStep == 3 {
				return nil, fmt.Errorf("mockfs open %s %s %d error", name, step, subStep)
			}
		}
		return mfs.Stat(name)
	}
	createErr := func(mfs afero.Fs, name string) (afero.File, error) {
		if step == "d4err" {
			subStep++
			if subStep == 1 {
				return nil, fmt.Errorf("mockfs create %s %s %d error", name, step, subStep)
			}
		}
		return mfs.Create(name)
	}

	tfs, err := testfs.CreateFs("TestOlfDssRemoveBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsOpen: openErr, AfsStat: statErr, AfsCreate: createErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	mkns(dss, "", []string{"d2/", "d2bis/", "f0", "d4err/", "f4err", "f5err", "f6"})
	mkcont(dss, "f0", tfs, "d/b.txt")
	mkcont(dss, "f4err", tfs, "d/b.txt")
	mkcont(dss, "f5err", tfs, "d/b.txt")
	mkcont(dss, "f6", tfs, "d/b.txt")
	mkns(dss, "d2", []string{"d3/", "f3", "d3err/", "f3err"})
	mkns(dss, "d2bis", []string{})
	mkns(dss, "d4err", []string{})
	mkns(dss, "d2/d3", []string{})
	mkcont(dss, "d2/f3", tfs, "a.txt")

	if err = dss.Remove("/z"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("//"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("/"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("nosuchdir/"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with params error")
	}
	if err = dss.Remove("nosuchfile"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with params error")
	}

	if err = dss.Remove("d2/d3/"); err != nil {
		t.Fatalf("TestOlfDssRemoveBase: Remove failed with error %v", err)
	}
	if err = dss.Remove("d2/f3"); err != nil {
		t.Fatalf("TestOlfDssRemoveBase: Remove failed with error %v", err)
	}
	if err = dss.Remove("d2bis/"); err != nil {
		t.Fatalf("TestOlfDssRemoveBase: Remove failed with error %v", err)
	}
	if err = dss.Remove("f0"); err != nil {
		t.Fatalf("TestOlfDssRemoveBase: Remove failed with error %v", err)
	}

	if err = dss.Remove("d2/d3err"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with no such entry error")
	}
	if err = dss.Remove("d2/f3err/"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with no such entry error")
	}

	step = "f4err"
	if err = dss.Remove("f4err"); err == nil {
		t.Fatalf("TestOlfDssRemoveBase should fail with mock open error")
	}
	//step = "f5err"
	//subStep = 0
	//if err = Dss.Remove("f5err"); err != nil {
	//	t.Fatalf("TestOlfDssLsnsBase: Remove failed with error %v", err)
	//}
	//step = ""
	//if err = Dss.Remove("f6"); err != nil {
	//	t.Fatalf("TestOlfDssLsnsBase: Remove failed with error %v", err)
	//}
	//if _, err = Dss.GetMeta("f6", false); err == nil {
	//	t.Fatalf("'f6' not really removed")
	//}
	//step = "d4err"
	//subStep = 0
	//if err = Dss.Remove("d4err/"); err == nil {
	//	t.Fatalf("TestOlfDssRemoveBase should fail with mock create error")
	//}
}

func noMoreTestOlfDssGetMeta(t *testing.T) {
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
		return nil
	}

	now := time.Now().Unix()
	mkns := func(dss Dss, npath string, children []string) error {
		err := dss.Mkns(npath, now, children, nil)
		if err != nil {
			t.Fatalf("TestOlfDssGetMeta: mkns failed with error %v", err)
		}
		return nil
	}
	mkcont := func(dss Dss, npath string, tfs *testfs.Fs, oripath string) error {
		fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
		if err != nil {
			t.Fatal(err.Error())
		}
		defer fi.Close()
		fo, err := dss.GetContentWriter(npath, now, nil, nil)
		if err != nil {
			t.Fatalf("TestOlfDssGetMeta: mkcont failed with error %v", err)
		}
		io.Copy(fo, fi)
		err = fo.Close()
		if err != nil {
			t.Fatalf("TestOlfDssGetMeta: mkcont failed with error %v", err)
		}

		return nil
	}

	openErr := func(afs afero.Fs, name string) (afero.File, error) {
		if step == "f4err" {
			subStep++
			if subStep == 1 {
				return nil, fmt.Errorf("mockfs open %s %s %d error", name, step, subStep)
			}
		}
		if step == "d4err" {
			subStep++
			if subStep == 5 {
				return nil, fmt.Errorf("mockfs open %s %s %d error", name, step, subStep)
			}
		}
		return afs.Open(name)
	}
	statErr := func(mfs afero.Fs, name string) (os.FileInfo, error) {
		if step == "f5err" {
			subStep++
			if subStep == 99 {
				return nil, fmt.Errorf("mockfs open %s %s %d error", name, step, subStep)
			}
		}
		return mfs.Stat(name)
	}

	tfs, err := testfs.CreateFs("TestOlfDssGetMeta", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{AfsOpen: openErr, AfsStat: statErr}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))

	mkns(dss, "", []string{"d2/", "d2bis/", "f0", "d4err/", "f4err", "f5err"})
	mkcont(dss, "f0", tfs, "d/b.txt")
	mkcont(dss, "f4err", tfs, "d/b.txt")
	mkcont(dss, "f5err", tfs, "d/b.txt")
	mkns(dss, "d2", []string{"d3/", "f3", "d3err/", "f3err"})
	mkns(dss, "d2bis", []string{})
	mkns(dss, "d4err", []string{})
	mkns(dss, "d2/d3", []string{})
	mkcont(dss, "d2/f3", tfs, "a.txt")

	if _, err = dss.GetMeta("/z", true); err == nil {
		t.Fatalf("TestOlfDssGetMeta should fail with params error")
	}
	if _, err = dss.GetMeta("//", true); err == nil {
		t.Fatalf("TestOlfDssGetMeta should fail with params error")
	}
	if _, err = dss.GetMeta("nosuchdir/", true); err == nil {
		t.Fatalf("TestOlfDssGetMeta should fail with params error")
	}
	if _, err = dss.GetMeta("nosuchfile", true); err == nil {
		t.Fatalf("TestOlfDssGetMeta should fail with params error")
	}
	if _, err = dss.GetMeta("e/se", true); err == nil {
		t.Fatalf("TestOlfDssGetMeta should fail with is a dir error")
	}
	if _, err = dss.GetMeta("e/se/c1.txt/", true); err == nil {
		t.Fatalf("TestOlfDssGetMeta should fail with is a file error")
	}

	meta1, err := dss.GetMeta("d2/d3/", true)
	if err != nil || meta1.GetSize() != 0 || meta1.GetCh() != "e3b0c44298fc1c149afbf4c8996fb924" {
		t.Fatalf("TestOlfDssGetMeta: GetMeta failed with error %v or meta %v", err, meta1)
	}
	meta1b, err := dss.GetMeta("d2/d3/", false)
	if err != nil || meta1b.GetSize() != 0 || meta1b.GetCh() != "e3b0c44298fc1c149afbf4c8996fb924" ||
		meta1b.GetChUnsafe() != "e3b0c44298fc1c149afbf4c8996fb924" ||
		!meta1b.Equals(meta1, false) || !meta1.Equals(meta1b, false) {
		t.Fatalf("TestOlfDssGetMeta: GetMeta failed with error %v or meta %v", err, meta1b)
	}
	meta2, err := dss.GetMeta("d2/f3", true)
	if err != nil || meta2.GetSize() != 241 || meta2.GetCh() != "484f617a695613aac4b346237aa01548" {
		t.Fatalf("TestOlfDssGetMeta: GetMeta failed with error %v or meta %v", err, meta2)
	}
	if meta2.Equals(meta1, false) || meta2.Equals(nil, false) {
		t.Fatalf("TestOlfDssGetMeta: Meta Equals failed")
	}
	meta3, err := dss.GetMeta("d2bis/", true)
	if err != nil || meta3.GetSize() != 0 || meta3.GetCh() != "e3b0c44298fc1c149afbf4c8996fb924" {
		t.Fatalf("TestOlfDssGetMeta: GetMeta failed with error %v or meta %v", err, meta3)
	}
	meta4, err := dss.GetMeta("f0", true)
	if err != nil || meta4.GetSize() != 241 || meta4.GetCh() != "484f617a695613aac4b346237aa01548" {
		t.Fatalf("TestOlfDssGetMeta: GetMeta failed with error %v or meta %v", err, meta4)
	}

	step = "f4err"
	_, err = dss.GetMeta("f0", true)
	if err == nil {
		t.Fatalf("TestOlfDssGetMeta should fail with mock open error")
	}
	step = "f5err"
	subStep = 0
	_, err = dss.GetMeta("f5err", true)
	//if err == nil {
	//	t.Fatalf("TestOlfDssGetMeta should fail with no content for entry error")
	//}
	//step = "d4err"
	//subStep = 0
	//_, err = Dss.GetMeta("d4err/", true)
	//if err == nil {
	//	t.Fatalf("TestOlfDssGetMeta should fail with no conten for entry error")
	//}

}

func noMoreTestOlfDssHistoryContent(t *testing.T) {
	mkns := func(dss Dss, npath string, tt int64, children []string) error {
		err := dss.Mkns(npath, tt, children, nil)
		if err != nil {
			t.Fatalf("TestOlfDssHistoryContent: mkns failed with error %v", err)
		}
		return nil
	}

	updns := func(dss Dss, npath string, tt int64, children []string) error {
		err := dss.Updatens(npath, tt, children, nil)
		if err != nil {
			t.Fatalf("TestOlfDssHistoryContent: updatens failed with error %v", err)
		}
		return nil
	}

	copyCont := func(tfs *testfs.Fs, dss Dss, fpath string, npath string, tt int64) {
		fi, err := os.Open(ufpath.Join(tfs.Path(), fpath))
		if err != nil {
			t.Fatalf("TestOlfDssHistoryContent: copyCont failed with error %v", err)
		}
		defer fi.Close()
		fo, err := dss.GetContentWriter(npath, tt, nil, nil)
		if err != nil {
			t.Fatalf("TestOlfDssHistoryContent: copyCont failed with error %v", err)
		}
		io.Copy(fo, fi)
		err = fo.Close()
		if err != nil {
			t.Fatalf("TestOlfDssHistoryContent: copyCont failed with error %v", err)
		}
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
		return nil
	}

	tfs, err := testfs.CreateFs("TestOlfDssHistoryContent", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	dss.SetCurrentTime(ttt)
	_ = mkns(dss, "", ttt, []string{"d1/", "f1"})
	_ = mkns(dss, "d1", ttt, []string{"d1a/", "f1b"})
	copyCont(tfs, dss, "a.txt", "f1", ttt)
	copyCont(tfs, dss, "a.txt", "d1/f1b", ttt)

	dss.SetCurrentTime(ttt + 10)
	_ = updns(dss, "d1", ttt, []string{"d1a/", "f1b", "f1c"})
	copyCont(tfs, dss, "d/b.txt", "f1", ttt+10)
	copyCont(tfs, dss, "d/b.txt", "d1/f1b", ttt+10)
	copyCont(tfs, dss, "d/b.txt", "d1/f1c", ttt+10)

	dssbt, err := NewOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path()}, 0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	dssbt.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	mt1, err := dssbt.GetMeta("d1/", true)
	if err != nil || mt1.GetSize() != 13 {
		t.Fatalf("TestOlfDssHistoryContent failed %v %v", mt1, err)
	}
	mt2, err := dssbt.GetMeta("f1", true)
	if err != nil || mt2.GetSize() != 52 {
		t.Fatalf("TestOlfDssHistoryContent failed %v %v", mt2, err)
	}
	mt3, err := dssbt.GetMeta("d1/f1b", true)
	if err != nil || mt3.GetSize() != 52 {
		t.Fatalf("TestOlfDssHistoryContent failed %v %v", mt3, err)
	}
	mt4, err := dssbt.GetMeta("d1/f1c", true)
	if err != nil || mt4.GetSize() != 52 {
		t.Fatalf("TestOlfDssHistoryContent failed %v %v", mt4, err)
	}

	dssbt, err = NewOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path()}, ttt+5, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	dssbt.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	mt1, err = dssbt.GetMeta("d1/", true)
	if err != nil || mt1.GetSize() != 9 {
		t.Fatalf("TestOlfDssHistoryContent failed %v %v", mt1, err)
	}
	mt2, err = dssbt.GetMeta("f1", true)
	if err != nil || mt2.GetSize() != 241 {
		t.Fatalf("TestOlfDssHistoryContent failed %v %v", mt2, err)
	}
	mt3, err = dssbt.GetMeta("d1/f1b", true)
	if err != nil || mt3.GetSize() != 241 {
		t.Fatalf("TestOlfDssHistoryContent failed %v %v", mt3, err)
	}
	if _, err = dssbt.GetMeta("d1/f1c", true); err == nil {
		t.Fatalf("TestOlfDssHistoryContent should fail with error no such entry")
	}
}

func noMoreTestOlfDssHistoryRO(t *testing.T) {
	mkns := func(dss Dss, npath string, tt int64, children []string) error {
		err := dss.Mkns(npath, tt, children, nil)
		if err != nil {
			t.Fatalf("TestOlfDssHistoryRO: mkns failed with error %v", err)
		}
		return nil
	}

	tfs, err := testfs.CreateFs("TestOlfDssHistoryRO", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "l")
	if err != nil {
		t.Fatal(err.Error())
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	// 1641667920 0x0000000061d9dd50
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	_ = mkns(dss, "", ttt, []string{"d1/"})
	_ = mkns(dss, "d1", ttt, []string{"d1a/", "f1b"})

	dssbt, err := NewOlfDss(OlfConfig{DssBaseConfig: DssBaseConfig{LocalPath: tfs.Path()}, Root: tfs.Path()}, ttt-42, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	dssbt.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	_, err = dssbt.Lsns("d1/d1a")
	if err == nil {
		t.Fatalf("TestOlfDssHistoryRO should fail with error no such entry")
	}
	err = dssbt.Mkns("d1/d1a", ttt, []string{}, nil)
	if err == nil {
		t.Fatalf("TestOlfDssHistoryRO should fail with error read-only")
	}
	_, err = dssbt.GetContentWriter("d1/f1b", ttt, nil, nil)
	if err == nil {
		t.Fatalf("TestOlfDssHistoryRO should fail with error read-only")
	}
	err = dssbt.Remove("d1")
	if err == nil {
		t.Fatalf("TestOlfDssHistoryRO should fail with error read-only")
	}
}

func noMoreTestOlfStat(t *testing.T) {

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

	tfs, err := testfs.CreateFs("TestOlfStat", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(tfs.Path(), "s")
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
	ttt := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	err = dss.Mkns("", ttt, []string{"a.txt", "d1/", "d2/"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = dss.Mkns("d1", time.Now().Unix(), nil, []ACLEntry{
		{User: "ub", Rights: Rights{Read: true, Execute: true}},
		{User: "ua", Rights: Rights{Read: true, Write: true, Execute: true}},
	}); err != nil {
		t.Fatal(err)
	}
	m, err := dss.GetMeta("d1/", true)
	if err != nil || len(m.GetAcl()) != 2 || m.GetAcl()[1].User != "ub" {
		t.Fatalf("TestOlfStat err %v meta %v", err, m)
	}

	fo, err := dss.GetContentWriter("a.txt", time.Now().Unix(), []ACLEntry{
		{User: "ud", Rights: Rights{Read: true, Execute: true}},
		{User: "uc", Rights: Rights{Read: true, Write: true, Execute: true}},
	}, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	io.Copy(fo, fi)
	fo.Close()
	m, err = dss.GetMeta("a.txt", true)
	if err != nil || len(m.GetAcl()) != 2 || m.GetAcl()[1].User != "ud" {
		t.Fatalf("TestOlfStat err %v meta %v", err, m)
	}
}

func noMoreTestTmpOlfBase(t *testing.T) {
	optionalSkip(t)
	if os.Getenv("CABRIDSS_KEEP_TEMP") == "" {
		t.Skip(fmt.Sprintf("Skipping %s because you didn't set CABRIDSS_KEEP_TEMP", t.Name()))
	}
	tfs, err := testfs.CreateFs("TestTmpOlfBase", tfsStartup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()

	dss, err := CreateOlfDss(tfs.Path(), "s")
	if err != nil {
		t.Fatal(err.Error())
	}
	dss.SetCurrentTime(int64(0x000000006055b7df))
	runTestBasic(t, tfs.Path(), dss)
}
