package internal

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

func Sha256ToStr64(cs []byte) string {
	if len(cs) != sha256.Size {
		panic(fmt.Sprintf("len %d != %d", len(cs), sha256.Size))
	}
	return fmt.Sprintf("%x", cs)
}

func Sha256ToStr32(cs []byte) string {
	if len(cs) != sha256.Size {
		panic(fmt.Sprintf("len %d != %d", len(cs), sha256.Size))
	}
	return fmt.Sprintf("%x", cs[0:sha256.Size/2])
}

func NameToHashStr32(name string) string {
	ht := sha256.Sum256([]byte(name))
	return Sha256ToStr32(ht[:])
}

func Str32ToSha256Trunc(s string) ([]byte, error) {
	var err error

	cs := make([]byte, sha256.Size/2)
	for ix := 0; ix < len(cs); ix++ {
		bs := s[2*ix : 2*ix+2]
		var scanned int
		_, err := fmt.Sscanf(string(bs), "%x", &scanned)
		cs[ix] = byte(scanned)
		if err != nil {
			return cs, err
		}
		if fmt.Sprintf("%02x", scanned) != bs {
			return cs, fmt.Errorf("invalid hexa code at pos %d: %s", 2*ix, bs)
		}
	}
	return cs, err
}

func Sha256ToPath(cs []byte, size string) string {
	_, csp := Sha256ToStr32Path(cs, size)
	return csp
}

func Sha256ToStr32Path(cs []byte, size string) (string, string) {
	if len(cs) != sha256.Size {
		panic(fmt.Sprintf("len %d != %d", len(cs), sha256.Size))
	}
	css := Sha256ToStr32(cs)
	return css, Str32ToPath(css, size)
}

func Str32ToPath(s string, size string) string {
	if size == "s" {
		return fmt.Sprintf("%s/%s", s[0:2], s[2:32])
	}
	if size == "m" {
		return fmt.Sprintf("%s/%s", s[0:3], s[3:32])
	}
	return fmt.Sprintf("%s/%s/%s", s[0:3], s[3:6], s[6:32])
}

func Int64ToStr16(i int64) string {
	return fmt.Sprintf("%016x", uint64(i))
}

func TimeToStr16(t time.Time) string {
	return Int64ToStr16(t.UnixNano())
}

func Str16ToInt64(s string) (int64, error) {
	if len(s) != 16 {
		return 0, fmt.Errorf("invalid hexa code len %s", s)
	}
	ui, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error decoding hexa code %s: %v", s, err)
	}
	i := int64(ui)
	if Int64ToStr16(i) != s {
		return 0, fmt.Errorf("invalid hexa code %s %x %s", s, ui, Int64ToStr16(i))
	}
	return i, nil
}

func Sec2Nano(sec int64) int64 {
	return sec * 1e9
}

func Nano2SecNano(secnano int64) (int64, int64) {
	return secnano / 1e9, secnano % 1e9
}
