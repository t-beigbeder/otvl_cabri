package cabridss

import (
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"testing"
)

func TestGetUserConfig(t *testing.T) {
	tfs, err := testfs.CreateFs("TestGetUserConfig", tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	uc, err := GetHomeUserConfig(DssBaseConfig{})
	if err != nil || uc.ClientId == "" {
		t.Fatal(err, uc)
	}
	uc2, err := GetHomeUserConfig(DssBaseConfig{})
	if err != nil || uc2.ClientId != uc.ClientId {
		t.Fatal(err, uc2, uc)
	}
	uc3, err := GetUserConfig(DssBaseConfig{}, tfs.Path())
	if err != nil || uc3.ClientId == uc.ClientId || len(uc3.Identities) != 1 {
		t.Fatal(err, uc3, uc)
	}
	id1, err := GenIdentity("id1")
	if err != nil {
		t.Fatal(err, id1)
	}
	if err := UserConfigPutIdentity(DssBaseConfig{}, tfs.Path(), id1); err != nil {
		t.Fatal(err)
	}
	id2, err := GenIdentity("id2")
	if err != nil {
		t.Fatal(err, id2)
	}
	if err := UserConfigPutIdentity(DssBaseConfig{}, tfs.Path(), id2); err != nil {
		t.Fatal(err)
	}
	id1bis, err := GenIdentity("id1")
	if err != nil {
		t.Fatal(err, id1)
	}
	if err := UserConfigPutIdentity(DssBaseConfig{}, tfs.Path(), id1bis); err != nil {
		t.Fatal(err)
	}
	uc4, err := GetUserConfig(DssBaseConfig{}, tfs.Path())
	if err != nil || len(uc4.Identities) != 3 {
		t.Fatal(err, uc4)
	}
}
