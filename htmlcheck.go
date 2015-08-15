package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type IterableTokenizer struct {
	*html.Tokenizer
}

func newIterableTokenizer(r io.Reader) IterableTokenizer {
	return IterableTokenizer{
		Tokenizer: html.NewTokenizer(r),
	}
}

func (t *IterableTokenizer) NextToken() html.Token {
	t.Next()
	token := t.Token()

	return token
}

func main() {
	url := getUrlFromCommandLine()

	fmt.Printf("Retrieving '%v'...\n", url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Couldn't get url: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Got response '%v'\n", response.Status)
	fmt.Println("Parsing html...")

	pageIds := make(map[string]int)
	tokenizer := newIterableTokenizer(response.Body)

	for token := tokenizer.NextToken(); token.Type != html.ErrorToken; token = tokenizer.NextToken() {
		switch token.Type {
		case html.StartTagToken, html.SelfClosingTagToken:

			for _, attribute := range token.Attr {
				if strings.ToLower(attribute.Key) == "id" {
					currentCount, exists := pageIds[attribute.Val]
					if !exists {
						pageIds[attribute.Val] = 1
					} else {
						pageIds[attribute.Val] = currentCount + 1
					}
				}
			}
		}
	}

	numTotalIds := len(pageIds)
	fmt.Printf("Found %v ids on page.\n", numTotalIds)

	foundDuplicates := false

	for id, count := range pageIds {
		if count > 1 {
			foundDuplicates = true
			fmt.Printf("\tWarning: %v shows up %v times.\n", id, count)
		}
	}

	if !foundDuplicates {
		fmt.Println("No duplicates ids found.")
	}
}

func getUrlFromCommandLine() (url string) {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: htmlcheck 'http://example.com'\nWrap the url in single quotes to be safe from special shell characters.")
		os.Exit(2)
	}

	return flag.Arg(0)
}
