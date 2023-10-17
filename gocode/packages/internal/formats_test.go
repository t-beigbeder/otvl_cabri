package internal_test

import (
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func make1024cs() [1024][32]byte {
	var css [1024][32]byte
	for v := 0; v < 1024; v++ {
		css[v] = sha256.Sum256([]byte(fmt.Sprintf("%d", v)))
	}
	return css
}

func TestSha256ToStr64(t *testing.T) {
	for _, cs := range make1024cs() {
		s := internal.Sha256ToStr64(cs[:])
		if len(s) != 64 {
			t.Fatalf("s is %s, its length is %d", s, len(s))
		}
	}
}

func TestSha256ToStr32(t *testing.T) {
	for _, cs := range make1024cs() {
		s2 := internal.Sha256ToStr64(cs[:])
		s1 := internal.Sha256ToStr32(cs[:])
		if len(s1) != 32 {
			t.Fatalf("s is %s, its length is %d", s1, len(s1))
		}
		if !strings.HasPrefix(s2, s1) {
			t.Fatalf("s2 is %s, s1 is %s", s2, s1)
		}
	}
}

func verifStr32ForCs(t *testing.T, s string, cs [32]byte) {
	csr, err := internal.Str32ToSha256Trunc(s)
	if err != nil {
		t.Fatalf("Str32ToSha256Trunc failed on %s (%v)", s, err)
	}
	if !reflect.DeepEqual(csr[:], cs[0:16]) {
		t.Fatalf("returned cs %x differs from original %x", csr, cs)
	}

}

func TestSha256ToPath8x3(t *testing.T) {
	for _, cs := range make1024cs() {
		ss, sp2 := internal.Sha256ToStr32Path(cs[:], "l")
		if ss != internal.Sha256ToStr32(cs[:]) {
			t.Fatalf("cs is %v ss is %s", cs, ss)
		}
		sp := internal.Sha256ToPath(cs[:], "l")
		if sp != sp2 {
			t.Fatalf("sp is %s sp2 is %s", sp, sp2)
		}
		if len(sp) != 34 {
			t.Fatalf("sp is %s, its length is %d", sp, len(sp))
		}
		sas := strings.Split(sp, "/")
		if len(sas) != 3 {
			t.Fatalf("sp %s cannot be parsed (%v)", sp, sas)
		}
		verifStr32ForCs(t, fmt.Sprintf("%s%s%s", sas[0], sas[1], sas[2]), cs)
	}
}

func TestStr32ToSha256Trunc(t *testing.T) {
	for ix, cs := range make1024cs() {
		if ix == 0 {
			s := internal.Sha256ToStr32(cs[:])
			s = s[0:5] + "z" + s[6:32]
			csr, err := internal.Str32ToSha256Trunc(s)
			if err == nil {
				t.Fatalf("Str32ToSha256Trunc should fail on %s (%v)", s, csr)
			}
		}
		if ix == 1 {
			s := internal.Sha256ToStr32(cs[:])
			s = s[0:6] + "zz" + s[8:32]
			csr, err := internal.Str32ToSha256Trunc(s)
			if err == nil {
				t.Fatalf("Str32ToSha256Trunc should fail on %s (%v)", s, csr)
			}
		}
		s := internal.Sha256ToStr32(cs[:])
		csr, err := internal.Str32ToSha256Trunc(s)
		if err != nil {
			t.Fatalf("Str32ToSha256Trunc failed on %s (%v)", s, err)
		}
		if !reflect.DeepEqual(csr[:], cs[0:16]) {
			t.Fatalf("returned cs %x differs from original %x", csr, cs)
		}
	}
}

func TestStr16ToInt64(t *testing.T) {
	decodeInt := func(i int64) time.Time {
		return time.Unix(i/1e9, i%1e9)
	}
	display := func(i int64) {
		sec, nano := internal.Nano2SecNano(i)
		tm := time.Unix(sec, nano)
		unano := tm.UnixNano()
		s16 := internal.TimeToStr16(tm)
		i64, _ := internal.Str16ToInt64(s16)
		dsp := fmt.Sprintf("i %d %x sec %d %x nano %d %x tm %v unano %d %x s16 %s i64 %d %x\n", i, i, sec, sec, nano, nano, tm, unano, unano, s16, i64, i64)
		fmt.Fprint(io.Discard, dsp)
		fmt.Fprint(os.Stdout, dsp)
	}
	display(0)
	display(-1)
	display(1)
	display(time.Now().UnixNano())
	rand.Seed(42)
	rs := make(map[int64]string)
	now := time.Now()
	rs[now.UnixNano()] = internal.TimeToStr16(now)
	fbd := time.Date(1918, time.April, 24, 23, 0, 0, 0, time.UTC)
	rs[fbd.UnixNano()] = internal.TimeToStr16(fbd)
	rs[-1] = internal.TimeToStr16(time.Unix(0, -1))
	dbd := time.Date(2125, time.May, 20, 23, 59, 0, 0, time.UTC)
	rs[dbd.UnixNano()] = internal.TimeToStr16(dbd)
	for v := 0; v < 16384; v++ {
		i := rand.Int63()
		rs[i] = internal.TimeToStr16(time.Unix(i/1e9, i%1e9))
		if len(rs[i]) != 16 {
			t.Fatalf("TimeToStr16 res is %s", rs[i])
		}
	}
	for i1 := range rs {
		i2, err := internal.Str16ToInt64(rs[i1])
		if err != nil || i2 != i1 {
			t1 := decodeInt(i1)
			t2 := decodeInt(i2)
			t.Fatalf("TestStr16ToInt64 failed err %v i2 %d i1 %d rs %s t1 %v t2 %v", err, i2, i1, rs[i1], t1, t2)
		}
	}
	_, err := internal.Str16ToInt64("ffff")
	if err == nil {
		t.Fatalf("TestStr16ToInt64 should fail with error invalid hex")
	}
	_, err = internal.Str16ToInt64("fffffffffffffffz")
	if err == nil {
		t.Fatalf("TestStr16ToInt64 should fail with error invalid hex")
	}
	_, err = internal.Str16ToInt64("fffffffffffffffF")
	if err == nil {
		t.Fatalf("TestStr16ToInt64 should fail with error invalid hex")
	}
}

func TestCoverPanics1(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestCoverPanics: Sha256ToStr64 did not panic")
		}

	}()
	internal.Sha256ToStr64([]byte{'0'})
}

func TestCoverPanics2(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestCoverPanics: Sha256ToStr32 did not panic")
		}
	}()
	internal.Sha256ToStr32([]byte{'0'})
}

func TestCoverPanics3(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestCoverPanics: Sha256ToPathx3 did not panic")
		}
	}()
	internal.Sha256ToPath([]byte{'0'}, "l")
}

func TestSizes(t *testing.T) {
	cs := sha256.Sum256([]byte("42"))
	ps := internal.Sha256ToPath(cs[:], "s")
	pm := internal.Sha256ToPath(cs[:], "m")
	pl := internal.Sha256ToPath(cs[:], "l")
	if ps != "73/475cb40a568e8da8a045ced110137e" || pm != "734/75cb40a568e8da8a045ced110137e" || pl != "734/75c/b40a568e8da8a045ced110137e" {
		t.Fatalf("ps %s pm %s pl %s", ps, pm, pl)
	}
}

func TestPath2Str32(t *testing.T) {
	for _, sz := range []string{"s", "m", "l"} {
		id, _ := uuid.NewUUID()
		sid1 := internal.NameToHashStr32(id.String())
		p := internal.Str32ToPath(sid1, sz)
		sid2, _ := internal.Path2Str32(p, sz)
		if sid2 != sid1 {
			t.Fatalf("%s %s != %s", sz, sid2, sid1)
		}
	}

}
