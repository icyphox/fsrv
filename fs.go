package main

import (
	"net/http"
	"os"
)

type nodirFileSystem struct {
	fs http.FileSystem
}

func (nd nodirFileSystem) Open(path string) (http.File, error) {
	f, err := nd.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		return nil, os.ErrNotExist
	}

	return f, nil
}
