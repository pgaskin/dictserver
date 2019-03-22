package main

import (
	"fmt"
	"os"

	"github.com/geek1011/dictserver/dictionary"
)

func main() {
	var txt, idx, db string
	switch len(os.Args) {
	case 3:
		txt = os.Args[1]
		idx = os.Args[2] + ".idx"
		db = os.Args[2] + ".db"
	case 4:
		txt = os.Args[1]
		idx = os.Args[2]
		db = os.Args[3]
	default:
		fmt.Printf("Usage: %s DICT_TXT_IN IDX_OUT DB_OUT\n", os.Args[0])
		fmt.Printf("   or: %s DICT_TXT_IN BASE_OUT\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Printf("Opening input file\n")
	f, err := os.OpenFile(txt, os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("Could not open input file '%s': %v\n", txt, err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Printf("Parsing input file\n")
	wm, err := dictionary.Parse(f)
	if err != nil {
		fmt.Printf("Could not parse dictionary '%s': %v\n", txt, err)
		os.Exit(1)
	}

	fmt.Printf("-- Parsed %d entries\n", wm.NumWords())

	fmt.Printf("Creating database\n")
	err = dictionary.CreateFile(wm, idx, db)
	if err != nil {
		fmt.Printf("Could not export dictionary file to '%s' and '%s': %v\n", idx, db, err)
		os.Exit(1)
	}

	fmt.Printf("Done\n")
}