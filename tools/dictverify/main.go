package main

import (
	"fmt"
	"os"

	"github.com/pgaskin/dictserver/dictionary"
)

func main() {
	var dictfile string
	switch len(os.Args) {
	case 2:
		dictfile = os.Args[1]
	default:
		fmt.Printf("Usage: %s DICT_FILE\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Printf("Opening dictionary\n")
	dict, err := dictionary.OpenFile(dictfile)
	if err != nil {
		fmt.Printf("Error opening dictionary: %v\n", err)
		os.Exit(1)
	}
	defer dict.Close()

	fmt.Printf("Verifying\n")
	err = dict.Verify()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done\n")
}
