package core

import (
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"code.google.com/p/gopass"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
)

type PGP struct {
	SecKeyRingPath string
	KeyRingPath    string
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

func GetDefaultSecKeyRingPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(usr.HomeDir, ".gnupg", "secring.gpg"), nil
}

func NewPGP() (*PGP, error) {
	defaultKeyRingPath, err := GetDefaultKeyRingPath()
	if err != nil {
		return nil, err
	}

	defaultSecKeyRingPath, err := GetDefaultSecKeyRingPath()
	if err != nil {
		return nil, err
	}

	return &PGP{
		KeyRingPath:    defaultKeyRingPath,
		SecKeyRingPath: defaultSecKeyRingPath,
	}, nil
}

func ReadKeyRing(keyRingPath string) (*openpgp.EntityList, error) {

	defaultKeyRing, err := os.Open(keyRingPath)
	if err != nil {
		return nil, err
	}

	defer defaultKeyRing.Close()

	entityList, err := openpgp.ReadKeyRing(defaultKeyRing)
	if err != nil {
		return nil, err
	}

	return &entityList, nil
}

func FetchKey(keyid string) error {
	var reply string

	fmt.Printf("PGP Key:%s was not found on your keyring, Do you want to import it? (y/n) ", keyid)
	fmt.Scanf("%s", &reply)
	if reply != "y" {
		return fmt.Errorf("Key %s not found and user skipped importing", keyid)
	}

	_, err := exec.Command("gpg", "--recv-keys", keyid).Output()

	if err != nil {
		return fmt.Errorf("Cannot import key %s from public servers", keyid)
	}

	fmt.Printf("PGP Key: %s imported correctly into the keyring\n", keyid)
	return nil
}

func ConfirmKey(entity *openpgp.Entity, config *Config) bool {
	var answer string
	fmt.Printf("Configuration file Signed-off by PGP Key: %s\n", entity.PrimaryKey.KeyIdShortString())

	for _, identity := range entity.Identities {
		fmt.Printf("* %s\n", identity.Name)
	}

	fmt.Printf("Proceed (y/n)? ")
	fmt.Scanf("%s", &answer)
	return answer == "y"
}

func HasKey(keyid string, entities *openpgp.EntityList) (*openpgp.Entity, error) {
	for _, entity := range *entities {
		if entity.PrimaryKey.CanSign() && entity.PrimaryKey.KeyIdShortString() == keyid {
			return entity, nil
		}
	}

	return nil, fmt.Errorf("cannot find key id: %s", keyid)
}

func (p *PGP) Sign(readed string, keyid string) (string, error) {
	entities, err := ReadKeyRing(p.SecKeyRingPath)
	if err != nil {
		return "", nil
	}

	entity, err := HasKey(keyid, entities)
	if err != nil {
		return "", err
	}

	password, err := gopass.GetPass(fmt.Sprintf("Please insert password for key with id '%s': ",
		entity.PrimaryKey.KeyIdShortString()))
	if err != nil {
		return "", err
	}

	err = entity.PrivateKey.Decrypt([]byte(password))
	if err != nil {
		return "", err
	}

	buff := new(bytes.Buffer)
	if err := openpgp.ArmoredDetachSign(buff, entity, bytes.NewReader([]byte(readed)), nil); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func (p *PGP) Verify(readed string, signed string) (*openpgp.Entity, error) {
	entities, err := ReadKeyRing(p.KeyRingPath)

	if err != nil {
		return nil, err
	}

	signer, err := openpgp.CheckArmoredDetachedSignature(
		entities,
		strings.NewReader(readed),
		strings.NewReader(signed),
	)

	if err != nil {
		return nil, err
	}

	return signer, nil
}
