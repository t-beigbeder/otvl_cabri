package joule

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func optionalSleep(t *testing.T) {
	if os.Getenv("JOULE_FAST_TESTS") == "" {
		time.Sleep(1100 * time.Millisecond)
	}
}

func optionalKeep(t *testing.T) {
	if os.Getenv("JOULE_KEEP_DEV_TESTS") == "" {
		if t.Name() == "TestControlC" {
			t.Skip(fmt.Sprintf("Skipping %s because you didn't set JOULE_KEEP_DEV_TESTS", t.Name()))
		}
	}
}
