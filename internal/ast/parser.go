package ast

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var escapableChars = strings.Split(`\~ \! \@ \# \$ \; \% \^ \: \& \? \* \( \) \- \_ \+ \= \{ \} \[ \] \' \" \\ \/ \. \, \| \< \>`, " ")

var lexdef = lexer.MustStateful(lexer.Rules{
	"Root": {
		lexer.Include("SpecChar"),
		{"Alternative", `\|`, nil},
		{"Quantifier", `{\d+}|{\d+,}|{\d+,\d+}|\?|\*|\+`, nil},
		{"Char", `[^\$\^\?\*\(\)\+\[\\\|]`, nil},
		{"Escaped", `\\`, lexer.Push("Escaped")},
		{"StartOfString", `\^|\\A`, nil},
		{"EndOfString", `\$`, nil},
		{"CharClassBeginExcept", `\[\^`, lexer.Push("CharClass")},
		{"CharClassBegin", `\[`, lexer.Push("CharClass")},
		{"GroupBegin", `\(`, lexer.Push("Group")},
	},
	"SpecChar": {
		{"CommonToken", `\\n|\\r|\\t`, nil},
		{"Anchor", `\\z|\\b|\\B`, nil},
		{"MetaChar", `\.|\\s|\\S|\\d|\\D|\\w|\\W|\\v`, nil},
	},
	"Escaped": {
		{"EscapedChar", "[" + strings.Join(escapableChars, "") + `\ ` + "]", lexer.Pop()},
	},
	"CharClass": {
		{"CharClassRange", `.-\\.|\\.-\\.|.-[^\]]|\\.-[^\]]`, nil},
		{"CharClassPresetBegin", `\[:`, lexer.Push("CharClassPreset")},
		{"CharClassEnd", `\]`, lexer.Pop()},
		lexer.Include("SpecChar"),
		{"CharClassEscaped", `\\`, lexer.Push("CharClassEscaped")},
		{"CharClassSingleChar", `\\.|[^\\]`, nil},
	},
	"CharClassEscaped": {
		{"CharClassEscapedChar", "[" + strings.Join(escapableChars, "") + `\ ` + "]", lexer.Pop()},
	},
	"CharClassPreset": {
		{"CharClassPreset", `alnum|alpha|ascii|blank|cntrl|digit|graph|lower|print|punct|space|upper|word|xdigit`, nil},
		{"CharClassPresetEnd", `:\]`, lexer.Pop()},
	},
	"Group": {
		{"GroupName", `\?P?<\w+>`, nil},
		{"GroupNonCapturing", `\?:`, nil},
		lexer.Include("Root"),
		{"GroupEnd", `\)`, lexer.Pop()},
	},
})

var parser = participle.MustBuild[Regexp](
	participle.Union[Expr](
		Group{},
		CharClass{},
		Char{},
		MetaChar{},
	),
	participle.Union[CharClassKind](
		CharClassSingleChar{},
		CharClassRange{},
		CharClassPreset{},
		CharClassMetaChar{},
	),
	participle.Lexer(lexdef))

func ParseString(re string) (*Regexp, error) {
	return parser.ParseString("", re)
}
