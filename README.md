# regen

Simple tool for generating sequences that match the entered regular expression

## Installation

Install the latest version with:

```bash
$ go install github.com/hikitani/regen@latest
```

## Usage

```bash
$ regen --help
regen - tool for generating sequences that match the entered regular expression
Usage: regen [OPTION]... [REGEX]

Options:
  -format string
        Set the output format of the generated data (plain or json) (default "plain")
  -num int
        Set the number of generated rows that match the regex (default 1)
  -quantifier-upper-bound int
        Set an upper bound for all quantifiers +, * and {n,} (default 10)

Example:
1. Generating three rows:
        $ regen -num 3 'cats?|(dog){2,3}'
        Row: 0; Text: dogdogdog
        Row: 1; Text: cats
        Row: 2; Text: cat

2.JSON output:
        $ regen -num 3 -format json '[a-zA-Z]{3,6}-[0-9]{3}'
        [{"row":0,"text":"QAFL-571"},{"row":1,"text":"rqk-483"},{"row":2,"text":"lYyIx-217"}]
```

## Limitation

* unsupported some char presets
* unsupported some meta chars
* anchors are ignored
* the names of the captured groups are ignored
