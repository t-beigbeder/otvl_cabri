package cabridss

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"testing"
)

func runTest(t *testing.T, ix Index) {
	if _, err, ok := ix.loadMeta("a", 1); err != nil || ok {
		t.Fatal(err)
	}
	if _, err, ok := ix.queryMetaTimes("a"); err != nil || ok {
		t.Fatal(err)
	}
	if err := ix.storeMetaTimes("a", []int64{1, 2}); err != nil {
		t.Fatal(err)
	}
	its, err, ok := ix.queryMetaTimes("a")
	if err != nil || !ok || len(its) != 2 {
		t.Fatal(err)
	}
	if err := ix.storeMetaTimes("a", nil); err != nil {
		t.Fatal(err)
	}
	its, err, ok = ix.queryMetaTimes("a")
	if err != nil || !ok || len(its) != 0 {
		t.Fatal(err)
	}
	if err := ix.storeMeta("a", 1, nil); err != nil {
		t.Fatal(err)
	}
	its, err, ok = ix.queryMetaTimes("a")
	if err != nil || !ok || len(its) != 1 {
		t.Fatal(err)
	}
	if bs, err, ok := ix.loadMeta("a", 1); err != nil || !ok || len(bs) != 0 {
		t.Fatal(err)
	}
	if err := ix.storeMeta("a", 2, []byte("y")); err != nil {
		t.Fatal(err)
	}
	its, err, ok = ix.queryMetaTimes("a")
	if err != nil || !ok || len(its) != 2 {
		t.Fatal(err)
	}
	if bs, err, ok := ix.loadMeta("a", 2); err != nil || !ok || len(bs) != 1 {
		t.Fatal(err)
	}
	if err := ix.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewMIndex(t *testing.T) {
	ix := NewMIndex()
	runTest(t, ix)
}

func TestNewPIndex(t *testing.T) {
	tfs, err := testfs.CreateFs("TestNewPIndex", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	ix, err := NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	runTest(t, ix)

	// persistency
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	its, err, ok := ix.queryMetaTimes("a")
	if err != nil || !ok || len(its) != 2 {
		t.Fatal(err)
	}
	if bs, err, ok := ix.loadMeta("a", 1); err != nil || !ok || len(bs) != 0 {
		t.Fatal(err)
	}
	if bs, err, ok := ix.loadMeta("a", 2); err != nil || !ok || len(bs) != 1 {
		t.Fatal(err)
	}
	if err := ix.Close(); err != nil {
		t.Fatal(err)
	}

	// lock
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err == nil {
		t.Fatalf("should fail with lock error")
	}

	// unlock
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := ix.Close(); err != nil {
		t.Fatal(err)
	}

	// clients
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	clId1 := uuid.New().String()
	if udd, err := ix.recordClient(clId1); err != nil || len(udd.Changed) != 1 {
		t.Fatal(err)
	}
	if _, err = ix.recordClient(clId1); err == nil {
		t.Fatal("should fail with error client registered")
	}
	if ok, err := ix.isClientKnown(clId1); !ok || err != nil {
		t.Fatal(ok, err)
	}
	var udd UpdatedData
	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 0 {
		t.Fatal(err, udd)
	}
	if err := ix.Close(); err != nil {
		t.Fatal(err)
	}
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ix.recordClient(clId1); err == nil {
		t.Fatal("should fail with error client registered")
	}
	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 0 {
		t.Fatal(err, udd)
	}

	clId2 := uuid.New().String()
	if ok, err := ix.isClientKnown(clId2); ok || err != nil {
		t.Fatal(ok, err)
	}
	if udd, err = ix.recordClient(clId2); err != nil || len(udd.Changed) != 1 {
		t.Fatal(err)
	}
	if err := ix.storeMetaTimes("b", []int64{3, 4}); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("b", 3, nil); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("b", 4, []byte("z")); err != nil {
		t.Fatal(err)
	}

	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 1 {
		t.Fatal(err, udd)
	}

	clId3 := uuid.New().String()
	if udd, err = ix.recordClient(clId3); err != nil || len(udd.Changed) != 2 {
		t.Fatal(err)
	}
	if err := ix.storeMeta("c", 4, nil); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("c", 5, []byte("tc")); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("d", 4, nil); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("d", 5, []byte("td")); err != nil {
		t.Fatal(err)
	}

	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 2 {
		t.Fatal(err, udd)
	}
	if udd, err = ix.updateClient(clId2, false); err != nil || len(udd.Changed) != 3 {
		t.Fatal(err, udd)
	}

	if err := ix.Close(); err != nil {
		t.Fatal(err)
	}
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	clId4 := uuid.New().String()
	if udd, err = ix.recordClient(clId4); err != nil || len(udd.Changed) != 4 {
		t.Fatal(err)
	}
	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 0 {
		t.Fatal(err, udd)
	}
	if udd, err = ix.updateClient(clId2, true); err != nil || len(udd.Changed) != 4 {
		t.Fatal(err, udd)
	}
	if udd, err = ix.updateClient(clId3, false); err != nil || len(udd.Changed) != 2 {
		t.Fatal(err, udd)
	}

	if err := ix.Close(); err != nil {
		t.Fatal(err)
	}
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 0 {
		t.Fatal(err, udd)
	}
	if udd, err = ix.updateClient(clId2, false); err != nil || len(udd.Changed) != 0 {
		t.Fatal(err, udd)
	}
	if udd, err = ix.updateClient(clId3, false); err != nil || len(udd.Changed) != 0 {
		t.Fatal(err, udd)
	}
}

func TestNewPCIndex(t *testing.T) {
	tfs, err := testfs.CreateFs("TestNewPCIndex", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tfs.Delete()
	ix, err := NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	runTest(t, ix)

	// clients indexes
	ix, err = NewPIndex(ufpath.Join(tfs.Path(), "pindex.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	clId1 := uuid.New().String()
	udd, err := ix.recordClient(clId1)
	if err != nil || len(udd.Changed) != 1 {
		t.Fatal(err)
	}
	ixCl1, err := NewPIndex(ufpath.Join(tfs.Path(), "pindexCl1.dat"), false, false)
	if err != nil {
		if err := ix.Close(); err != nil {
			t.Fatal(err)
		}

		t.Fatal(err)
	}
	if err = ixCl1.updateData(udd, false); err != nil {
		t.Fatal(err)
	}

	clId2 := uuid.New().String()
	if udd, err = ix.recordClient(clId2); err != nil || len(udd.Changed) != 1 {
		t.Fatal(err)
	}
	ixCl2, err := NewPIndex(ufpath.Join(tfs.Path(), "pindexCl2.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = ixCl2.updateData(udd, false); err != nil {
		t.Fatal(err)
	}

	if err := ix.storeMetaTimes("b", []int64{3, 4}); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("b", 3, nil); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("b", 4, []byte("z")); err != nil {
		t.Fatal(err)
	}

	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 1 {
		t.Fatal(err, udd)
	}
	if err = ixCl1.updateData(udd, false); err != nil {
		t.Fatal(err)
	}

	clId3 := uuid.New().String()
	if udd, err = ix.recordClient(clId3); err != nil || len(udd.Changed) != 2 {
		t.Fatal(err)
	}
	ixCl3, err := NewPIndex(ufpath.Join(tfs.Path(), "pindexCl3.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = ixCl3.updateData(udd, false); err != nil {
		t.Fatal(err)
	}

	if err := ix.storeMeta("c", 4, nil); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("c", 5, []byte("tc")); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("d", 4, nil); err != nil {
		t.Fatal(err)
	}
	if err := ix.storeMeta("d", 5, []byte("td")); err != nil {
		t.Fatal(err)
	}

	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 2 {
		t.Fatal(err, udd)
	}
	if err = ixCl1.updateData(udd, false); err != nil {
		t.Fatal(err)
	}
	if udd, err = ix.updateClient(clId2, false); err != nil || len(udd.Changed) != 3 {
		t.Fatal(err, udd)
	}
	if err = ixCl2.updateData(udd, false); err != nil {
		t.Fatal(err)
	}

	clId4 := uuid.New().String()
	if udd, err = ix.recordClient(clId4); err != nil || len(udd.Changed) != 4 {
		t.Fatal(err)
	}
	ixCl4, err := NewPIndex(ufpath.Join(tfs.Path(), "pindexCl4.dat"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = ixCl4.updateData(udd, false); err != nil {
		t.Fatal(err)
	}

	if udd, err = ix.updateClient(clId1, false); err != nil || len(udd.Changed) != 0 {
		t.Fatal(err, udd)
	}
	if udd, err = ix.updateClient(clId2, true); err != nil || len(udd.Changed) != 4 {
		t.Fatal(err, udd)
	}
	if err = ixCl2.updateData(udd, false); err != nil {
		t.Fatal(err)
	}
	if udd, err = ix.updateClient(clId3, false); err != nil || len(udd.Changed) != 2 {
		t.Fatal(err, udd)
	}
	if err = ixCl3.updateData(udd, false); err != nil {
		t.Fatal(err)
	}

	for i, ixCl := range []Index{ixCl1, ixCl2, ixCl3, ixCl4} {
		its, err, ok := ixCl.queryMetaTimes("a")
		if err != nil || !ok || len(its) != 2 {
			t.Fatal(err)
		}
		if bs, err, ok := ix.loadMeta("a", 1); err != nil || !ok || len(bs) != 0 {
			t.Fatal(err)
		}
		if bs, err, ok := ix.loadMeta("a", 2); err != nil || !ok || len(bs) != 1 {
			t.Fatal(err)
		}

		its, err, ok = ixCl.queryMetaTimes("b")
		if err != nil || !ok || len(its) != 2 {
			t.Fatal(err)
		}
		if bs, err, ok := ix.loadMeta("b", 3); err != nil || !ok || len(bs) != 0 {
			t.Fatal(err)
		}
		if bs, err, ok := ix.loadMeta("b", 4); err != nil || !ok || len(bs) != 1 {
			t.Fatal(err)
		}

		its, err, ok = ixCl.queryMetaTimes("c")
		if err != nil || !ok || len(its) != 2 {
			t.Fatal(i, its, err, ok)
		}
		if bs, err, ok := ix.loadMeta("c", 4); err != nil || !ok || len(bs) != 0 {
			t.Fatal(err)
		}
		if bs, err, ok := ix.loadMeta("c", 5); err != nil || !ok || len(bs) != 2 {
			t.Fatal(err)
		}
	}
}

func TestMIndexClient(t *testing.T) {
	ix := NewMIndex()
	clId1 := uuid.New().String()
	udd, err := ix.recordClient(clId1)
	if err != nil || len(udd.Changed) != 0 {
		t.Fatal(err)
	}
	ok, err := ix.isClientKnown(clId1)
	if !ok || err != nil {
		t.Error(ok, err)
	}
	clId2 := uuid.New().String()
	ok, err = ix.isClientKnown(clId2)
	if ok || err != nil {
		t.Error(ok, err)
	}
	udd, err = ix.updateClient(clId1, false)
	if err != nil || len(udd.Changed) != 0 {
		t.Fatal(err)
	}
}

func TestPIndexRepair(t *testing.T) {
	if os.Getenv("CABRIDSS_KEEP_DEV_TESTS") == "" {
		t.Skip(fmt.Sprintf("Skipping %s because you didn't set CABRIDSS_KEEP_DEV_TESTS", t.Name()))
	}
	ix, err := NewPIndex("/home/guest/Documents/tmp/index.bdb", true, false)
	if err != nil {
		t.Fatal(err)
	}
	defer ix.Close()
	os.WriteFile("/tmp/ixDump.txt", []byte((ix.Dump())), 0666)
	ds, err := ix.Repair(false)
	if err != nil {
		t.Error(err)
	}
	_ = ds
}
