package dictionary

import (
	"io"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/pgaskin/dictutil/examples/webster1913-convert/webster1913"
)

// WordMap is an im-memory word Store used and returned by Parse. Although fast,
// it consumes huge amounts of memory and shouldn't be used if possible.
type WordMap map[string][]*Word

// HasWord implements Store.
func (wm WordMap) HasWord(word string) bool {
	_, ok := wm[word]
	return ok
}

// GetWords implements Store, but will never return an error.
func (wm WordMap) GetWords(word string) ([]*Word, bool, error) {
	ws, ok := wm[word]
	return ws, ok, nil
}

// GetWord is deprecated.
func (wm WordMap) GetWord(word string) (*Word, bool, error) {
	ws, exists, err := wm.GetWords(word)
	if len(ws) == 0 {
		return nil, exists, err
	}
	return ws[0], exists, err
}

// NumWords implements Store.
func (wm WordMap) NumWords() int {
	return len(wm)
}

// Lookup is deprecated.
func (wm WordMap) Lookup(word string) (*Word, bool, error) {
	return Lookup(wm, word)
}

// LookupWord is a shortcut for LookupWord.
func (wm WordMap) LookupWord(word string) ([]*Word, bool, error) {
	return LookupWord(wm, word)
}

// Parse parses Webster's Unabridged Dictionary of 1913 into a WordMap. Note:
// For dictserver > v1.3.1, this now uses the parser I implemented for dictutil
// which is much more efficient and accurate.
func Parse(r io.Reader) (WordMap, error) {
	refRe := regexp.MustCompile(`See(?: under)? ([A-Z][a-z]+)\.`)

	d, err := webster1913.Parse(r, func(i int, w string) {})
	if err != nil {
		return nil, err
	}

	wm := WordMap{}
	for _, e := range d {
		w := &Word{}

		w.Word = e.Headword
		w.Etymology = e.Etymology
		w.Info = e.Info
		for _, d := range e.Meanings {
			x := WordMeaning{
				Text:    d.Text,
				Example: d.Example,
			}
			for _, m := range refRe.FindAllStringSubmatch(d.Text, -1) {
				x.ReferencedWords = append(x.ReferencedWords, strings.ToLower(m[1]))
			}
			w.Meanings = append(w.Meanings, x)
		}
		if len(e.PhraseDefns) != 0 {
			w.Notes = append(w.Notes, strings.Join(e.PhraseDefns, "\u00A0\u00A0\u00A0"))
		}
		if len(e.Synonyms) != 0 {
			w.Notes = append(w.Notes, strings.Join(e.Synonyms, "\u00A0\u00A0\u00A0"))
		}
		w.Extra = e.Extra
		w.Credit = "Webster's Unabridged Dictionary (1913)"
		for _, m := range refRe.FindAllStringSubmatch(e.Etymology, -1) {
			w.ReferencedWords = append(w.ReferencedWords, strings.ToLower(m[1]))
		}

		wm[e.Headword] = append(wm[e.Headword], w)
		for _, v := range e.Variant {
			wm[v] = append(wm[v], w)
			w.Alternates = append(w.Alternates, v)
		}
	}

	debug.FreeOSMemory()

	return wm, nil
}
