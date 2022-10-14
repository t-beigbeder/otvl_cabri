package internal

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"sort"
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
		s.h = sha256.New()
	}
	return s.h.Write(p)
}

func ShaFrom(r io.Reader) string {
	w := ShaWriter{}
	io.Copy(&w, r)
	cs := w.h.Sum(nil)
	return Sha256ToStr32(cs)
}
