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
	addr      string
	namelen   int
	key       string
	storepath string
	index     string
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
	switch r.Method {
	case "POST":
		key := r.FormValue("key")
		fmt.Println(key)
		if key != s.key {
			fmt.Fprintf(w, "incorrect key")
			log.Printf("incorrect key: %+v", key)
			return
		}
		r.ParseMultipartForm(20 << 20)
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
	case "GET":
		http.ServeFile(w, r, s.index)
	default:
		fmt.Fprintf(w, "unsupported method")
	}
}

func (s *settings) readSettings() {
	flag.StringVar(&s.url, "url", "localhost", "url for fsrv to serve files")
	flag.StringVar(&s.addr, "addr", "0.0.0.0:9393", "address to listen on")
	flag.StringVar(&s.storepath, "storepath", "uploads", "path to store uploaded files")
	flag.IntVar(&s.namelen, "namelen", 5, "length of random filename")
	flag.StringVar(&s.key, "key", "secret", "secret key; generate this yourself")
	flag.StringVar(&s.index, "index", "index.html", "path to index html file")
	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	st := settings{}
	st.readSettings()

	http.HandleFunc("/", st.uploadFile)

	log.Println("listening on " + st.addr)
	http.ListenAndServe(st.addr, nil)
}
