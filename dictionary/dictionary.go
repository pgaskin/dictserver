// Package dictionary parses the Webster's 1913 dictionary.
package dictionary

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/kljensen/snowball"
)

// Dictionary represents a loaded dictionary.
type Dictionary struct {
	Words   []*Word
	WordMap map[string]*Word
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

var defnCategoryRe = regexp.MustCompile(`^[0-9]+.\s*(\([^\) ]+\))\s*$`)
var singleDefnRe = regexp.MustCompile(`^Defn:\s*(.+)`)
var defnRe = regexp.MustCompile(`^[0-9]+.\s*(.+)`)
var infoAndEtymRe = regexp.MustCompile(`^(.+)\s*Etym:\s*(.+)\s*$`)
var seeOtherRe = regexp.MustCompile(`^\s*See ([A-Za-z]+)\.\s*$`)

// Load loads and parses the dictionary.
func Load() (*Dictionary, error) {
	d := Dictionary{[]*Word{}, map[string]*Word{}}

	buf, err := box.MustBytes("dictionary.txt")
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(bytes.NewReader(buf))

	var s string
	err = nil
	var w *Word
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

		if len(d.Words)%10000 == 0 {
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

				d.Words = append(d.Words, w)
				if _, e := d.WordMap[w.Word]; !e { // Don't overwrite after first definition
					d.WordMap[w.Word] = w
				}
				for _, aw := range w.Alternates {
					if _, e := d.WordMap[aw]; !e { // Don't overwrite after first definition
						d.WordMap[aw] = w
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

	return &d, nil
}

// Lookup looks up a word in the dictionary.
func (d *Dictionary) Lookup(word string) (*Word, bool) {
	// TODO: reduce memory usage, maybe use a binary search instead of a map internally?
	w, ok := d.WordMap[strings.ToLower(word)]
	if !ok {
		if st, err := snowball.Stem(strings.ToLower(word), "english", true); err == nil {
			w, ok = d.WordMap[st]
		}
	}
	if ok {
		if len(w.Meanings) == 1 && strings.HasPrefix(w.Meanings[0].Text, "See") {
			if seeOtherRe.MatchString(w.Meanings[0].Text) {
				if nw, ok := d.WordMap[strings.ToLower(seeOtherRe.FindStringSubmatch(w.Meanings[0].Text)[1])]; ok {
					w.Meanings = nw.Meanings
				}
			}
		}
	}
	return w, ok
}

var box packr.Box

func init() {
	box = packr.NewBox("./data/")
}
