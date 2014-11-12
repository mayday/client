package core

import (
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

type PGP struct {
	KeyRing openpgp.EntityList
}

type PGPSignature struct {
	Keys  []openpgp.Key
	KeyId uint64
}

func GetDefaultKeyRingPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(usr.HomeDir, ".gnupg", "pubring.gpg"), nil
}

func NewPGP() (*PGP, error) {
	pgp := PGP{}

	defaultKeyRingPath, err := GetDefaultKeyRingPath()
	if err != nil {
		return nil, err
	}

	defaultKeyRing, err := os.Open(defaultKeyRingPath)
	if err != nil {
		return nil, err
	}

	defer defaultKeyRing.Close()

	entityList, err := openpgp.ReadKeyRing(defaultKeyRing)
	if err != nil {
		return nil, err
	} else {
		pgp.KeyRing = entityList
	}

	return &pgp, nil
}

func (p *PGP) CheckPGPSignature(readed []byte, signed []byte) (*PGPSignature, error) {
	message, err := openpgp.ReadMessage(bytes.NewBuffer(signed), p.KeyRing, nil, nil)

	//TODO: Handle the case on which the signature is not on the keyring
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(message.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("error reading message body: %s", err)
	}

	if string(contents) != string(readed) {
		return nil, fmt.Errorf("Incorrect signature, not valid")
	}

	if message.SignatureError != nil || message.Signature == nil {
		return nil, fmt.Errorf("failed to validate signature: %s",
			message.SignatureError)
	}

	return &PGPSignature{
		Keys:  p.KeyRing.KeysById(message.SignedByKeyId),
		KeyId: message.SignedByKeyId,
	}, nil
}
