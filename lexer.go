package jsonparser

import (
	"errors"
	// "fmt"
	"unicode"
)

type tokenType uint8

const (
	LeftBrace		tokenType = iota
	RightBrace
	LeftBracket
	RightBracket
	Comma
	Colon
	String
	Number
	Boolean
	Null
)

type jsonToken struct {
	typ tokenType
	val string
}

type State uint8 // status of DFA

// space in json
var whiteSpace = map[rune]bool{ ' ': true, '\t': true, '\r': true, '\n': true }
// delimiters
var delimiters = map[rune]bool {
	' ': true, '\t': true, '\r': true, '\n': true,
	'{':true, '}':true, '[': true, ']': true,
	'"': true, ',': true, ':': true,
}

// 将json字符串解析成token流
func tokenize(source string) []jsonToken {
	var runeSource = []rune(source)
	var cursor = 0
	var jts []jsonToken

	var r rune
	var token jsonToken
	var err error

	for cursor < len(runeSource) {
		r = runeSource[cursor]
		switch {
		case r == '{':
			jts = append(jts, jsonToken{ typ: LeftBrace, val: "{" })
			cursor++
		case r == '}':
			jts = append(jts, jsonToken{ typ: RightBrace, val: "}" })
			cursor++
		case r == '[':
			jts = append(jts, jsonToken{ typ: LeftBracket, val: "[" })
			cursor++
		case r == ']':
			jts = append(jts, jsonToken{ typ: RightBracket, val: "]" })
			cursor++
		case r == ':':
			jts = append(jts, jsonToken{ typ: Colon, val: ":" })
			cursor++
		case r == ',':
			jts = append(jts, jsonToken{ typ: Comma, val: "," })
			cursor++
		case r == 't' || r == 'f':
			token, cursor, err = tokenizeBoolean(runeSource, cursor)
			if err != nil {
				jts = append(jts, token)
			}
		case r == 'n':
			token, cursor, err = tokenizeNull(runeSource, cursor)
			if err != nil {
				jts = append(jts, token)
			}
		case r == '"':
			token, cursor, err = tokenizeString(runeSource, cursor)
			if err != nil {
				jts = append(jts, token)
			}
		case unicode.IsDigit(r) || r == '+' || r == '-':
			token, cursor, err = tokenizeNumber(runeSource, cursor)
			if err != nil {
				jts = append(jts, token)
			}
		case whiteSpace[r]: cursor++
		default:
			var i = cursor
			for ; i < len(runeSource) && !delimiters[r]; i++ { r = runeSource[i] }
			// TODO: build error runeSource[cursor:i]
			cursor = i
		}
	}

	return jts
}

func tokenizeBoolean(source []rune, start int) (jsonToken, int, error) {
	const (
		Initial State = iota
		T
		Tr
		Tru
		F
		Fa
		Fal
		Fals
		Legal
		Panic
	)

	var r rune
	var stage = Initial
	var cursor = start

	for ; cursor < len(source); cursor++ {
		r = source[cursor]
		if delimiters[r] {
			break
		}

		switch stage {
		case Initial:	if r == 't' { stage = T; continue }
						if r == 'f' { stage = F; continue }
						stage = Panic
		case T:			if r == 'r' { stage = Tr } else { stage = Panic }
		case Tr:		if r == 'u' { stage = Tru } else { stage = Panic }
		case Tru:		if r == 'e' { stage = Legal } else { stage = Panic }
		case F:			if r == 'a' { stage = Fa } else { stage = Panic }
		case Fa:		if r == 'l' { stage = Fal } else { stage = Panic }
		case Fal:		if r == 's' { stage = Fals } else { stage = Panic }
		case Fals:		if r == 'e' { stage = Legal } else { stage = Panic }
		case Legal:		stage = Panic
		case Panic:		stage = Panic
		}
	}

	// illegal: tru<stop>	trua<stop>	truea<stop>
	// legal:	true<stop>	false<stop>
	var err error
	if stage != Legal {
		err = errors.New(string(source[start:cursor])) // TODO: build error
	}
	
	var token = jsonToken {
		typ: Boolean,
		val: string(source[start:cursor]),
	}

	return token, cursor, err
}
func tokenizeNull(source []rune, start int) (jsonToken, int, error) {
	const (
		Initial State = iota
		N
		Nu
		Nul
		Legal
		Panic
	)

	var r rune
	var stage = Initial
	var cursor = start

	for ; cursor < len(source); cursor++ {
		r = source[cursor]

		if delimiters[r] {
			break
		}

		switch stage {
		case Initial:	if r == 'n' { stage = N } else { stage = Panic }
		case N: 		if r == 'u' { stage = Nu } else { stage = Panic }
		case Nu: 		if r == 'l' { stage = Nul } else { stage = Panic }
		case Nul:		if r == 'l' { stage = Legal } else { stage = Panic }
		case Legal: 	stage = Panic
		case Panic: 	stage = Panic
		}
	}

	// illegal: nu<stop>	nulb<stop>	nullb<stop>
	// legal:	null<stop>
	var err error
	if stage != Legal {
		err = errors.New(string(source[start:cursor])) // TODO: build error
	}

	var token = jsonToken {
		typ: Null,
		val: string(source[start:cursor]),
	}

	return token, cursor, err
}
func tokenizeNumber(source []rune, start int) (jsonToken, int, error) {
	const (
		Initial		State = iota
		Neg
		Zero
		Integer
		Dot
		Frac
		E
		ESign
		Exp
		Panic
	)

	var r rune
	var stage = Initial
	var cursor = start

	for ; cursor < len(source); cursor++ {
		r = source[cursor]
		if delimiters[r] {
			break
		}
		switch stage {
		case Initial:
			if r == '0' { stage = Zero; continue }
			if r == '-' { stage = Neg; continue }
			if unicode.IsDigit(r) { stage = Integer; continue }
			stage = Panic
		case Zero:
			if r == '.' { stage = Dot; continue }
			if r == 'e' || r == 'E' { stage = E; continue }
			stage = Panic
		case Neg:
			if r == '0' { stage = Zero; continue }
			if unicode.IsDigit(r) { stage = Integer; continue }
			stage = Panic
		case Integer:
			if unicode.IsDigit(r) { stage = Integer; continue }
			if r == '.' { stage = Dot; continue }
			if r == 'e' || r == 'E' { stage = E; continue }
			stage = Panic
		case Dot:
			if unicode.IsDigit(r) { stage = Frac; continue }
			stage = Panic
		case Frac:
			if unicode.IsDigit(r) { stage = Frac; continue }
			if r == 'e' || r == 'E' { stage = E; continue }
			stage = Panic
		case E:
			if r == '-' || r == '+' { stage = ESign; continue }
			if unicode.IsDigit(r) { stage = Exp; continue }
			stage = Panic
		case ESign:
			if unicode.IsDigit(r) { stage = Exp; continue }
			stage = Panic
		case Exp:
			if unicode.IsDigit(r) { stage = Exp; continue }
			stage = Panic
		case Panic: stage = Panic
		}
	}

	var err error
	if stage != Zero && stage != Integer && stage != Frac && stage != Exp {
		err = errors.New(string(source[start:cursor])) // TODO: build error
	}

	var token = jsonToken {
		typ: Number,
		val: string(source[start:cursor]),
	}

	return token, cursor, err
}
func tokenizeString(source []rune, start int) (jsonToken, int, error) {
	const (
		Initial 	State = iota
		Open
		Escape
		Unicode
		Hex
		HexHex
		HexHexHex
		Close
		Panic
	)

	var r rune
	var stage = Initial
	var cursor = start

	loop:
	for ; cursor < len(source); cursor++ {
		r = source[cursor]
		switch stage {
		case Initial: if r == '"' { stage = Open } else { stage = Panic }
		case Open:
			if r == '"' { stage = Close; break loop }
			if r == '\\' { stage = Escape; continue }
			if r >= 0x0020 && r <= 0x10FFF && !unicode.IsControl(r) { stage = Open; continue }
			stage = Panic
		case Escape:
			if r == '\\' || r == 'b' || r == 'f' || r == 'n' || r == 'r' || r == 't' || r == '"' {
				stage = Open
				continue
			}
			if r == 'u' { stage = Unicode; continue }
			stage = Panic
		case Unicode:
			if unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f' { stage = Hex; continue }
			if r == '"' { break loop }
			stage = Panic
		case Hex:
			if unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f' { stage = HexHex; continue }
			if r == '"' { break loop }
			stage = Panic
		case HexHex:
			if unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f' { stage = HexHexHex; continue }
			if r == '"' { break loop }
			stage = Panic
		case HexHexHex:
			if unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f' { stage = Open; continue }
			if r == '"' { break loop }
			stage = Panic
		case Panic:
			if r == '"' { break loop }
			if r == '\\' { cursor++ } // skip next rune
			stage = Panic
		}
	}

	if cursor++; cursor >= len(source) {
		cursor = len(source)
	}

	var err error
	if stage != Close {
		err = errors.New(string(source[start:cursor])) // TODO: build error
	}

	var token = jsonToken {
		typ: String,
		val: string(source[start:cursor]),
	}

	return token, cursor, err
}
