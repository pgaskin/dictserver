//go:generate packr -v -z
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/geek1011/dictserver/dictionary"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
)

var dict *dictionary.Dictionary

func main() {
	addr := pflag.StringP("addr", "a", ":8000", "Address to listen on")
	help := pflag.BoolP("help", "h", false, "Show this message")
	pflag.Parse()

	if *help || pflag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "Usage: dictserver [OPTIONS]\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nRun the server and go to it in a web browser for API documentation. Warning: the dictionary data uses about 150 MB of memory and takes a few seconds to load for the first time.\n")
		os.Exit(1)
	}

	// TODO: save loaded data
	fmt.Println("Loading dictionary")
	var err error
	dict, err = dictionary.Load()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%d words loaded\n", len(dict.Words))

	r := httprouter.New()

	r.GET("/", cors(noCache(handleAPI)))
	r.GET("/word/:word", cors(handleWord))

	fmt.Printf("Listening on %s\n", *addr)
	err = http.ListenAndServe(*addr, r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

type jmap map[string]interface{}

func (j jmap) WriteTo(w io.Writer) (int64, error) {
	buf, err := json.Marshal(j)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(buf)
	return int64(n), err
}

type apiResponse struct {
	Status string      `json:"status"`
	Result interface{} `json:"result"`
}

func (r apiResponse) WriteTo(w io.Writer) (int64, error) {
	buf, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(buf)
	return int64(n), err
}

func noCache(h httprouter.Handle) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Expires", time.Unix(0, 0).Format(time.RFC1123))
		w.Header().Set("Pragma", "no-cache")
		w.Header().Del("ETag")
		h(w, r, p)
	})
}

func cors(h httprouter.Handle) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h(w, r, p)
	})
}

func handleAPI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	base := "http://" + r.Host

	w.WriteHeader(http.StatusOK)
	jmap{
		"word_url": base + "/word/{word}",
	}.WriteTo(w)
}

func handleWord(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	word, ok := dict.Lookup(p.ByName("word"))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		apiResponse{"error", "word not in dictionary"}.WriteTo(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	apiResponse{"successs", word}.WriteTo(w)
}
