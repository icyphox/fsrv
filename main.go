package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type settings struct {
	url       string
	port      string
	namelen   int
	key       string
	storepath string
}

func randName(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890~-")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *settings) uploadFile(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	fmt.Println(key)
	if key != s.key {
		fmt.Fprintf(w, "incorrect key")
		log.Printf("incorrect key: %+v", key)
		return
	}
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	log.Printf("file: %+v\t%+v bytes", handler.Filename, handler.Size)

	ext := filepath.Ext(handler.Filename)
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Println(err)
	}

	newFile := randName(5) + ext
	diskFile := filepath.Join(s.storepath, newFile)
	os.WriteFile(diskFile, fileBytes, 0644)
	log.Printf("wrote: %+v", diskFile)

	fileUrl := s.url + "/" + newFile
	fmt.Fprintf(w, "%v", fileUrl)
}

func (s *settings) readSettings() {
	flag.StringVar(&s.url, "url", "localhost", "url for fsrv to serve files")
	flag.StringVar(&s.port, "port", "9393", "port to listen on")
	flag.StringVar(&s.storepath, "storepath", "uploads", "path to store uploaded files")
	flag.IntVar(&s.namelen, "namelen", 5, "length of random filename")
	flag.StringVar(&s.key, "key", "secret", "secret key; generate this yourself")
	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	st := settings{}
	st.readSettings()

	http.HandleFunc("/", st.uploadFile)

	log.Println("listening on " + st.port)
	http.ListenAndServe(":"+st.port, nil)
}
