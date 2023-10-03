package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/hikitani/regen/internal/ast"
	"github.com/hikitani/regen/internal/generator"
)

func main() {
	v, err := ast.ParseString(`(?P<foo>a?[bc]*)+[a-z[:alnum:]]{1,20}$|^\d+(\s|foo)`)
	if err != nil {
		panic(err)
	}

	spew.Config.Indent = "  "
	spew.Config.DisableCapacities = true
	spew.Config.DisableMethods = true
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisablePointerMethods = true
	spew.Config.SortKeys = true
	spew.Config.Dump(v)

	gen, err := generator.New(`[\.\@]`, generator.QuantifierConstraint{MaxLen: 10})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated string: %s\n", gen())
}
