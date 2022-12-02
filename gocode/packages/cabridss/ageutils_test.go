package cabridss

import (
	"bytes"
	"encoding/json"
	"filippo.io/age"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"strings"
	"testing"
)

func TestStartWithEncrypt(t *testing.T) {
	ehw := func(rs ...age.Recipient) *bytes.Buffer {
		bsa := bytes.Buffer{}
		wc, err := age.Encrypt(&bsa, rs...)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(wc, strings.NewReader("Hello world"))
		if err != nil {
			t.Fatal(err)
		}
		err = wc.Close()
		if err != nil {
			t.Fatal(err)
		}
		return &bsa
	}
	k1, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = k1, k2

	rd, err := age.Decrypt(ehw(k1.Recipient()), k1)
	if err != nil {
		t.Fatal(err)
	}
	bsb, err := io.ReadAll(rd)
	if err != nil {
		t.Fatal(err)
	}
	_ = bsb
	rd, err = age.Decrypt(ehw(k1.Recipient()), k2)
	if err == nil {
		t.Fatalf("age.Decrypt should fail with error")
	}
	rd, err = age.Decrypt(ehw(k1.Recipient(), k2.Recipient()), k1)
	if err != nil {
		t.Fatal(err)
	}
	bsb, err = io.ReadAll(rd)
	if err != nil {
		t.Fatal(err)
	}
	rd, err = age.Decrypt(ehw(k1.Recipient(), k2.Recipient()), k2)
	if err != nil {
		t.Fatal(err)
	}
	bsb, err = io.ReadAll(rd)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEncryptArmor(t *testing.T) {
	ear := func(msg string, rs ...age.Recipient) []byte {
		bsa := bytes.Buffer{}
		wc, err := age.Encrypt(&bsa, rs...)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(wc, strings.NewReader(msg))
		if err != nil {
			t.Fatal(err)
		}
		err = wc.Close()
		if err != nil {
			t.Fatal(err)
		}
		bsb, err := json.Marshal(bsa.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		return bsb
	}
	dar := func(jbs []byte, ids ...age.Identity) string {
		var bs []byte
		err := json.Unmarshal(jbs, &bs)
		if err != nil {
			t.Fatal(err)
		}
		rd, err := age.Decrypt(bytes.NewReader(bs), ids...)
		if err != nil {
			t.Fatal(err)
		}
		bss, err := io.ReadAll(rd)
		if err != nil {
			t.Fatal(err)
		}
		return string(bss)
	}
	k1, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = k1, k2
	s := dar(ear("Hello world", k1.Recipient()), k1)
	_ = s
}

func TestEncryptX25519(t *testing.T) {
	ear := func(msg string, srs ...string) []byte {
		bsa := bytes.Buffer{}
		var rs []age.Recipient
		for _, sr := range srs {
			r, err := age.ParseX25519Recipient(sr)
			if err != nil {
				t.Fatal(err)
			}
			rs = append(rs, r)
		}
		wc, err := age.Encrypt(&bsa, rs...)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(wc, strings.NewReader(msg))
		if err != nil {
			t.Fatal(err)
		}
		err = wc.Close()
		if err != nil {
			t.Fatal(err)
		}
		bsb, err := json.Marshal(bsa.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		return bsb
	}
	dar := func(jbs []byte, sids ...string) string {
		var bs []byte
		err := json.Unmarshal(jbs, &bs)
		if err != nil {
			t.Fatal(err)
		}
		var ids []age.Identity
		for _, sid := range sids {
			id, err := age.ParseX25519Identity(sid)
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		rd, err := age.Decrypt(bytes.NewReader(bs), ids...)
		if err != nil {
			t.Fatal(err)
		}
		bss, err := io.ReadAll(rd)
		if err != nil {
			t.Fatal(err)
		}
		return string(bss)
	}
	k1, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	id1 := k1.String()
	id2 := k2.String()
	rec1 := k1.Recipient().String()
	rec2 := k2.Recipient().String()
	s := dar(ear("Hello world", rec1, rec2), id1, id2)
	s = dar(ear("Hello world", rec1, rec2), id1)
	s = dar(ear("Hello world", rec1, rec2), id2)
	s = dar(ear("Hello world", rec1), id1, id2)
	s = dar(ear("Hello world", rec2), id1, id2)
	_ = s
}

func TestGenIdentity(t *testing.T) {
	idc, err := GenIdentity("")
	if err != nil {
		t.Fatal(err)
	}
	em, err := EncryptMsg("", idc.PKey)
	if err != nil {
		t.Fatal(err)
	}
	dm, err := DecryptMsg(em, idc.Secret)
	if err != nil || dm != "" {
		t.Fatal(err, dm)
	}
	em, err = EncryptMsg("TestGenIdentity", idc.PKey)
	if err != nil {
		t.Fatal(err)
	}
	dm, err = DecryptMsg(em, idc.Secret)
	if err != nil || dm != "TestGenIdentity" {
		t.Fatal(err, dm)
	}
	bsa := bytes.Buffer{}
	wc, err := Encrypt(&bsa, idc.PKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = wc.Write([]byte("TestGenIdentity\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err = wc.Close(); err != nil {
		t.Fatal(err)
	}
	rd, err := Decrypt(bytes.NewReader(bsa.Bytes()), idc.Secret)
	if err != nil {
		t.Fatal(err)
	}
	bss, err := io.ReadAll(rd)
	if err != nil || string(bss) != "TestGenIdentity\n" {
		t.Fatal(err, bss)
	}

}

func TestEncryptMsgWithPass(t *testing.T) {
	bs, err := EncryptMsgWithPass("TestEncryptMsgWithPass", "secretLifeOfArabia")
	if err != nil {
		t.Fatal(err)
	}
	msg, err := DecryptMsgWithPass(bs, "secretLifeOfArabia")
	if err != nil || msg != "TestEncryptMsgWithPass" {
		t.Fatal(err, msg)
	}
	msg, err = DecryptMsgWithPass(bs, "secretLifeOfArabia1")
	if err == nil {
		t.Fatal(err)
	}
}

func TestEncryptFileWithPass(t *testing.T) {
	tfs, err := testfs.CreateFs("TestGetUserConfig", tfsStartup)
	if err != nil {
		t.Fatal(err)
	}
	defer tfs.Delete()
	if err = EncryptFileWithPass(ufpath.Join(tfs.Path(), "a.txt"), ufpath.Join(tfs.Path(), "a.etxt"), "TestEncryptFileWithPass"); err != nil {
		t.Fatal(err)
	}
	if err = DecryptFileWithPass(ufpath.Join(tfs.Path(), "a.etxt"), ufpath.Join(tfs.Path(), "a.ctxt"), "TestEncryptFileWithPassBad"); err == nil {
		t.Fatal("should fail with error")
	}
	if err = DecryptFileWithPass(ufpath.Join(tfs.Path(), "a.etxt"), ufpath.Join(tfs.Path(), "a.ctxt"), "TestEncryptFileWithPass"); err != nil {
		t.Fatal(err)
	}
}
