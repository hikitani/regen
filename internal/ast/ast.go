package ast

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Regexp struct {
	Alternatives []*Alternative `(@@ (Alternative @@)?)?`
}

func (*Regexp) node() {}

type Expr interface{ value() }
type Node interface{ node() }

type Alternative struct {
	StartOfString bool   `@StartOfString?`
	Exprs         []Expr `@@*`
	EndOfString   bool   `@EndOfString?`
}

func (*Alternative) node() {}

type Group struct {
	Kind         *GroupKind     `GroupBegin @(GroupName | GroupNonCapturing)?`
	Alternatives []*Alternative `(@@ (Alternative @@)?)? GroupEnd`
	Quantifier   *Quantifier    `@Quantifier?`
}

func (Group) value() {}
func (Group) node()  {}

type GroupKind struct {
	Name         string
	NonCapturing bool
}

func (gk *GroupKind) Capture(values []string) error {
	v := values[0]
	if v == "?:" {
		gk.NonCapturing = true
		return nil
	}

	const capturedName = "GroupName"
	groupNameRe := regexp.MustCompile(`\?P?<(?P<` + capturedName + `>\w+)>`)
	groups := groupNameRe.FindStringSubmatch(v)
	if name := groups[groupNameRe.SubexpIndex(capturedName)]; name == "" {
		return fmt.Errorf("group '%s' not found for re `%s`", capturedName, groupNameRe)
	} else {
		gk.Name = name
		return nil
	}
}

type CharClass struct {
	Expect     CharClassExpect `@(CharClassBeginExcept | CharClassBegin)`
	Chars      []CharClassKind `@@+ CharClassEnd`
	Quantifier *Quantifier     `@Quantifier?`
}

func (CharClass) value() {}
func (CharClass) node()  {}

type CharClassExpect bool

func (b *CharClassExpect) Capture(values []string) error {
	if values[0] == "[^" {
		*b = true
	}
	return nil
}

type Char struct {
	Escaped    bool        `@Escaped?`
	Char       string      `@(EscapedChar | Char)`
	Quantifier *Quantifier `@Quantifier?`
}

func (Char) value() {}
func (Char) node()  {}

type MetaChar struct {
	Char       string      `@MetaChar`
	Quantifier *Quantifier `@Quantifier?`
}

func (MetaChar) value() {}
func (MetaChar) node()  {}

type Quantifier struct {
	From int
	To   *int
}

func (q *Quantifier) Capture(values []string) error {
	switch v := values[0]; v {
	case "":
		q.From = 1
		to := 1
		q.To = &to
	case "*":
	case "+":
		q.From = 1
	case "?":
		to := 1
		q.To = &to
	default:
		rangeWithoutBrackets := v[1 : len(v)-1]
		var from, to string
		var ok bool
		from, to, ok = strings.Cut(rangeWithoutBrackets, ",")
		if !ok {
			if _, err := strconv.Atoi(rangeWithoutBrackets); err != nil {
				return fmt.Errorf("invalid quantifier: got %s", v)
			} else {
				from, to = rangeWithoutBrackets, rangeWithoutBrackets
			}
		}

		if from, err := strconv.Atoi(from); err != nil {
			return fmt.Errorf("invalid 'from': %w", err)
		} else {
			q.From = from
		}

		if to == "" {
			break
		}

		if to, err := strconv.Atoi(to); err != nil {
			return fmt.Errorf("invalid 'to': %w", err)
		} else {
			q.To = &to
		}
	}

	return nil
}

type CharClassKind interface{ charClassKind() }

type CharClassSingleChar struct {
	Escaped bool   `@CharClassEscaped?`
	Char    string `@(CharClassEscapedChar | CharClassSingleChar)`
}

func (CharClassSingleChar) charClassKind() {}

type CharClassMetaChar struct {
	Char string `@MetaChar`
}

func (CharClassMetaChar) charClassKind() {}

type CharClassRange struct {
	Range string `@CharClassRange`
}

func (CharClassRange) charClassKind() {}

type CharClassPreset struct {
	Preset string `CharClassPresetBegin @CharClassPreset CharClassPresetEnd`
}

func (CharClassPreset) charClassKind() {}

type Visitor interface {
	Visit(v Node) Visitor
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *Regexp:
		for _, a := range n.Alternatives {
			Walk(v, a)
		}
	case *Alternative:
		for _, e := range n.Exprs {
			if e, ok := e.(Node); ok {
				Walk(v, e)
			}
		}
	case Group:
		for _, a := range n.Alternatives {
			Walk(v, a)
		}
	case CharClass:
	case Char:
	case MetaChar:
	default:
		panic(fmt.Sprintf("Walk: unexpected node type %T", n))
	}

	v.Visit(nil)
}
