package main

import (
	"load-file-by-url/loader"
	"os"
)

func main() {
	fileURL := os.Args[1]

	fl := loader.NewFileLoader(100)
	items, err := fl.GetFileContent(fileURL)
	if err != nil {
		panic(err)
	}

	results := make(chan string)
	go doSomenthing(results, items)

	for result := range results {
		print(result)
	}
}

func doSomenthing(ch chan<- string, items []string) {
	for _, item := range items {
		ch <- item
	}
	close(ch)
}
