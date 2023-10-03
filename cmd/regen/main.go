package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hikitani/regen/internal/generator"
)

type jsonResult struct {
	Row  int    `json:"row"`
	Text string `json:"text"`
}

func generatedTextToPlain(texts []string) string {
	var sb strings.Builder
	for i, text := range texts {
		sb.WriteString(fmt.Sprintf("Row: %d; Text: %s\n", i, text))
	}
	return sb.String()
}

func generatedTextToJson(texts []string) string {
	var jsonResults []jsonResult
	for i, text := range texts {
		jsonResults = append(jsonResults, jsonResult{
			Row:  i,
			Text: text,
		})
	}

	b, err := json.Marshal(&jsonResults)
	if err != nil {
		panic(fmt.Sprintf("generatedTextToJson: %s", err))
	}
	return string(append(b, '\n'))
}

func main() {
	log.SetFlags(log.Lmsgprefix)
	log.SetPrefix("regen: ")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "regen - tool for generating sequences that match the entered regular expression\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: regen [OPTION]... [REGEX]\n\n")
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExample:\n1. Generating three rows:\n\t$ regen -num 3 'cats?|(dog){2,3}'\n")

		fmt.Fprintln(flag.CommandLine.Output(), "\tRow: 0; Text: dogdogdog")
		fmt.Fprintln(flag.CommandLine.Output(), "\tRow: 1; Text: cats")
		fmt.Fprintln(flag.CommandLine.Output(), "\tRow: 2; Text: cat")

		fmt.Fprintf(flag.CommandLine.Output(), "\n2.JSON output:\n\t$ regen -num 3 -format json '[a-zA-Z]{3,6}-[0-9]{3}'\n")
		fmt.Fprintln(flag.CommandLine.Output(), `	[{"row":0,"text":"QAFL-571"},{"row":1,"text":"rqk-483"},{"row":2,"text":"lYyIx-217"}]`)
	}

	quantifierUpperBound := flag.Int("quantifier-upper-bound", 10, "Set an upper bound for all quantifiers +, * and {n,}")
	textNum := flag.Int("num", 1, "Set the number of generated rows that match the regex")
	outputType := flag.String("format", "plain", "Set the output format of the generated data (plain or json)")

	flag.Parse()

	if *quantifierUpperBound < 0 {
		log.Fatal("Invalid -quantifier-upper-bound: must be positive")
	}

	if *textNum < 1 {
		log.Fatal("Invalid -num: min is 1")
	}

	var formatter func([]string) string
	switch *outputType {
	case "plain":
		formatter = generatedTextToPlain
	case "json":
		formatter = generatedTextToJson
	default:
		log.Fatal("Invalid -format: type plain or json")
	}

	regex := flag.Arg(0)
	if regex == "" {
		log.Fatal("Enter a regular expression")
	}

	gen, err := generator.New(regex, generator.QuantifierConstraint{MaxLen: *quantifierUpperBound})
	if err != nil {
		log.Fatalf("Failed to create generator: %ss", err)
	}
	var texts []string
	for i := 0; i < *textNum; i++ {
		texts = append(texts, gen())
	}

	fmt.Fprint(os.Stdout, formatter(texts))
}
