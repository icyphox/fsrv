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
	"strings"
	"time"

	"github.com/alexedwards/flow"
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

func getKey(r *http.Request) string {
	var key string
	key = r.FormValue("key")
	if key == "" {
		header := r.Header.Get("Authorization")
		split := strings.Split(header, " ")
		if len(split) == 2 {
			key = split[1]
		}
	}
	return key
}

func (s *settings) upload(w http.ResponseWriter, r *http.Request) {
	key := getKey(r)
	useragent := r.Header.Get("User-Agent")
	if key != s.key {
		fmt.Fprintf(w, "incorrect key")
		log.Printf("incorrect key: %+v", key)
		return
	}
	r.ParseMultipartForm(512 << 20)
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
	abs, err := filepath.Abs(diskFile)
	if err != nil {
		log.Println(err)
	}
	runHooks(abs)

	fileUrl := s.url + "/" + newFile
	if strings.Contains(useragent, "curl/") || strings.Contains(useragent, "Igloo/") {
		fmt.Fprintf(w, "%v", fileUrl)
	} else {
		http.Redirect(w, r, fileUrl, http.StatusSeeOther)
	}
}

func (s *settings) indexpage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, s.index)
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

	mux := flow.New()

	st := settings{}
	st.readSettings()

	mux.HandleFunc("/", st.upload, "POST")
	mux.HandleFunc("/", st.indexpage, "GET")
	mux.Handle("/...", http.FileServer(nodirFileSystem{http.Dir(st.storepath)}))

	log.Println("listening on " + st.addr)
	http.ListenAndServe(st.addr, mux)
}
