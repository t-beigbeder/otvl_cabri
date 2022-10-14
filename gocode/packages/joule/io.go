package joule

import "io"

type c2w struct {
	oc chan []byte
}

func (c c2w) Write(p []byte) (n int, err error) {
	c.oc <- p
	return len(p), nil
}

func newC2w(oc chan []byte) io.Writer {
	return c2w{oc: oc}
}
