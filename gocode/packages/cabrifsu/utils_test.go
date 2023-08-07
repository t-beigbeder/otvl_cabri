package cabrifsu

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func optionalSkip(t *testing.T) {
	if os.Getenv("CABRIFSU_SKIP_DEV_TESTS") != "" {
		if t.Name() == "theBegining" ||
			t.Name() == "theEnd" {
			t.Skip(fmt.Sprintf("Skipping %s because you set CABRIFSU_SKIP_DEV_TESTS", t.Name()))
		}
	}
}

func runCommand(cmdLine string) (stdout, stderr []byte, err error) {
	elems := strings.Split(os.ExpandEnv(cmdLine), " ")
	cmd := exec.Command(elems[0], elems[1:]...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	stdout = outb.Bytes()
	stderr = errb.Bytes()
	return
}
