package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pgaskin/dictserver/dictionary"
	"github.com/spf13/pflag"
)

var version = "dev"

type ctxKey string

func main() {
	addr := pflag.StringP("addr", "a", ":8000", "Address to listen on")
	help := pflag.BoolP("help", "h", false, "Show this message")
	pflag.Parse()

	var dictfile string
	if n := pflag.NArg(); *help || n != 1 {
		fmt.Printf("Usage: dictserver [options] DICT_FILE\n   or: dictserver [options] DB_BASE\n\nVersion: dictserver %s\n\nOptions:\n", version)
		pflag.PrintDefaults()
		fmt.Printf("\nArguments:\n  DICT_FILE is the path to the dict file. It can be generated using tools/dictparse.\n")
		os.Exit(1)
	} else {
		dictfile = pflag.Arg(0)
	}

	fmt.Printf("Opening dictionary '%s'\n", dictfile)
	dict, err := dictionary.OpenFile(dictfile)
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
	words, exists, err := dict.LookupWord(chi.URLParam(r, "word"))

	switch {
	case err != nil:
		resp{
			statusError,
			fmt.Sprintf("failed to look up word: %v", err),
		}.WriteTo(w, http.StatusInternalServerError)
	case !exists:
		resp{
			statusSuccess,
			[]*dictionary.Word{},
		}.WriteTo(w, http.StatusNotFound)
	default:
		var obj struct {
			*dictionary.Word
			AdditionalWords []*dictionary.Word `json:"additional_words"` // words with the same headword (embedded rather than returning an array for backwards compatibility)
			ReferencedWords []*dictionary.Word `json:"referenced_words"` // referenced words (for the entire word, not just meanings)
		}

		for i, w := range words {
			if i == 0 {
				obj.Word = w
			} else {
				obj.AdditionalWords = append(obj.AdditionalWords, w)
			}
			if len(w.ReferencedWords) != 0 {
				for _, r := range w.ReferencedWords {
					nw, exists, err := dict.GetWords(r)
					if err == nil && exists {
						obj.ReferencedWords = append(obj.ReferencedWords, nw...)
					}
				}
			}
		}

		resp{
			statusSuccess,
			obj,
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	err := enc.Encode(r)
	if err != nil {
		panic(err)
	}
}
