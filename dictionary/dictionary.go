// Package dictionary contains code related to looking up and storing words in dictionaries. The parser currently
// supports the Project Gutenberg's edition of the Webster's Unabridged 1913 dictionary.
package dictionary

import (
	"fmt"
	"strings"

	"github.com/kljensen/snowball"
)

// Store is a backend for storing dictionary entries.
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
// applies stemming to the word if no direct match is found.
func Lookup(store Store, word string) (*Word, bool, error) {
	ws, exists, err := LookupWord(store, word)
	if len(ws) == 0 {
		return nil, exists, err
	}
	return ws[0], exists, err
}

// LookupWord looks up a word in the dictionary. It applies stemming to the word
// if no direct match is found.
func LookupWord(store Store, word string) ([]*Word, bool, error) {
	var err error

	ws := strings.ToLower(strings.TrimSpace(word))
	if !store.HasWord(ws) {
		ws, err = snowball.Stem(strings.ToLower(word), "english", true)
		if err != nil {
			return nil, false, fmt.Errorf("could not stem word: %v", err) // this should never happen
		}
		if !store.HasWord(ws) {
			ws = strings.TrimRight(strings.ToLower(word), "s") // sometimes stemming removes too much, just need to remove the plural
			if !store.HasWord(ws) {
				ws = strings.Replace(strings.Replace(strings.ToLower(word), "ing", "", 1), "ly", "", 1)
				if !store.HasWord(ws) {
					return nil, false, nil
				}
			}
		}
	}

	w, exists, err := store.GetWords(ws)
	if err != nil {
		return nil, true, fmt.Errorf("error getting word '%s': %v", ws, err)
	} else if !exists {
		panic("word should exist if HasWord")
	}

	return w, true, nil
}
