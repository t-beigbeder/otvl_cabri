package cabridss

import (
	"bytes"
	"encoding/json"
	"filippo.io/age"
	"fmt"
	"io"
	"strings"
)

var ageId age.Identity

func GenIdentity(alias string) (IdentityConfig, error) {
	xi, err := age.GenerateX25519Identity()
	if err != nil {
		return IdentityConfig{}, fmt.Errorf("in GenIdentity: %w", err)
	}
	return IdentityConfig{alias, xi.Recipient().String(), xi.String()}, nil
}

// EncryptMsg encrypts a msg to one or more X25519 srs recipients encoded as strings.
//
// Every recipient will be able to decrypt the result.
//
// It returns the encrypted content as json encoded bytes.
func EncryptMsg(msg string, srs ...string) ([]byte, error) {
	bsa := bytes.Buffer{}
	var rs []age.Recipient
	for _, sr := range srs {
		r, err := age.ParseX25519Recipient(sr)
		if err != nil {
			return nil, fmt.Errorf("in EncryptMsg: %w", err)
		}
		rs = append(rs, r)
	}
	wc, err := age.Encrypt(&bsa, rs...)
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsg: %w", err)
	}
	_, err = io.Copy(wc, strings.NewReader(msg))
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsg: %w", err)
	}
	err = wc.Close()
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsg: %w", err)
	}
	bsb, err := json.Marshal(bsa.Bytes())
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsg: %w", err)
	}
	return bsb, nil
}

// DecryptMsg decrypts jbs encrypted content to one or more sids X25519 identities encoded as strings.
// It returns the message in cleartext
//
// jbs are the json encoded bytes
//
// All identities will be tried until one successfully decrypts the content.
func DecryptMsg(jbs []byte, sids ...string) (string, error) {
	var bs []byte
	err := json.Unmarshal(jbs, &bs)
	if err != nil {
		return "", fmt.Errorf("in DecryptMsg: %w", err)
	}
	var ids []age.Identity
	for _, sid := range sids {
		id, err := age.ParseX25519Identity(sid)
		if err != nil {
			return "", fmt.Errorf("in DecryptMsg: %w", err)
		}
		ids = append(ids, id)
	}
	rd, err := age.Decrypt(bytes.NewReader(bs), ids...)
	if err != nil {
		return "", fmt.Errorf("in DecryptMsg: %w", err)
	}
	bss, err := io.ReadAll(rd)
	if err != nil {
		return "", fmt.Errorf("in DecryptMsg: %w", err)
	}
	return string(bss), nil
}

// Encrypt encrypts a file to one or more X25519 srs recipients encoded as strings.
//
// Writes to the returned WriteCloser are encrypted and written to dst as an age file.
// Every recipient will be able to decrypt the file.
//
// The caller must call Close on the WriteCloser when done for the last chunk to be encrypted and flushed to dst.
func Encrypt(dst io.Writer, srs ...string) (io.WriteCloser, error) {
	var rs []age.Recipient
	for _, sr := range srs {
		r, err := age.ParseX25519Recipient(sr)
		if err != nil {
			return nil, fmt.Errorf("in EncryptMsg: %w", err)
		}
		rs = append(rs, r)
	}
	return age.Encrypt(dst, rs...)
}

// Decrypt decrypts a file encrypted to one or more sids X25519 identities encoded as strings.
//
// It returns a Reader reading the decrypted plaintext of the age file read from src.
// All identities will be tried until one successfully decrypts the file.
func Decrypt(src io.Reader, sids ...string) (io.Reader, error) {
	var ids []age.Identity
	for _, sid := range sids {
		id, err := age.ParseX25519Identity(sid)
		if err != nil {
			return nil, fmt.Errorf("in DecryptMsg: %w", err)
		}
		ids = append(ids, id)
	}
	return age.Decrypt(src, ids...)
}
