//go:build test_testfs

package testfs

import (
	"io/ioutil"
	"log"
	"os"
)

func (f *Fs) plugosint() {
	f.plugin.writable = true
	f.plugin.osCreate = os.Create
	f.plugin.osFileWriteString = (*os.File).WriteString
	f.plugin.ioutilReadFile = ioutil.ReadFile
}

func init() {
	log.Printf("testfs mock initialized")
}
