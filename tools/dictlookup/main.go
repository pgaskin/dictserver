package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/geek1011/dictserver/dictionary"
)

func main() {
	var idx, db, word string
	switch len(os.Args) {
	case 3:
		idx = os.Args[1] + ".idx"
		db = os.Args[1] + ".db"
		word = os.Args[2]
	case 4:
		idx = os.Args[1]
		db = os.Args[2]
		word = os.Args[3]
	default:
		fmt.Printf("Usage: %s IDX DB WORD\n", os.Args[0])
		fmt.Printf("   or: %s DB_BASE WORD\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Printf("Opening dictionary\n")
	dict, err := dictionary.OpenFile(idx, db)
	if err != nil {
		fmt.Printf("Error opening dictionary: %v\n", err)
		os.Exit(1)
	}
	defer dict.Close()

	fmt.Printf("Looking up word\n")
	w, exists, err := dict.Lookup(word)
	if err != nil {
		fmt.Printf("Error looking up word: %v\n", err)
		os.Exit(1)
	} else if !exists {
		fmt.Printf("\n%s: word not in dictionary\n", strings.ToUpper(word))
	} else {
		fmt.Printf("\n%s:\n", strings.ToUpper(strings.Join(append([]string{w.Word}, w.Alternates...), ", ")))
		fmt.Println(w.Info)
		if w.Etymology != "" {
			fmt.Println(w.Etymology)
		}
		for i, meaning := range w.Meanings {
			fmt.Printf("\n %d. %s\n", i+1, meaning.Text)
			if meaning.Example != "" {
				fmt.Printf("    Example: %s\n", meaning.Example)
			}
		}
		if w.Extra != "" {
			fmt.Printf("Extra: %#v\n", w.Extra)
		}
	}
	fmt.Printf("\n")
}
