package core

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"regexp"
	"time"
)

func GetDefaultDirectory() (string, error) {
	usr, err := user.Current()

	if err != nil {
		return "", err
	}

	base := path.Join(usr.HomeDir, ".mayday")

	if _, err := os.Stat(base); os.IsNotExist(err) {
		err := os.MkdirAll(base, 0700)
		if err != nil {
			return "", err
		}
	}

	return base, nil
}

func GetDefaultReportsDirectory() (string, error) {
	base, err := GetDefaultDirectory()

	if err != nil {
		return "", err
	}

	reports := path.Join(base, ".reports")

	if _, err := os.Stat(reports); os.IsNotExist(err) {
		err := os.MkdirAll(reports, 0700)
		if err != nil {
			return "", err
		}
	}

	return base, nil
}

func GetTempReportDirectory() (string, error) {
	hostname, err := os.Hostname()

	if err != nil {
		return "", err
	}

	current := time.Now().Local()

	base, err := GetDefaultReportsDirectory()
	if err != nil {
		return "", err
	}

	//TODO: Make the report format configurable via CLI
	reportPath, err := ioutil.TempDir(base,
		fmt.Sprintf("%s-%d-%d-%d-", hostname, current.Year(), current.Month(), current.Day()))

	if err != nil {
		return "", err
	}

	return reportPath, nil
}

func GetReportConfigFiles(base string) (string, string) {
	return path.Join(base, "config.yaml"), path.Join(base, "config.yaml.sig")
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}

func CopyDir(source string, dest string) (err error) {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			CopyDir(sourcefilepointer, destinationfilepointer)
		} else {
			CopyFile(sourcefilepointer, destinationfilepointer)
		}

	}

	return nil
}

func MangleCommand(command string) string {
	//Ported from https://github.com/sosreport/sos/blob/48a99c95078bab306cb56bb1a05420d88bf15a64/sos/plugins/__init__.py
	regexes := map[string](string){
		"^/(usr/|)(bin|sbin)/": "",
		"[^\\w\\-\\.\\/]+":     "_",
		"/":                    ".",
		"[\\.|_|-|\\t]+":       "",
	}

	for regex, replace := range regexes {
		command = regexp.MustCompile(regex).ReplaceAllLiteralString(command, replace)
	}

	return command
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
