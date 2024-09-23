package main

import (
	"fmt"

	wiki "github.com/trietmn/go-wiki"
)

func main() {
	results, suggestion, err := wiki.Search("Gemini", 3, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n%s\n", results, suggestion)
}
