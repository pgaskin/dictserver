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
	// HasWord checks if the Store contains a word as-is (i.e. do not do any additional processing or trimming).
	HasWord(word string) bool
	// GetWord gets a word from the Store. If it does not exist, exists will be false, and word and err will be nil.
	GetWord(word string) (w *Word, exists bool, err error)
	// NumWords returns the number of words in the Store.
	NumWords() int
	// Lookup should call Lookup on itself.
	Lookup(word string) (*Word, bool, error)
}

// Word represents a word.
type Word struct {
	Word       string   `json:"word,omitempty"`
	Alternates []string `json:"alternates,omitempty"`
	Info       string   `json:"info,omitempty"`
	Etymology  string   `json:"etymology,omitempty"`
	Meanings   []struct {
		Text    string `json:"text,omitempty"`
		Example string `json:"example,omitempty"`
	} `json:"meanings,omitempty"`
	Extra  string `json:"extra,omitempty"`
	Credit string `json:"credit,omitempty"`
}

// Lookup looks up a word in the dictionary. It applies stemming to the word if no direct match is found. If
// the entry is a reference to another, it will insert the referenced meanings into the result.
func Lookup(store Store, word string) (*Word, bool, error) {
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

	w, exists, err := store.GetWord(ws)
	if err != nil {
		return nil, true, fmt.Errorf("error getting word '%s': %v", ws, err)
	} else if !exists {
		panic("word should exist")
	}

	if len(w.Meanings) == 1 && strings.HasPrefix(w.Meanings[0].Text, "See") && seeOtherRe.MatchString(w.Meanings[0].Text) {
		nw, exists, err := store.GetWord(strings.ToLower(seeOtherRe.FindStringSubmatch(w.Meanings[0].Text)[1]))
		if err == nil && exists {
			w.Meanings = nw.Meanings
		}
	}

	return w, true, nil
}
