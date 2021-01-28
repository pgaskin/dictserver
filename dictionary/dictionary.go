// Package dictionary contains code related to looking up and storing words in dictionaries. The parser currently
// supports the Project Gutenberg's edition of the Webster's Unabridged 1913 dictionary.
package dictionary

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Store is a backend for storing dictionary entries. Implementations should not
// return duplicate entries, but it is not a bug to do so.
type Store interface {
	// NumWords returns the number of words in the Store.
	NumWords() int
	// HasWord checks if the Store contains a word as-is (i.e. do not do any additional processing or trimming).
	HasWord(word string) bool

	// GetWords gets a word, which can have multiple instances, from the Store.
	// If it does not exist, exists will be false, and word and err will be nil.
	GetWords(word string) (w []*Word, exists bool, err error)
	// GetWord is deprecated.
	GetWord(word string) (w *Word, exists bool, err error)

	// LookupWord should call LookupWord on itself.
	LookupWord(word string) ([]*Word, bool, error)
	// Lookup is deprecated.
	Lookup(word string) (*Word, bool, error)
}

// Word represents a word.
type Word struct {
	Word            string        `json:"word,omitempty" msgpack:"w"`
	Alternates      []string      `json:"alternates,omitempty" msgpack:"a"`
	Info            string        `json:"info,omitempty" msgpack:"i"`
	Etymology       string        `json:"etymology,omitempty" msgpack:"e"`
	Meanings        []WordMeaning `json:"meanings,omitempty" msgpack:"m"`
	Notes           []string      `json:"notes,omitempty" msgpack:"n"`
	Extra           string        `json:"extra,omitempty" msgpack:"x"`
	Credit          string        `json:"credit,omitempty" msgpack:"c"`
	ReferencedWords []string      `json:"referenced_words" msgpack:"r"` // note: this does not include words referenced within meanings
}

type WordMeaning struct {
	Text            string   `json:"text,omitempty" msgpack:"t"`
	Example         string   `json:"example,omitempty" msgpack:"e"`
	ReferencedWords []string `json:"referenced_words" msgpack:"r"`
}

// Lookup looks up the first entry for a word in the dictionary (deprecated). It
// applies normalization and stemming to the word if no direct match is found.
func Lookup(store Store, word string) (*Word, bool, error) {
	ws, exists, err := LookupWord(store, word)
	if len(ws) == 0 {
		return nil, exists, err
	}
	return ws[0], exists, err
}

var (
	normSpaceRe     = regexp.MustCompile(`\s+`)
	normDashRe      = regexp.MustCompile(`\p{Pd}`)
	normADashRe     = regexp.MustCompile(`-+`)
	normOpenCloseRe = regexp.MustCompile(`^(?:\p{Pi}|\p{Ps}|["'])+|(?:\p{Pf}|\p{Pe}|["'])+$`)
	normTransform   = transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
)

// LookupWord looks up a word in the dictionary. It applies normalization and
// stemming to the word if no direct match is found.
func LookupWord(store Store, word string) ([]*Word, bool, error) {
	var err error

	ws := word

	for a := 0; a < 2; a++ {
		// trim leading and trailing spaces
		if ws = strings.ToLower(strings.TrimSpace(word)); store.HasWord(ws) {
			goto found
		}

		// collapse all whitespace into a single space
		if ws = normSpaceRe.ReplaceAllLiteralString(ws, " "); store.HasWord(ws) {
			goto found
		}

		// trim leading and trailing opening/closing punctuation
		if ws = normOpenCloseRe.ReplaceAllLiteralString(ws, ""); store.HasWord(ws) {
			goto found
		}

		// replace all unicode dash-like characters with a dash
		if ws = normDashRe.ReplaceAllLiteralString(ws, "-"); store.HasWord(ws) {
			goto found
		}

		// collapse multiple dashes
		if ws = normADashRe.ReplaceAllLiteralString(ws, "-"); store.HasWord(ws) {
			goto found
		}

		for b := 0; b < 2; b++ {
			// stem
			if wst, err := snowball.Stem(ws, "english", true); err == nil && store.HasWord(wst) {
				ws = wst
				goto found
			}

			// sometimes stemming removes too much
			if wst := strings.TrimRight(ws, "s"); store.HasWord(wst) {
				ws = wst
				goto found
			}

			// sometimes stemming removes too much
			if wst := strings.TrimSuffix(strings.TrimSuffix(ws, "ly"), "ing"); store.HasWord(ws) {
				ws = wst
				goto found
			}

			// try again, but fold all unicode chars into their bases
			if b == 0 {
				if ws, _, err = transform.String(normTransform, ws); err != nil {
					break
				} else if store.HasWord(ws) {
					goto found
				}
			}
		}

		// try again, but remove dashes
		if a == 0 {
			if ws = strings.Replace(ws, "-", "", -1); store.HasWord(ws) {
				goto found
			}
		}
	}

	return nil, false, nil

found:
	w, exists, err := store.GetWords(ws)
	if err != nil {
		return nil, true, fmt.Errorf("error getting word '%s': %v", ws, err)
	} else if !exists {
		panic("word should exist if HasWord")
	}

	return w, true, nil
}
