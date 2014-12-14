package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"time"
)

type Environment interface {
	GetDefaultDirectory() (string, error)
	GetDefaultReportsDirectory() (string, error)
	GetTempReportDirectory() (string, error)
	GetHostName() (string, error)
	GetCurrentTime() time.Time
	GetHomeDir() (string, error)
	GetDefaultStoragePath() (string, error)
}

type DefaultEnvironment struct{}

func (env DefaultEnvironment) GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return usr.HomeDir, nil
}

func (env DefaultEnvironment) GetDefaultStoragePath() (string, error) {
	base, err := env.GetDefaultDirectory()

	if err != nil {
		return "", err
	}

	dir, err := CreateDirIfNotExists(path.Join(base, ".files"), 0700)

	if err != nil {
		return "", err
	}

	return dir, nil
}

func (env DefaultEnvironment) GetHostName() (string, error) {
	return os.Hostname()
}

func (env DefaultEnvironment) GetCurrentTime() time.Time {
	return time.Now().Local()
}

func (env DefaultEnvironment) GetDefaultDirectory() (string, error) {
	home, err := env.GetHomeDir()

	if err != nil {
		return "", err
	}

	dir, err := CreateDirIfNotExists(path.Join(home, ".mayday"), 0700)
	if err != nil {
		return "", err
	}

	return dir, nil
}

func (env DefaultEnvironment) GetDefaultReportsDirectory() (string, error) {
	base, err := env.GetDefaultDirectory()

	if err != nil {
		return "", err
	}

	dir, err := CreateDirIfNotExists(path.Join(base, ".reports"), 0700)

	if err != nil {
		return "", err
	}

	return dir, nil
}

func (env DefaultEnvironment) GetTempReportDirectory() (string, error) {
	hostname, err := env.GetHostName()

	if err != nil {
		return "", err
	}

	base, err := env.GetDefaultReportsDirectory()
	if err != nil {
		return "", err
	}

	current := env.GetCurrentTime()
	formatted := fmt.Sprintf("%s-%d-%d-%d-", hostname, current.Year(), current.Month(), current.Day())
	reportPath, err := ioutil.TempDir(base, formatted)
	if err != nil {
		return "", err
	}

	return reportPath, nil
}
