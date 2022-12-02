package cabridss

import (
	"bytes"
	"encoding/json"
	"filippo.io/age"
	"fmt"
	"io"
	"os"
	"strings"
)

var ageId age.Identity

// IdentityConfig refers to an age identity identified by an alias,
// Identities are used for encryption (PKeys of the ACL users using identities aliases)
// and for decryption (secrets of the DSS aclusers using identities aliases)
// "" is the default alias for an identity when none is provided
type IdentityConfig struct {
	Alias  string `json:"alias"`
	PKey   string `json:"pKey"`
	Secret string `json:"secret"`
}

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

// EncryptMsgWithPass encrypts a msg with a given password using Scrypt encoding
//
// It returns the encrypted content as json encoded bytes.
func EncryptMsgWithPass(msg string, pass string) ([]byte, error) {
	r, err := age.NewScryptRecipient(pass)
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsgWithPass: %w", err)
	}
	bsa := bytes.Buffer{}
	wc, err := age.Encrypt(&bsa, r)
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsgWithPass: %w", err)
	}
	_, err = io.Copy(wc, strings.NewReader(msg))
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsgWithPass: %w", err)
	}
	err = wc.Close()
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsgWithPass: %w", err)
	}
	bsb, err := json.Marshal(bsa.Bytes())
	if err != nil {
		return nil, fmt.Errorf("in EncryptMsgWithPass: %w", err)
	}
	return bsb, nil
}

// DecryptMsgWithPass decrypts jbs encrypted content with a given password using Scrypt encoding
// It returns the message in cleartext
//
// jbs are the json encoded bytes
func DecryptMsgWithPass(jbs []byte, pass string) (string, error) {
	id, err := age.NewScryptIdentity(pass)
	if err != nil {
		return "", fmt.Errorf("in DecryptMsgWithPath: %w", err)
	}
	var bs []byte
	if err = json.Unmarshal(jbs, &bs); err != nil {
		return "", fmt.Errorf("in DecryptMsgWithPath: %w", err)
	}
	rd, err := age.Decrypt(bytes.NewReader(bs), id)
	if err != nil {
		return "", fmt.Errorf("in DecryptMsgWithPath: %w", err)
	}
	bss, err := io.ReadAll(rd)
	if err != nil {
		return "", fmt.Errorf("in DecryptMsgWithPath: %w", err)
	}
	return string(bss), nil
}

// EncryptFileWithPass encrypts a file source to a file target with a given password pass using Scrypt encoding
func EncryptFileWithPass(source, target, pass string) error {
	bs, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("in EncryptFileWithPass: %w", err)
	}
	ebs, err := EncryptMsgWithPass(string(bs), pass)
	if err != nil {
		return fmt.Errorf("in EncryptFileWithPass: %w", err)
	}
	if err := os.WriteFile(target, ebs, 0o666); err != nil {
		return fmt.Errorf("in EncryptFileWithPass: %w", err)
	}
	return nil
}

// DecryptFileWithPass decrypts a file source to a file target with a given password pass using Scrypt encoding
func DecryptFileWithPass(source, target, pass string) error {
	ebs, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("in DecryptFileWithPass: %w", err)
	}
	bs, err := DecryptMsgWithPass(ebs, pass)
	if err != nil {
		return fmt.Errorf("in DecryptFileWithPass: %w", err)
	}
	if err := os.WriteFile(target, []byte(bs), 0o666); err != nil {
		return fmt.Errorf("in DecryptFileWithPass: %w", err)
	}
	return nil
}
