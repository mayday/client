package core

import (
	"io"
	"os"
	"regexp"
)

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
