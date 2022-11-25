package cabridss

import (
	"bytes"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"testing"
	"time"
)

func tfsStartup(tfs *testfs.Fs) error {
	if err := tfs.RandTextFile("a.txt", 41); err != nil {
		return err
	}
	if err := os.Mkdir(ufpath.Join(tfs.Path(), "d"), 0755); err != nil {
		return err
	}
	if err := tfs.RandTextFile("d/b.txt", 20); err != nil {
		return err
	}
	if err := os.Mkdir(ufpath.Join(tfs.Path(), ".cabri"), 0755); err != nil {
		return err
	}
	return nil
}

var ucpCount = 0
var ids []IdentityConfig
var currentUserConfig UserConfig
var mtCount = time.Date(2018, time.April, 24, 23, 0, 0, 0, time.UTC).Unix() - 1

func mtimeCount() int64 { mtCount += 1; return mtCount }

func newUcp(tfs *testfs.Fs) (ucp string, uc UserConfig, err error) {
	ucpCount += 1
	ucp = ufpath.Join(tfs.Path(), fmt.Sprintf(".cabri-i%d", ucpCount))
	if ucpCount == 1 {
		uc1, err1 := GetUserConfig(DssBaseConfig{}, ucp)
		if err1 != nil {
			return
		}
		ids = uc1.Identities
	}
	id, err := GenIdentity(fmt.Sprintf("id-%d", ucpCount))
	ids = append(ids, id)
	for _, id = range ids {
		UserConfigPutIdentity(DssBaseConfig{}, ucp, id)
	}
	uc, _ = GetUserConfig(DssBaseConfig{}, ucp)
	return
}

func ucpPkeys() []string {
	return IdPkeys(currentUserConfig)
}

func ucpSecrets() []string {
	return IdSecrets(currentUserConfig)
}

func aclFromUcp() (acl []ACLEntry) {
	for _, id := range currentUserConfig.Identities {
		acl = append(acl, ACLEntry{User: id.PKey, Rights: Rights{Write: true}})
	}
	return
}

func serverIndex(dss HDss) (Index, Index) {
	px, ok := dss.(*ODss).proxy.(*webDssImpl)
	if !ok {
		px, ok := dss.(*ODss).proxy.(*eDssImpl)
		if !ok {
			return nil, nil
		}
		if !px.libApi {
			return nil, nil
		}
		return px.apc.GetConfig().(webDssClientConfig).libDss.GetIndex(), dss.GetIndex()
	}
	if !px.libApi {
		return nil, nil
	}
	return px.apc.GetConfig().(webDssClientConfig).libDss.GetIndex(), dss.GetIndex()
}

func dumpIx(six, cix Index) {
	println("six")
	println(six.Dump())
	println("cix")
	println(cix.Dump())
	println()
}

func runTestBasic(t *testing.T, createDssCb func(*testfs.Fs) error, newDssCb func(*testfs.Fs) (HDss, error)) error {
	optionalSkip(t)
	tfs, err := testfs.CreateFs(t.Name(), tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	if err = createDssCb(tfs); err != nil {
		return err
	}
	dss, err := newDssCb(tfs)
	if err != nil {
		t.Fatal(err)
	}
	defer dss.Close()
	six, cix := serverIndex(dss)
	_, _ = six, cix

	if err := dss.Mkns("", mtimeCount(), []string{"d1é/", "d2/"}, nil); err != nil {
		return err
	}
	// check client index if applicable
	if cix != nil {
		dumpIx(six, cix)
		ts, err, ok := cix.queryMetaTimes("")
		if err != nil || len(ts) == 0 || !ok {
			return fmt.Errorf("ts %v err %v ok %v", ts, err, ok)
		}
	}

	cs, err := dss.Lsns("")
	if err != nil {
		return err
	}
	if cs[0] != "d1é/" {
		return fmt.Errorf("%v %v", cs, err)
	}
	if err := dss.Mkns("d1é", mtimeCount(), []string{"a.txt"}, nil); err != nil {
		return err
	}
	if err := dss.Mkns("d2", mtimeCount(), []string{"a.txt"}, nil); err != nil {
		return err
	}
	fi, err := os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		return err
	}
	defer fi.Close()

	fo, err := dss.GetContentWriter("d1é/a.txt", mtimeCount(), nil, func(err error, size int64, ch string) {
		if err != nil {
			t.Log(err)
		}
		if dss.IsEncrypted() {
			return
		}
		if size != 241 {
			t.Logf("size %d != 241", size)
		}
		if ch != "484f617a695613aac4b346237aa01548" {
			t.Logf("%s != %s", ch, "484f617a695613aac4b346237aa01548")
		}
	})
	if err != nil {
		t.Log(err)
		return err
	}
	if _, err = io.Copy(fo, fi); err != nil {
		t.Log(err)
		return err
	}
	if err = fo.Close(); err != nil {
		t.Log(err)
		return err
	}

	fi.Close()
	fi, err = os.Open(ufpath.Join(tfs.Path(), "a.txt"))
	if err != nil {
		return err
	}
	defer fi.Close()

	fo, err = dss.GetContentWriter("d2/a.txt", mtimeCount(), nil, nil)
	if err != nil {
		t.Log(err)
		return err
	}
	if _, err = io.Copy(fo, fi); err != nil {
		t.Log(err)
		return err
	}
	if err = fo.Close(); err != nil {
		t.Log(err)
		return err
	}

	rc, err := dss.GetContentReader("d1é/a.txt")
	if err != nil {
		return err
	}
	defer rc.Close()
	var buf bytes.Buffer
	io.Copy(&buf, rc)
	if buf.Len() != 241 {
		return fmt.Errorf("%d != 241", buf.Len())
	}

	meta, err := dss.GetMeta("d1é/a.txt", true)
	if err != nil {
		return err
	}
	if meta.GetSize() != 241 || meta.GetCh() != "484f617a695613aac4b346237aa01548" {
		return fmt.Errorf("meta %v", meta)
	}
	isDup, err := dss.IsDuplicate(meta.GetCh())
	if err != nil || (!dss.IsEncrypted() && !isDup) {
		return fmt.Errorf("%v %v", err, isDup)
	}

	meta, err = dss.GetMeta("d1é/", true)
	if err != nil {
		return err
	}
	if meta.GetSize() != 6 || meta.GetCh() != "10fbdce5d5e2ba7e0249a4a8921faede" {
		return fmt.Errorf("meta %v", meta)
	}

	if err = dss.Remove("d1é/a.txt"); err != nil {
		return err
	}
	meta, err = dss.GetMeta("d1é/", true)
	if err != nil {
		return err
	}
	if meta.GetSize() != 0 || meta.GetCh() != "e3b0c44298fc1c149afbf4c8996fb924" {
		return fmt.Errorf("meta %v", meta)
	}
	// check client index if applicable
	if cix != nil {
		dumpIx(six, cix)
		ts, err, ok := dss.GetIndex().queryMetaTimes("d2/a.txt")
		if err != nil || len(ts) == 0 || !ok {
			return fmt.Errorf("ts %v err %v ok %v", ts, err, ok)
		}
	}

	// check index persistency
	if err = dss.Close(); err != nil {
		return err
	}
	dss, err = newDssCb(tfs)
	if err != nil {
		return err
	}
	defer dss.Close()
	meta, err = dss.GetMeta("d2/a.txt", true)
	if err != nil {
		return err
	}
	six, cix = serverIndex(dss)
	if cix != nil {
		dumpIx(six, cix)
		ts, err, ok := dss.GetIndex().queryMetaTimes("d2/a.txt")
		if err != nil || len(ts) == 0 || !ok {
			return fmt.Errorf("ts %v err %v ok %v", ts, err, ok)
		}
	}
	return nil
}

func dssMkTestFile(dss Dss, src string, dest string) error {
	fi, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fi.Close()

	fo, err := dss.GetContentWriter(dest, time.Now().Unix(), nil, nil)
	if err != nil {
		return err
	}
	io.Copy(fo, fi)
	return fo.Close()
}

func prepareTestHistory(t *testing.T, tfsPath string, dss HDss) error {
	optionalSkip(t)
	ttr := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	fan := ufpath.Join(tfsPath, "a.txt")
	fbn := ufpath.Join(tfsPath, "d/b.txt")

	dss.SetCurrentTime(ttr + 27*3600)
	if err := dss.Mkns("", 0, []string{"d1/", "f"}, nil); err != nil {
		return err
	}
	dssMkTestFile(dss, fan, "f")
	if err := dss.Mkns("d1", 0, []string{"f1a", "f1b"}, nil); err != nil {
		return err
	}
	dssMkTestFile(dss, fan, "d1/f1a")
	dssMkTestFile(dss, fan, "d1/f1b")

	dss.SetCurrentTime(ttr)
	if err := dss.Updatens("", 0, []string{}, nil); err != nil {
		return err
	}

	dss.SetCurrentTime(ttr + 49*3600)
	if err := dss.Updatens("", 0, []string{"d2/", "f"}, nil); err != nil {
		return err
	}
	if err := dss.Mkns("d2", 0, []string{"d2a/", "d2b/"}, nil); err != nil {
		return err
	}
	if err := dss.Mkns("d2/d2a", 0, []string{"f22a"}, nil); err != nil {
		return err
	}
	dssMkTestFile(dss, fan, "d2/d2a/f22a")
	if err := dss.Mkns("d2/d2b", 0, nil, nil); err != nil {
		return err
	}

	dss.SetCurrentTime(ttr + 75*3600)
	dss.Remove("d2/d2a/f22a")

	dss.SetCurrentTime(ttr + 95*3600)
	if err := dss.Updatens("d2/d2a", 0, []string{"f22a"}, nil); err != nil {
		return err
	}
	dssMkTestFile(dss, fbn, "d2/d2a/f22a")

	dss.SetCurrentTime(ttr + 121*3600)
	if err := dss.Updatens("", 0, []string{"d2/", "f", "d1/"}, nil); err != nil {
		return err
	}
	mHes, err := dss.GetHistory("", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes[""]})
	mHes, err = dss.GetHistory("d1/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d1/"]})
	mHes, err = dss.GetHistory("d2/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d2/"]})
	return nil
}

func runTestHistory(t *testing.T, createDssCb func(*testfs.Fs) error, newDssCb func(*testfs.Fs) (HDss, error)) error {
	optionalSkip(t)
	ttr := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()

	tfs, err := testfs.CreateFs(t.Name(), tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	if err = createDssCb(tfs); err != nil {
		return err
	}
	dss, err := newDssCb(tfs)
	if err != nil {
		t.Fatal(err)
	}
	defer dss.Close()

	if err := prepareTestHistory(t, tfs.Path(), dss); err != nil {
		return err
	}

	mHes, err := dss.GetHistory("", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes[""]})
	mHes, err = dss.GetHistory("d1/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d1/"]})
	mHes, err = dss.GetHistory("d2/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d2/"]})

	mHes, err = dss.GetHistory("d2/d2a/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d2/d2a/"]})

	mHes, err = dss.GetHistory("d2/d2a/f22a", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d2/d2a/f22a"]})

	mHes, err = dss.GetHistory("d2/", true)
	if err != nil {
		return err
	}
	fmt.Printf("\nGH d2/\n%s\n", internal.MapSliceStringer[HistoryInfo]{mHes})

	mHes, err = dss.GetHistory("", true)
	if err != nil {
		return err
	}
	fmt.Printf("\nGH ROOT\n%s\n", internal.MapSliceStringer[HistoryInfo]{mHes})

	ftrh := func(npath string, recursive bool, evaluate bool, startH int64, endH int64) error {
		start, end := startH, endH
		if startH != 0 && startH < 1000 {
			start = ttr + startH*3600
		}
		if endH != 0 && endH < 1000 {
			end = ttr + endH*3600
		}
		mHes, err = dss.RemoveHistory(npath, recursive, evaluate, start, end)
		if err != nil {
			return err
		}
		fmt.Printf("\nRH \"%s\" %s-%s\n%s\n", npath, UnixUTC(start), UnixUTC(end), internal.MapSliceStringer[HistoryInfo]{mHes})
		return nil
	}

	for _, np := range []string{"f", "d2/", ""} {
		if err = ftrh(np, true, true, 30, 40); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 40, 60); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 60, 90); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 90, 100); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 100, 130); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 40, 90); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 60, 100); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 40, 100); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 100, 0); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 0, 40); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 0, 0); err != nil {
			return err
		}
	}

	sti, perr := dss.ScanStorage()
	if perr != nil {
		return perr
	}
	_ = sti

	mai, err := dss.AuditIndex()
	if err != nil {
		return err
	}
	fmt.Printf("\nAudit info\n%s\n", internal.MapSliceStringer[AuditIndexInfo]{mai})
	return nil
}

func runTestMultiHistory(t *testing.T, createDssCb func(*testfs.Fs) error, newDssCb func(*testfs.Fs) (HDss, error)) error {
	optionalSkip(t)
	tfs, err := testfs.CreateFs(t.Name(), tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	if err = createDssCb(tfs); err != nil {
		return err
	}
	dss, err := newDssCb(tfs)
	if err != nil {
		t.Fatal(err)
	}
	defer dss.Close()
	ttr := time.Date(2022, time.January, 8, 18, 52, 0, 0, time.UTC).Unix()
	if err := prepareTestHistory(t, tfs.Path(), dss); err != nil {
		return err
	}

	mHes, err := dss.GetHistory("", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes[""]})
	mHes, err = dss.GetHistory("d1/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d1/"]})
	mHes, err = dss.GetHistory("d2/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d2/"]})

	mHes, err = dss.GetHistory("d2/d2a/", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d2/d2a/"]})

	mHes, err = dss.GetHistory("d2/d2a/f22a", false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", internal.SliceStringer[HistoryInfo]{mHes["d2/d2a/f22a"]})

	mHes, err = dss.GetHistory("d2/", true)
	if err != nil {
		return err
	}
	fmt.Printf("\nGH d2/\n%s\n", internal.MapSliceStringer[HistoryInfo]{mHes})

	mHes, err = dss.GetHistory("", true)
	if err != nil {
		return err
	}
	fmt.Printf("\nGH ROOT\n%s\n", internal.MapSliceStringer[HistoryInfo]{mHes})

	ftrh := func(npath string, recursive bool, evaluate bool, startH int64, endH int64) error {
		start, end := startH, endH
		if startH != 0 && startH < 1000 {
			start = ttr + startH*3600
		}
		if endH != 0 && endH < 1000 {
			end = ttr + endH*3600
		}
		mHes, err = dss.RemoveHistory(npath, recursive, evaluate, start, end)
		if err != nil {
			return err
		}
		fmt.Printf("\nRH \"%s\" %s-%s\n%s\n", npath, UnixUTC(start), UnixUTC(end), internal.MapSliceStringer[HistoryInfo]{mHes})
		return nil
	}

	for _, np := range []string{"f", "d2/", ""} {
		if err = ftrh(np, true, true, 30, 40); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 40, 60); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 60, 90); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 90, 100); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 100, 130); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 40, 90); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 60, 100); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 40, 100); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 100, 0); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 0, 40); err != nil {
			return err
		}
		if err = ftrh(np, true, true, 0, 0); err != nil {
			return err
		}
	}

	sti, perr := dss.ScanStorage()
	if perr != nil {
		return perr
	}
	_ = sti

	mai, err := dss.AuditIndex()
	if err != nil {
		return err
	}
	fmt.Printf("\nAudit info\n%s\n", internal.MapSliceStringer[AuditIndexInfo]{mai})
	return nil
}
