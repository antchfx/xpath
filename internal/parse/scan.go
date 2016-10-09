package parse

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

type itemType int

const (
	itemComma      itemType = iota // ','
	itemSlash                      // '/'
	itemAt                         // '@'
	itemDot                        // '.'
	itemLParens                    // '('
	itemRParens                    // ')'
	itemLBracket                   // '['
	itemRBracket                   // ']'
	itemStar                       // '*'
	itemPlus                       // '+'
	itemMinus                      // '-'
	itemEq                         // '='
	itemLt                         // '<'
	itemGt                         // '>'
	itemBang                       // '!'
	itemDollar                     // '$'
	itemApos                       // '\''
	itemQuote                      // '"'
	itemUnion                      // '|'
	itemNe                         // '!='
	itemLe                         // '<='
	itemGe                         // '>='
	itemAnd                        // '&&'
	itemOr                         // '||'
	itemDotDot                     // '..'
	itemSlashSlash                 // '//'
	itemName                       // XML Name
	itemString                     // Quoted string constant
	itemNumber                     // Number constant
	itemAxe                        // Axe (like child::)
	itemEof                        // END
)

type scanner struct {
	text, name, prefix string

	pos       int
	curr      rune
	typ       itemType
	strval    string  // text value at current pos
	numval    float64 // number value at current pos
	canBeFunc bool
}

func (s *scanner) nextChar() bool {
	if s.pos >= len(s.text) {
		s.curr = rune(0)
		return false
	}
	s.curr = rune(s.text[s.pos])
	s.pos += 1
	return true
}

func (s *scanner) nextItem() bool {
	s.skipSpace()
	switch s.curr {
	case 0:
		s.typ = itemEof
		return false
	case ',', '@', '(', ')', '|', '*', '[', ']', '+', '-', '=', '#', '$':
		s.typ = asItemType(s.curr)
		s.nextChar()
	case '<':
		s.typ = itemLt
		s.nextChar()
		if s.curr == '=' {
			s.typ = itemLe
			s.nextChar()
		}
	case '>':
		s.typ = itemGt
		s.nextChar()
		if s.curr == '=' {
			s.typ = itemGe
			s.nextChar()
		}
	case '!':
		s.typ = itemBang
		s.nextChar()
		if s.curr == '=' {
			s.typ = itemNe
			s.nextChar()
		}
	case '.':
		s.typ = itemDot
		s.nextChar()
		if s.curr == '.' {
			s.typ = itemDotDot
			s.nextChar()
		} else if isDigit(s.curr) {
			s.typ = itemNumber
			s.numval = s.scanFraction()
		}
	case '/':
		s.typ = itemSlash
		s.nextChar()
		if s.curr == '/' {
			s.typ = itemSlashSlash
			s.nextChar()
		}
	case '"', '\'':
		s.typ = itemString
		s.strval = s.scanString()
	default:
		if isDigit(s.curr) {
			s.typ = itemNumber
			s.numval = s.scanNumber()
		} else if isName(s.curr) {
			s.typ = itemName
			s.name = s.scanName()
			s.prefix = ""
			// "foo:bar" is one itemem not three because it doesn't allow spaces in between
			// We should distinct it from "foo::" and need process "foo ::" as well
			if s.curr == ':' {
				s.nextChar()
				// can be "foo:bar" or "foo::"
				if s.curr == ':' {
					// "foo::"
					s.nextChar()
					s.typ = itemAxe
				} else { // "foo:*", "foo:bar" or "foo: "
					s.prefix = s.name
					if s.curr == '*' {
						s.nextChar()
						s.name = "*"
					} else if isName(s.curr) {
						s.name = s.scanName()
					} else {
						panic(fmt.Sprintf("%s has an invalid qualified name.", s.text))
					}
				}
			} else {
				s.skipSpace()
				if s.curr == ':' {
					s.nextChar()
					// it can be "foo ::" or just "foo :"
					if s.curr == ':' {
						s.nextChar()
						s.typ = itemAxe
					} else {
						panic(fmt.Sprintf("%s has an invalid qualified name.", s.text))
					}
				}
			}
			s.skipSpace()
			s.canBeFunc = s.curr == '('
		} else {
			panic(fmt.Sprintf("%s has an invalid token.", s.text))
		}
	}
	return true
}

func (s *scanner) skipSpace() {
Loop:
	for {
		if !unicode.IsSpace(s.curr) || !s.nextChar() {
			break Loop
		}
	}
}

func (s *scanner) scanFraction() float64 {
	var (
		i = s.pos - 2
		c = 1 // '.'
	)
	for isDigit(s.curr) {
		s.nextChar()
		c++
	}
	v, err := strconv.ParseFloat(s.text[i:i+c], 64)
	if err != nil {
		panic(fmt.Errorf("xpath: scanFraction parse float got error: %v", err))
	}
	return v
}

func (s *scanner) scanNumber() float64 {
	var (
		c int
		i = s.pos - 1
	)
	for isDigit(s.curr) {
		s.nextChar()
		c++
	}
	if s.curr == '.' {
		s.nextChar()
		c++
		for isDigit(s.curr) {
			s.nextChar()
			c++
		}
	}
	v, err := strconv.ParseFloat(s.text[i:i+c], 64)
	if err != nil {
		panic(fmt.Errorf("xpath: scanNumber parse float got error: %v", err))
	}
	return v
}

func (s *scanner) scanString() string {
	var (
		c   = 0
		end = s.curr
	)
	s.nextChar()
	i := s.pos - 1
	for s.curr != end {
		if !s.nextChar() {
			panic(errors.New("xpath: scanString got unclosed string"))
		}
		c++
	}
	s.nextChar()
	return s.text[i : i+c]
}

func (s *scanner) scanName() string {
	var (
		c int
		i = s.pos - 1
	)
	for isName(s.curr) {
		c++
		if !s.nextChar() {
			break
		}
	}
	return s.text[i : i+c]
}

func isName(r rune) bool {
	return string(r) != ":" && string(r) != "/" &&
		(unicode.Is(first, r) || unicode.Is(second, r) || string(r) == "*")
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func asItemType(r rune) itemType {
	switch r {
	case ',':
		return itemComma
	case '@':
		return itemAt
	case '(':
		return itemLParens
	case ')':
		return itemRParens
	case '|':
		return itemUnion
	case '*':
		return itemStar
	case '[':
		return itemLBracket
	case ']':
		return itemRBracket
	case '+':
		return itemPlus
	case '-':
		return itemMinus
	case '=':
		return itemEq
	case '$':
		return itemDollar
	}
	panic(fmt.Errorf("unknown item: %v", r))
}
