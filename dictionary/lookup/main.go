package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/geek1011/dictserver/dictionary"
)

func main() {
	dict, err := dictionary.Load()
	if err != nil {
		panic(err)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)

	fmt.Printf("\nWords = %d\n", len(dict.Words))

	for _, word := range os.Args[1:] {
		w, ok := dict.Lookup(word)
		if !ok {
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
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
