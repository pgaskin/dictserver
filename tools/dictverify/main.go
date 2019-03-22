package main

import (
	"fmt"
	"os"

	"github.com/geek1011/dictserver/dictionary"
)

func main() {
	var idx, db string
	switch len(os.Args) {
	case 2:
		idx = os.Args[1] + ".idx"
		db = os.Args[1] + ".db"
	case 3:
		idx = os.Args[1]
		db = os.Args[2]
	default:
		fmt.Printf("Usage: %s IDX DB\n", os.Args[0])
		fmt.Printf("   or: %s DB_BASE\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Printf("Opening dictionary\n")
	dict, err := dictionary.OpenFile(idx, db)
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
