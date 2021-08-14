package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/h2non/filetype"
)

func runHooks(file string) {
	hooks, err := os.ReadDir("hooks")
	if err != nil {
		log.Println(err)
	}
	for _, h := range hooks {
		hookFile := getHook(file)
		if h.Name() == hookFile {
			log.Println("running hook:", hookFile)
			cmd := exec.Command(filepath.Join("hooks", h.Name()), file)
			stdout, _ := cmd.StdoutPipe()
			cmd.Start()
			s := bufio.NewScanner(stdout)
			for s.Scan() {
				fmt.Println(s.Text())
			}
		}
	}
}

// Checks the MIME type of file and returns
// the corresponding hook file.
func getHook(file string) string {
	// Not sure how many bytes the magic number takes, but 16
	// is a good guess. I think.
	magic := make([]byte, 16)

	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	_, err = f.Read(magic)
	if err != nil {
		log.Println(err)
	}

	t, err := filetype.Match(magic)
	if err != nil {
		log.Println(err)
	}
	return t.Extension + ".sh"
}
