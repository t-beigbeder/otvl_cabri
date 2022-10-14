package plumber

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type lsnsPl struct {
	fCount *int64
}

type lsnsMeta struct {
	npath string
	size  int64
	mTime int64
}

func (uiCtx *testUiContext) getLsnsPl() *lsnsPl {
	return uiCtx.Payload.(*lsnsPl)
}

func randChildren(npath string, max int) []string {
	max = rand.Intn(max) + 1
	res := make([]string, max)
	for i := 0; i < max; i++ {
		hot := rand.Intn(2)
		ext := ""
		if hot > 0 {
			ext = "/"
		}
		res[i] = fmt.Sprintf("%x%s", i, ext)
	}
	return res
}

func getMeta(ctx context.Context, path string) lsnsMeta {
	time.Sleep(10 * time.Millisecond)
	return lsnsMeta{npath: path, size: int64(rand.Intn(8192) + 1), mTime: time.Now().Unix() - int64(rand.Intn(8192))}
}

func lsns(ctx context.Context, path string) (chs []string) {
	time.Sleep(10 * time.Millisecond)
	uiCtx := fromContext(ctx)
	chs = randChildren(path, 16)
	curFc := atomic.AddInt64(uiCtx.getLsnsPl().fCount, int64(len(chs)))
	if curFc > 1000 {
		chs = nil
	}
	return
}

func lsnsRecursV1(ctx context.Context, path string) (metas []lsnsMeta) {
	chs := lsns(ctx, path)
	var dchs []string
	for _, ch := range chs {
		if ch[len(ch)-1] == '/' {
			dchs = append(dchs, ch)
		}
	}

	var metas1, metas2 []lsnsMeta
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for _, meta := range Parallelize[string, lsnsMeta](ctx, "", func(ctx context.Context, ch string) lsnsMeta {
			chPath := fmt.Sprintf("%s%s", path, ch)
			return getMeta(ctx, chPath)
		}, chs...) {
			metas1 = append(metas1, meta)
		}
	}()
	go func() {
		defer wg.Done()
		for _, subMetas := range Parallelize[string, []lsnsMeta](ctx, "", func(ctx context.Context, ch string) []lsnsMeta {
			chPath := fmt.Sprintf("%s%s", path, ch)
			return lsnsRecursV1(ctx, chPath)
		}, dchs...) {
			for _, meta := range subMetas {
				metas2 = append(metas2, meta)
			}

		}
	}()
	wg.Wait()
	for _, meta := range metas1 {
		metas = append(metas, meta)
	}
	for _, meta := range metas2 {
		metas = append(metas, meta)
	}

	return
}

func TestLsnsRecursV1(t *testing.T) {
	ctx, cancel := newContext()
	defer cancel()
	ctx = ContextWithConfig(ctx,
		Config{PlizerEnabled: true,
			RglatorsByName: map[string]uint{}})
	uiCtx := fromContext(ctx)
	fCount := int64(0)
	uiCtx.Payload = &lsnsPl{fCount: &fCount}

	metas := lsnsRecursV1(ctx, "/")
	fmt.Printf("metas\n")
	for _, meta := range metas {
		fmt.Printf("%v\n", meta)
	}
}

func lsnsRecursRef(ctx context.Context, path string) (metas []lsnsMeta) {
	chs := lsns(ctx, path)
	var dchs []string
	for _, ch := range chs {
		if ch[len(ch)-1] == '/' {
			dchs = append(dchs, ch)
		}
	}

	var plizedGetMetas Launchable = func(ctx context.Context, iChs interface{}) (iOutput interface{}) {
		chs := Retype[string](Untype[string](iChs.([]string)))
		var metas []lsnsMeta
		for _, meta := range Parallelize[string, lsnsMeta](
			ctx, "",
			func(ctx context.Context, ch string) lsnsMeta {
				chPath := fmt.Sprintf("%s%s", path, ch)
				return getMeta(ctx, chPath)
			},
			chs...) {
			metas = append(metas, meta)
		}
		iOutput = metas
		return
	}

	var plizedLsnsMetas Launchable = func(ctx context.Context, iChs interface{}) (iOutput interface{}) {
		dchs := Retype[string](Untype[string](iChs.([]string)))
		var metas []lsnsMeta
		for _, subMetas := range Parallelize[string, []lsnsMeta](
			ctx, "",
			func(ctx context.Context, ch string) []lsnsMeta {
				chPath := fmt.Sprintf("%s%s", path, ch)
				return lsnsRecursRef(ctx, chPath)
			},
			dchs...) {
			for _, meta := range subMetas {
				metas = append(metas, meta)
			}
		}
		iOutput = metas
		return
	}

	iOutputs := LaunchAndWait(ctx,
		[]string{"", ""},
		[]Launchable{plizedGetMetas, plizedLsnsMetas},
		[]interface{}{chs, dchs},
	)
	outputs := Retype[[]lsnsMeta](iOutputs)
	metas1 := outputs[0]
	for _, meta := range metas1 {
		metas = append(metas, meta)
	}
	metas2 := outputs[1]
	for _, meta := range metas2 {
		metas = append(metas, meta)
	}

	return
}

func TestLsnsRecursRefSeq(t *testing.T) {
	ctx, cancel := newContext()
	defer cancel()
	ctx = ContextWithConfig(ctx,
		Config{PlizerEnabled: false,
			RglatorsByName: map[string]uint{}})
	uiCtx := fromContext(ctx)
	fCount := int64(0)
	uiCtx.Payload = &lsnsPl{fCount: &fCount}

	metas := lsnsRecursRef(ctx, "/")
	fmt.Printf("metas\n")
	for _, meta := range metas {
		fmt.Printf("%v\n", meta)
	}
}

func TestLsnsRecursRefPar(t *testing.T) {
	ctx, cancel := newContext()
	defer cancel()
	ctx = ContextWithConfig(ctx,
		Config{PlizerEnabled: true,
			RglatorsByName: map[string]uint{}})
	uiCtx := fromContext(ctx)
	fCount := int64(0)
	uiCtx.Payload = &lsnsPl{fCount: &fCount}

	metas := lsnsRecursRef(ctx, "/")
	fmt.Printf("metas\n")
	for _, meta := range metas {
		fmt.Printf("%v\n", meta)
	}
}
