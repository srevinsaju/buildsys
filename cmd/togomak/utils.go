package main

import (
	"github.com/moby/sys/mountinfo"
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"log"
	"path"
	"path/filepath"
)

func autoDetectFilePath(cwd string) string {
	fs := afero.NewOsFs()
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		panic(err)
	}
	p := path.Join(cwd, meta.ConfigFileName)
	exists, err := afero.Exists(fs, p)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		return p
	}
	p2 := path.Join(cwd, meta.BuildDirPrefix, meta.ConfigFileName)
	exists, err = afero.Exists(fs, p2)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		return p2
	}

	mountPoint, err := mountinfo.Mounted(absPath)
	if mountPoint {
		log.Fatalf("Couldn't find %s. Searched until %s", meta.ConfigFileName, absPath)
	}

	return autoDetectFilePath(path.Join(cwd, ".."))

}
