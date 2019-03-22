package dictionary

import (
	"bufio"
	"io"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/kljensen/snowball"
)

// WordMap is an im-memory word Store used and returned by Parse. Although fast, it consumes huge
// amounts of memory (~500 MB) and shouldn't be used if possible.
type WordMap map[string]*Word

// HasWord implements Store.
func (wm WordMap) HasWord(word string) bool {
	_, ok := wm[word]
	return ok
}

// GetWord implements Store, but will never return an error.
func (wm WordMap) GetWord(word string) (*Word, bool, error) {
	w, ok := wm[word]
	return w, ok, nil
}

// NumWords implements Store.
func (wm WordMap) NumWords() int {
	return len(wm)
}

// Lookup is a shortcut for Lookup.
func (wm WordMap) Lookup(word string) (*Word, bool, error) {
	return Lookup(wm, word)
}

// Regexps used by the parser.
var (
	defnCategoryRe = regexp.MustCompile(`^[0-9]+.\s*(\([^\) ]+\))\s*$`)
	singleDefnRe   = regexp.MustCompile(`^Defn:\s*(.+)`)
	defnRe         = regexp.MustCompile(`^[0-9]+.\s*(.+)`)
	infoAndEtymRe  = regexp.MustCompile(`^(.+)\s*Etym:\s*(.+)\s*$`)
	seeOtherRe     = regexp.MustCompile(`^\s*See ([A-Za-z]+)\.\s*$`)
)

// Parse parses Webster's Unabridged Dictionary of 1913 into a WordMap. It will
// only return an error if the reader returns one. If the data is corrupt, the
// results are undefined (but will be tried to be parsed as best as possible).
// WARNING: Parse uses huge amounts of memory (~600 MB) and cpu time (~30s).
func Parse(rd io.Reader) (WordMap, error) {
	wm := WordMap{}
	r := bufio.NewReader(rd)

	var i int
	var s string
	var w *Word
	var err error
	var temp string
	for {
		s, err = r.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			break
		}
		s = strings.Trim(s, "\r\n\t ")

		i++
		if i%10000 == 0 {
			debug.FreeOSMemory()
		}

		if len(s) > 0 && strings.ToUpper(s) == s { // Start of a word
			if w != nil {
				spl := strings.Split(temp, "\n")
				state := "info"
				for i, s := range spl {
					if state == "defn-category" && defnRe.MatchString(s) { // Categories cannot contain numbered defns, only lettered.
						state = ""
					}
					// Parse order is VERY important.
					if state == "info" && s == "" { // Info is up to the first blank line.
						if infoAndEtymRe.MatchString(w.Info) {
							w.Etymology = infoAndEtymRe.FindStringSubmatch(w.Info)[2]
							w.Info = infoAndEtymRe.FindStringSubmatch(w.Info)[1]
						}
						state = ""
						continue
					} else if state == "info" {
						w.Info += s + " "
						continue
					} else if defnCategoryRe.MatchString(s) {
						state = "defn-category" // Can contain single defn.
						w.Meanings = append(w.Meanings, struct {
							Text    string `json:"text,omitempty"`
							Example string `json:"example,omitempty"`
						}{Text: defnCategoryRe.FindStringSubmatch(s)[1]})
						continue
					} else if state == "defn-category" {
						if singleDefnRe.MatchString(s) {
							w.Meanings[len(w.Meanings)-1].Text += " " + singleDefnRe.FindStringSubmatch(s)[1]
						} else {
							if len(w.Meanings[len(w.Meanings)-1].Text) > 5 && len(spl[i-1]) < 55 && strings.HasSuffix(strings.TrimSpace(spl[i-1]), ".") && !strings.Contains(s, "Note: ") && !strings.Contains(s, "-- ") { // If last line (shorter than hard wrap) and ends with a dot.
								state = "defn-example"
								w.Meanings[len(w.Meanings)-1].Example = s
							} else {
								w.Meanings[len(w.Meanings)-1].Text += " " + s
							}
						}
						continue
					} else if defnRe.MatchString(s) {
						state = "defn" // Contains text until blank line.
						w.Meanings = append(w.Meanings, struct {
							Text    string `json:"text,omitempty"`
							Example string `json:"example,omitempty"`
						}{Text: defnRe.FindStringSubmatch(s)[1]})
						continue
					} else if state == "defn" {
						if len(w.Meanings[len(w.Meanings)-1].Text) > 5 && len(spl[i-1]) < 55 && strings.HasSuffix(strings.TrimSpace(spl[i-1]), ".") && !strings.Contains(s, "Note: ") && !strings.Contains(s, "-- ") { // If last line (shorter than hard wrap) and ends with a dot.
							state = "defn-example"
							w.Meanings[len(w.Meanings)-1].Example = s
						} else {
							w.Meanings[len(w.Meanings)-1].Text += " " + s
						}
						continue
					} else if state == "defn-example" {
						w.Meanings[len(w.Meanings)-1].Example += " " + s
						continue
					} else if singleDefnRe.MatchString(s) {
						state = "single-defn" // Contains text until blank line.
						w.Meanings = append(w.Meanings, struct {
							Text    string `json:"text,omitempty"`
							Example string `json:"example,omitempty"`
						}{Text: singleDefnRe.FindStringSubmatch(s)[1]})
						continue
					} else if state == "single-defn" {
						w.Meanings[len(w.Meanings)-1].Text += " " + s
						continue
					} else if s == "" {
						continue
					} else {
						w.Extra += "\n" + s
						continue
					}
				}

				if _, e := wm[w.Word]; !e { // Don't overwrite after first definition
					wm[w.Word] = w
				}
				for _, aw := range w.Alternates {
					if _, e := wm[aw]; !e { // Don't overwrite after first definition
						wm[aw] = w
					}
				}
				if sw, err := snowball.Stem(w.Word, "english", false); err == nil {
					if _, e := wm[sw]; !e { // Don't overwrite after first definition
						wm[sw] = w
					}
				}
			}

			spl := strings.Split(strings.ToLower(s), ";")
			w = &Word{Word: strings.TrimSpace(spl[0]), Alternates: []string{}, Credit: "Webster's Unabridged Dictionary (1913)"}
			if len(spl) > 1 {
				spl = spl[1:]
				for _, alt := range spl {
					w.Alternates = append(w.Alternates, strings.TrimSpace(alt))
				}
			}

			temp = ""
			continue
		}

		if w == nil {
			continue
		}

		temp += s + "\n"
		continue
	}
	if err != nil {
		return nil, err
	}

	debug.FreeOSMemory()

	return wm, nil
}
