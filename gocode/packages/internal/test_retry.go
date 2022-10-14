package internal

import (
	"testing"
)

const MAX_RETRY = 3

func Retry(t *testing.T, f func(t *testing.T) error) error {
	for i := 0; i < MAX_RETRY; i++ {
		if err := f(t); err != nil {
			if i == MAX_RETRY-1 {
				t.Fatalf("Retry: test %s failed with error %v", t.Name(), err)
				return err
			}
			continue
		}
		break
	}
	return nil
}
