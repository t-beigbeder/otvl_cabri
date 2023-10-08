package cabridss

import "testing"

func TestS3SessionList(t *testing.T) {
	oc := getOC()
	s3s := NewS3Session(oc, nil)
	err := s3s.Initialize()
	if err != nil {
		t.Fatal(err)
	}
	rs, err := s3s.List("content-")
	if err != nil {
		t.Fatal(err)
	}
	_ = rs
}
