package cabridss

import (
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/mockfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"testing"
	"time"
)

func TestOlfACLBase(t *testing.T) {

	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		return nil
	}

	copy := func(tfs *testfs.Fs, dss Dss, target string, acl []ACLEntry) error {
		fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
		if err != nil {
			return err
		}
		defer fi.Close()
		fo, err := dss.GetContentWriter(target, time.Now().Unix(), acl, nil)
		if err != nil {
			return err
		}
		if _, err = io.Copy(fo, fi); err != nil {
			return err
		}
		return fo.Close()
	}

	tfs, err := testfs.CreateFs("TestOlfACLBase", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	dss, err := CreateOlfDss(OlfConfig{
		DssBaseConfig: DssBaseConfig{
			ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
			LocalPath: tfs.Path()}, Root: tfs.Path(), Size: "s"})
	if err != nil {
		t.Fatal(err.Error())
	}
	var aes13 = []ACLEntry{
		{User: "ua", Rights: Rights{Read: true, Write: true}},
		{User: "uc", Rights: Rights{Read: true, Write: true}},
	}
	var aes2 = []ACLEntry{
		{User: "ub", Rights: Rights{Read: true, Write: true}},
	}
	var aes3 = []ACLEntry{
		{User: "uc", Rights: Rights{Read: true, Write: true}},
	}
	dss.Mkns("", time.Now().Unix(), nil, aes13)
	if dss, err = NewOlfDss(OlfConfig{
		DssBaseConfig: DssBaseConfig{
			ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
			LocalPath: tfs.Path()}, Root: tfs.Path()}, 0, []string{"ua", "ub"}); err != nil {
		t.Fatal(err)
	}
	cbs := mockfs.MockCbs{}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	dss.Updatens("", time.Now().Unix(), []string{"d13/", "d2/", "d3/"}, aes13)
	if err = dss.Mkns("d13", time.Now().Unix(), []string{"f13a.txt"}, aes13); err != nil {
		t.Fatal(err)
	}
	meta, err := dss.GetMeta("", true)
	if err != nil {
		t.Fatal(err)
	}
	if meta.GetAcl()[0].GetUser() != "ua" || !meta.GetAcl()[1].GetRights().Read {
		t.Fatal(err)
	}

	if err = dss.Updatens("d13", time.Now().Unix(), []string{"f13b.txt", "f13only3.txt"}, aes13); err != nil {
		t.Fatal(err)
	}
	if err = copy(tfs, dss, "d13/f13b.txt", aes13); err != nil {
		t.Fatal(err)
	}
	if err = copy(tfs, dss, "d13/f13only3.txt", aes3); err != nil {
		t.Fatal(err)
	}
	if err = dss.Mkns("d2", time.Now().Unix(), []string{"f2a.txt"}, aes2); err != nil {
		t.Fatal(err)
	}
	if err = dss.Updatens("d2", time.Now().Unix(), []string{"f2b.txt"}, aes2); err != nil {
		t.Fatal(err)
	}
	if err = copy(tfs, dss, "d2/f2b.txt", aes2); err != nil {
		t.Fatal(err)
	}
	if err = dss.Mkns("d3", time.Now().Unix(), []string{"f3.txt"}, aes3); err != nil {
		t.Fatal(err)
	}
	if err = copy(tfs, dss, "d3/f3.txt", aes13); err == nil {
		t.Fatalf("GetContentWriter should fail with permission denied error")
	}
	if err = dss.Remove("d3/f3.txt"); err == nil {
		t.Fatalf("Remove should fail with permission denied error")
	}
	if err = dss.Updatens("d3", time.Now().Unix(), []string{"f3err.txt"}, aes3); err == nil {
		t.Fatalf("Updatens should fail with permission denied error")
	}
	if err = dss.Remove("d3/"); err == nil {
		t.Fatalf("Remove should fail with permission denied error")
	}
	if err = copy(tfs, dss, "d13/f13only3.txt", aes3); err == nil {
		t.Fatalf("GetContentWriter should fail with permission denied error")
	}
	if _, err = dss.GetMeta("d13/f13only3.txt", true); err == nil {
		t.Fatalf("GetMeta should fail with permission denied error")
	}
	if _, err = dss.GetContentReader("d13/f13only3.txt"); err == nil {
		t.Fatalf("GetContentReader should fail with permission denied error")
	}

	if dss, err = NewOlfDss(OlfConfig{
		DssBaseConfig: DssBaseConfig{
			ConfigDir: ufpath.Join(tfs.Path(), ".cabri"),
			LocalPath: tfs.Path()}, Root: tfs.Path()}, 0, []string{"uc"}); err != nil {
		t.Fatal(err)
	}
	dss.SetAfs(mockfs.New(afero.NewOsFs(), &cbs))
	dss.Updatens("", time.Now().Unix(), []string{"d13/", "d2/", "d3/"}, aes13)
	if err = dss.Updatens("d13", time.Now().Unix(), []string{"f13c.txt"}, aes13); err != nil {
		t.Fatal(err)
	}
	if _, err = dss.Lsns("d2"); err == nil {
		t.Fatalf("Lsns should fail with permission denied error")
	}
	if _, err = dss.GetMeta("d2/", true); err == nil {
		t.Fatalf("GetMeta should fail with permission denied error")
	}
	if err = dss.Updatens("d2", time.Now().Unix(), []string{"f2err.txt"}, aes2); err == nil {
		t.Fatalf("Updatens should fail with permission denied error")
	}
	if err = dss.Updatens("d3", time.Now().Unix(), []string{"f3c.txt"}, aes3); err != nil {
		t.Fatal(err)
	}
}

func TestNotUx(t *testing.T) {
	startup := func(tfs *testfs.Fs) error {
		if err := tfs.RandTextFile("a.txt", 41); err != nil {
			return err
		}
		return nil
	}

	tfs, err := testfs.CreateFs("TestNotUx", startup)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	pa := ufpath.Join(tfs.Path(), "a.txt")
	fi, err := os.Stat(pa)
	if err != nil {
		t.Fatal(err)
	}
	acl := getSysAclNotUx(fi)
	if len(acl) != 1 || !acl[0].Rights.Write {
		t.Fatalf("acl %+v", acl)
	}
	acl[0].Rights.Write = false
	if err = setSysAclNotUx(pa, acl); err != nil {
		t.Fatal(err)
	}
	if fi, err = os.Stat(pa); err != nil {
		t.Fatal(err)
	}
	acl = getSysAclNotUx(fi)
	if len(acl) != 1 || acl[0].Rights.Write {
		t.Fatalf("acl %+v", acl)
	}
}
