package cabriui

import (
	"os"
	"testing"
)

var sampleOptions SampleOptions

func TestSampleStartup(t *testing.T) {
	optionalSkip(t)
	err := CLIRun[SampleOptions, SampleVars](
		nil, os.Stdout, os.Stderr,
		sampleOptions, nil,
		SampleStartup, SampleShutdown)
	if err != nil {
		t.Fatal(err)
	}

}
