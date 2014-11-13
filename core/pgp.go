package core

import (
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
)

type PGP struct {
	KeyRingPath string
	KeyRing     *openpgp.EntityList
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
	defaultKeyRingPath, err := GetDefaultKeyRingPath()
	if err != nil {
		return nil, err
	}

	return &PGP{
		KeyRingPath: defaultKeyRingPath,
	}, nil
}

func (p *PGP) UpdateKeyRing() error {
	defaultKeyRing, err := os.Open(p.KeyRingPath)
	if err != nil {
		return err
	}

	defer defaultKeyRing.Close()

	entityList, err := openpgp.ReadKeyRing(defaultKeyRing)
	if err != nil {
		return err
	} else {
		p.KeyRing = &entityList
	}

	return nil
}

func (p *PGP) FetchPGPKey(keyId string) error {
	var reply string

	fmt.Printf("PGP Key:%s was not found on your keyring, Do you want to import it? (y/n) ", keyId)
	fmt.Scanf("%s", &reply)

	if reply != "y" {
		return fmt.Errorf("Key %s not found and user skipped importing", keyId)
	}

	_, err := exec.Command("gpg", "--recv-keys", keyId).Output()

	if err != nil {
		return fmt.Errorf("Cannot import key %s from public servers", keyId)
	}

	fmt.Printf("PGP Key: %s imported correctly into the keyring\n", keyId)
	return nil
}

func (p *PGP) CheckPGPSignature(readed string, signed string) (*PGPSignature, error) {
	err := p.UpdateKeyRing()

	if err != nil {
		return nil, err
	}

	message, err := openpgp.ReadMessage(bytes.NewBuffer([]byte(signed)), p.KeyRing, nil, nil)
	if err != nil {
		return nil, err
	}

	//The message is signed but the signature is missing from the keyring
	if message.IsSigned && message.SignedBy == nil {
		err := p.FetchPGPKey(strconv.FormatUint(message.SignedByKeyId, 16))
		if err != nil {
			return nil, err
		}

		return p.CheckPGPSignature(readed, signed)
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
