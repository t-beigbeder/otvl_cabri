package internal

import (
	"crypto/sha256"
	"sort"
)

// Ns2Content provides the namespace content
// along with its truncated SHA256 checksum as a path
func Ns2Content(children []string, size string) (string, string, string) {
	sort.Strings(children)
	content := ""
	for _, child := range children {
		content += child + "\n"
	}
	cs := sha256.Sum256([]byte(content))
	path := ""
	if size != "" {
		path = Sha256ToPath(cs[:], size)
	}
	return content, Sha256ToStr32(cs[:]), path
}

func BytesToSha256Str(bs []byte) string {
	cs := sha256.Sum256(bs)
	return Sha256ToStr32(cs[:])
}

type NpType string

func (npath NpType) ExistIn(nps []string) bool {
	for _, onp := range nps {
		if onp == string(npath) {
			return true
		}
	}
	return false
}
