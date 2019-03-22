package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/geek1011/dictserver/dictionary"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/pflag"
)

var version = "v2.0.0-dev"

type ctxKey string

func main() {
	addr := pflag.StringP("addr", "a", ":8000", "Address to listen on")
	help := pflag.BoolP("help", "h", false, "Show this message")
	pflag.Parse()

	var idx, db string
	if n := pflag.NArg(); *help || (n != 1 && n != 2) {
		fmt.Printf("Usage: dictserver [options] IDX_FILE DB_FILE\n   or: dictserver [options] DB_BASE\n\nVersion: dictserver %s\n\nOptions:\n", version)
		pflag.PrintDefaults()
		fmt.Printf("\nArguments:\n  IDX_FILE and DB_FILE are the path to the respective *.idx and *.db files. If the basename\n  is the same, DB_BASE can be used instead. These files can be generated using tools/dictparse.\n")
		os.Exit(1)
	} else if n == 1 {
		idx = pflag.Arg(0) + ".idx"
		db = pflag.Arg(0) + ".db"
	} else {
		idx = pflag.Arg(0)
		db = pflag.Arg(1)
	}

	fmt.Printf("Opening dictionary '%s' and '%s'\n", idx, db)
	dict, err := dictionary.OpenFile(idx, db)
	if err != nil {
		fmt.Printf("Error opening dictionary: %v\n", err)
		os.Exit(1)
	}
	defer dict.Close()
	fmt.Printf("-- Loaded %d entries\n", dict.NumWords())

	fmt.Printf("Listening on http://%s\n", *addr)
	err = http.ListenAndServe(*addr, router(dict))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}

func router(dict dictionary.Store) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.GetHead)
	r.Use(middleware.NoCache) // TODO: cache based on dict file last mod date
	r.Use(middleware.StripSlashes)
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))
	r.Use(middleware.SetHeader("Server", "dictserver ("+version+")"))
	r.Use(middleware.WithValue(ctxKey("dict"), dict))

	r.NotFound(handleNotFound)
	r.Get("/", handleAPI)
	r.Get("/word/{word}", handleWord)

	return r
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	resp{
		statusError,
		"not found",
	}.WriteTo(w, http.StatusNotFound)
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	base := "http://" + r.Host
	resp{
		statusSuccess,
		map[string]string{
			"word_url": base + "/word/{word}",
		},
	}.WriteTo(w, http.StatusOK)
}

func handleWord(w http.ResponseWriter, r *http.Request) {
	dict := r.Context().Value(ctxKey("dict")).(dictionary.Store)
	word, exists, err := dict.Lookup(chi.URLParam(r, "word"))

	switch {
	case err != nil:
		resp{
			statusError,
			fmt.Errorf("error looking up word: %v", err),
		}.WriteTo(w, http.StatusInternalServerError)
	case !exists:
		resp{
			statusError,
			fmt.Errorf("word not found"),
		}.WriteTo(w, http.StatusNotFound)
	default:
		resp{
			statusSuccess,
			word,
		}.WriteTo(w, http.StatusOK)
	}
}

type status string

const (
	statusSuccess status = "success"
	statusError   status = "error"
)

type resp struct {
	Status status      `json:"status"`
	Result interface{} `json:"result"`
}

func (r resp) WriteTo(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	err := enc.Encode(r)
	if err != nil {
		panic(err)
	}
}
