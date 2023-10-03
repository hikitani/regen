package generator

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/hikitani/regen/internal/ast"
)

func New(re string, constraint QuantifierConstraint) (func() string, error) {
	root, err := ast.ParseString(re)
	if err != nil {
		return nil, fmt.Errorf("parse regex: %w", err)
	}

	var v validator
	ast.Walk(&v, root)
	if v.err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	return func() string {
		return genFromNode(root, constraint)
	}, nil
}

type validator struct {
	err error
}

func (vd *validator) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch n := n.(type) {
	case ast.MetaChar:
		if strings.Index(`.\s\S\d\D\w\W\v`, n.Char) == -1 {
			vd.err = fmt.Errorf("meta char '%s' is invalid", n.Char)
			return nil
		}

		if err := quantifierValidate(n.Quantifier); err != nil {
			vd.err = fmt.Errorf("invalid quantifier: %w", err)
			return nil
		}

		if n.Char == `\v` {
			vd.err = errors.New("unimplemented meta char '\\v'")
			return nil
		}
	case ast.Char:
		if err := quantifierValidate(n.Quantifier); err != nil {
			vd.err = fmt.Errorf("invalid quantifier: %w", err)
			return nil
		}
	case ast.CharClass:
		for _, class := range n.Chars {
			switch class := class.(type) {
			case ast.CharClassRange:
				if err := charClassRangeValidate(class); err != nil {
					vd.err = fmt.Errorf("invalid char class range: %w", err)
					return nil
				}
			case ast.CharClassMetaChar:
				if strings.Index(`.\s\S\d\D\w\W\v`, class.Char) == -1 {
					vd.err = fmt.Errorf("char class meta char '%s' is invalid", class.Char)
					return nil
				}

				if class.Char == `\v` {
					vd.err = errors.New("unimplemented char class meta char '\\v'")
					return nil
				}
			case ast.CharClassPreset:
				vd.err = errors.New("unimplemented char class presets (e.g. [:alnum:])")
				return nil
			}
		}
	}

	return vd
}

func quantifierValidate(q *ast.Quantifier) error {
	if q == nil {
		return nil
	}

	if q.To != nil {
		if *q.To < q.From {
			return fmt.Errorf("invalid quantifier range, expected from >= to, got (from=%d, to=%d)", q.From, *q.To)
		}
	}

	return nil
}

func charClassRangeValidate(class ast.CharClassRange) error {
	from, to, ok := strings.Cut(class.Range, "-")
	if !ok {
		panic(fmt.Sprintf("charClassRangeValidate: invalid class range, got %s", class.Range))
	}

	if len(from) != 1 || len(to) != 1 {
		panic(fmt.Sprintf("charClassRangeValidate: invalid class range, got %s", class.Range))
	}

	if to[0] < from[0] {
		return fmt.Errorf("expected to >= from, got (from='%s', to='%s')", string(from), string(to))
	}

	return nil
}

type QuantifierConstraint struct {
	MaxLen int
}

func genFromNode(v ast.Node, constraint QuantifierConstraint) string {
	var sb strings.Builder
	var chars string
	switch n := v.(type) {
	case *ast.Regexp:
		chars = genFromRoot(n, constraint)
	case ast.Char:
		chars = genChar(n, constraint)
	case ast.MetaChar:
		chars = genFromMetaChar(n, constraint)
	case ast.CharClass:
		chars = genFromCharClass(n, constraint)
	case *ast.Alternative:
		chars = genFromalternative(n, constraint)
	case ast.Group:
		chars = genFromGroup(n, constraint)
	}

	sb.WriteString(chars)
	return sb.String()
}

func getSeqLen(q *ast.Quantifier, constraint QuantifierConstraint) int {
	if q == nil {
		return 1
	}

	var maxLen int
	minLen := q.From
	if to := q.To; to != nil {
		maxLen = max(minLen, *to)
	} else {
		maxLen = constraint.MaxLen
	}

	return rand.Intn(maxLen-minLen+1) + minLen
}

func genChar(cn ast.Char, contraints QuantifierConstraint) string {
	return strings.Repeat(cn.Char, getSeqLen(cn.Quantifier, contraints))
}

func genFromMetaChar(mcn ast.MetaChar, constraint QuantifierConstraint) string {
	nchars := getSeqLen(mcn.Quantifier, constraint)
	switch mcn.Char {
	case ".":
		return makeRandCharsFromSeq(nchars, anyChars)
	case `\s`:
		return makeRandCharsFromSeq(nchars, spaceChars)
	case `\S`:
		return makeRandCharsFromSeq(nchars, nonSpaceChars)
	case `\d`:
		return makeRandCharsFromSeq(nchars, digitChars)
	case `\D`:
		return makeRandCharsFromSeq(nchars, nonDigitChars)
	case `\w`:
		return makeRandCharsFromSeq(nchars, wordChars)
	case `\W`:
		return makeRandCharsFromSeq(nchars, nonWordChars)
	case `\v`:
		panic("genFromMetaChar: unimplemented \\v")
	}

	panic(fmt.Sprintf("genFromMetaChar: unexpected char '%s'", mcn.Char))
}

func genFromCharClass(ccn ast.CharClass, constraint QuantifierConstraint) string {
	nchars := getSeqLen(ccn.Quantifier, constraint)
	var charSeqSB strings.Builder
	for _, class := range ccn.Chars {
		switch class := class.(type) {
		case ast.CharClassSingleChar:
			charSeqSB.WriteString(class.Char)
		case ast.CharClassRange:
			from, to, ok := strings.Cut(class.Range, "-")
			if !ok {
				panic(fmt.Sprintf("genFromCharClass: invalid class range, got %s", class.Range))
			}

			if len(from) != 1 || len(to) != 1 {
				panic(fmt.Sprintf("genFromCharClass: invalid class range, got %s", class.Range))
			}

			charSeqSB.WriteString(charsFromCharClassRange(from[0], to[0]))
		case ast.CharClassMetaChar:
			switch class.Char {
			case ".":
				charSeqSB.WriteString(anyChars)
			case `\s`:
				charSeqSB.WriteString(spaceChars)
			case `\S`:
				charSeqSB.WriteString(nonSpaceChars)
			case `\d`:
				charSeqSB.WriteString(digitChars)
			case `\D`:
				charSeqSB.WriteString(nonDigitChars)
			case `\w`:
				charSeqSB.WriteString(wordChars)
			case `\W`:
				charSeqSB.WriteString(nonWordChars)
			case `\v`:
				panic("genFromCharClass: unimplemented \\v")
			}
		case ast.CharClassPreset:
			panic("genFromCharClass: unimplemented CharClassPreset")
		}
	}

	resultSeq := charUniq(charSeqSB.String())

	var chars string
	if ccn.Expect {
		chars = charsWithExpect(anyChars, resultSeq)
	} else {
		chars = resultSeq
	}

	return makeRandCharsFromSeq(nchars, chars)
}

func genFromalternative(n *ast.Alternative, constraint QuantifierConstraint) string {
	charsArr := make([]string, 0, len(n.Exprs))
	for _, expr := range n.Exprs {
		if expr, ok := expr.(ast.Node); !ok {
			panic(fmt.Sprintf("genFromConstraint: expr is not Node"))
		} else {
			charsArr = append(charsArr, genFromNode(expr, constraint))
		}
	}
	return strings.Join(charsArr, "")
}

func genFromGroup(n ast.Group, constraint QuantifierConstraint) string {
	selectedAlternative := rand.Intn(len(n.Alternatives))
	chars := genFromNode(n.Alternatives[selectedAlternative], constraint)
	return strings.Repeat(chars, getSeqLen(n.Quantifier, constraint))
}

func genFromRoot(n *ast.Regexp, constraint QuantifierConstraint) string {
	selectedAlternative := rand.Intn(len(n.Alternatives))
	return genFromNode(n.Alternatives[selectedAlternative], constraint)
}

func charUniq(chars string) string {
	uniqChars := map[byte]struct{}{}
	for _, b := range []byte(chars) {
		uniqChars[b] = struct{}{}
	}

	var sb strings.Builder
	sb.Grow(len(uniqChars))
	for b, _ := range uniqChars {
		sb.WriteByte(b)
	}
	return sb.String()
}

func charsWithExpect(targetChars, expectChars string) string {
	var sb strings.Builder
	for _, ch := range []byte(targetChars) {
		if strings.IndexByte(expectChars, ch) == -1 {
			sb.WriteByte(ch)
		}
	}

	return sb.String()
}

func charsFromCharClassRange(from, to byte) string {
	if to < from {
		panic(fmt.Sprintf("charsFromCharClassRange: to < from (from='%s', to='%s')", string(from), string(to)))
	}

	var sb strings.Builder
	for i := from; i <= to; i++ {
		sb.WriteByte(i)
	}
	return sb.String()
}

const (
	anyChars      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ \t`~!@\"#№$;%^:&?*()-_=+'<>,.1234567890"
	wordChars     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_"
	digitChars    = "1234567890"
	spaceChars    = " \t\n"
	nonWordChars  = " \t\n`~!@\"#№$;%^:&?*()-=+'<>,."
	nonDigitChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ \t\n`~!@\"#№$;%^:&?*()-_=+'<>,."
	nonSpaceChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`~!@\"#№$;%^:&?*()-_=+'<>,.1234567890"
)

func makeRandCharsFromSeq(sz int, letters string) string {
	var sb strings.Builder
	sb.Grow(sz)
	for i := 0; i < sz; i++ {
		sb.WriteByte(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}
