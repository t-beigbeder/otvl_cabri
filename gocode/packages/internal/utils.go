package internal

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"sort"
	"strconv"
	"time"
)

type SliceStringer[T fmt.Stringer] struct{ Slice []T }

func (ss SliceStringer[T]) String() (res string) {
	for _, s := range ss.Slice {
		if res != "" {
			res += "\n"
		}
		res += s.String()
	}
	return
}

type StringStringer string

func (s StringStringer) String() string { return string(s) }

type StringsStringer []string

func (sss StringsStringer) String() string { return fmt.Sprintf("%v", []string(sss)) }

type StringSliceEOL []string

func (ss StringSliceEOL) String() (res string) {
	for _, s := range ss {
		if res != "" {
			res += "\n"
		}
		res += s
	}
	return
}

type MapSliceStringer[T fmt.Stringer] struct{ Map map[string][]T }

func (ms MapSliceStringer[T]) String() (res string) {
	keys := []string{}
	for k, _ := range ms.Map {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if res != "" {
			res += "\n"
		}
		res += "\"" + k + "\"\n" + SliceStringer[T]{Slice: ms.Map[k]}.String()
	}
	return
}

type ShaWriter struct {
	bs []byte
	h  hash.Hash
}

func (s *ShaWriter) Write(p []byte) (n int, err error) {
	if s.bs == nil {
		s.bs = make([]byte, 8192)
	}
	return s.h.Write(p)
}

func ShaFrom(r io.Reader) (string, error) {
	w := ShaWriter{h: sha256.New()}
	if _, err := io.Copy(&w, r); err != nil {
		return "", err
	}
	cs := w.h.Sum(nil)
	return Sha256ToStr32(cs), nil
}

func CheckTimeStamp(value string) (unix int64, err error) {
	if value == "" {
		return
	}
	var ts time.Time
	if ts, err = time.Parse(time.RFC3339, value); err == nil {
		unix = ts.Unix()
		return
	}
	if unix, err = strconv.ParseInt(value, 10, 64); err == nil {
		return
	}
	err = fmt.Errorf("timestamp %s must be either RFC3339 (eg 2020-08-13T11:56:41Z) or a unix time integer", value)
	return
}
