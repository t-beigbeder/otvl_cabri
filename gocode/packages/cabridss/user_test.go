package cabridss

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"strings"
	"testing"
)

func TestGetUserConfig(t *testing.T) {
	tfs, err := testfs.CreateFs("TestGetUserConfig", tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	uc, err := GetHomeUserConfig(DssBaseConfig{})
	if err != nil && !strings.Contains(err.Error(), "passphrase can't be empty") {
		t.Fatal(err)
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
	if err := EncryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path()); err != nil {
		t.Fatal(err)
	}
	id3, err := GenIdentity("id3")
	if err != nil {
		t.Fatal(err, id3)
	}
	if err := UserConfigPutIdentity(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path(), id3); err != nil {
		t.Fatal(err)
	}
	uc5, err := GetUserConfig(DssBaseConfig{}, tfs.Path())
	if err == nil {
		t.Fatal(fmt.Errorf("should fail with error unmarshalling %v", uc5))
	}
	uc6, err := GetUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfigBad"}, tfs.Path())
	if err == nil {
		t.Fatal(fmt.Errorf("should fail with error unmarshalling %v", uc6))
	}
	uc7, err := GetUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path())
	if err != nil || len(uc7.Identities) != 4 {
		t.Fatal(err, uc7)
	}
	if err := DecryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path()); err != nil {
		t.Fatal(err)
	}
	if err := EncryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path()); err != nil {
		t.Fatal(err)
	}
	if err := DecryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfigBad"}, tfs.Path()); err == nil {
		t.Fatal("should fail")
	}
	if err := DecryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path()); err != nil {
		t.Fatal(err)
	}
	if err := DecryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path()); err == nil {
		t.Fatal("should fail")
	}
	if err := EncryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path()); err != nil {
		t.Fatal(err)
	}
	if err := EncryptUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path()); err == nil {
		t.Fatal("should fail")
	}
	if err := SaveUserConfig(DssBaseConfig{ConfigPassword: "TestGetUserConfig"}, tfs.Path(), uc6); err != nil {
		t.Fatal(err)
	}
}
