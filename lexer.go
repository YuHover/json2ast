package jsonparser

import (
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
    loc location
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
func tokenize(source string) ([]jsonToken, []jsonError) {
	var runeSource = []rune(source)
	var cursor = 0
    var lineNum = 1
	var jts []jsonToken
    var jerrs []jsonError

	var r rune
	var token jsonToken
	var err error

	for cursor < len(runeSource) {
		r = runeSource[cursor]
		switch {
		case r == '{':
			jts = append(jts, jsonToken{ typ: LeftBrace, val: "{", loc: location{lineNum, cursor } })
			cursor++
		case r == '}':
			jts = append(jts, jsonToken{ typ: RightBrace, val: "}", loc: location{lineNum, cursor } })
			cursor++
		case r == '[':
			jts = append(jts, jsonToken{ typ: LeftBracket, val: "[", loc: location{lineNum, cursor } })
			cursor++
		case r == ']':
			jts = append(jts, jsonToken{ typ: RightBracket, val: "]", loc: location{lineNum, cursor } })
			cursor++
		case r == ':':
			jts = append(jts, jsonToken{ typ: Colon, val: ":", loc: location{lineNum, cursor } })
			cursor++
		case r == ',':
			jts = append(jts, jsonToken{ typ: Comma, val: ",", loc: location{lineNum, cursor } })
			cursor++
		case r == 't' || r == 'f':
			token, cursor, lineNum, err = tokenizeBoolean(runeSource, cursor, lineNum)
			if err != nil { jerrs = append(jerrs, err.(jsonError)) }
            jts = append(jts, token)
		case r == 'n':
            token, cursor, lineNum, err = tokenizeNull(runeSource, cursor, lineNum)
            if err != nil { jerrs = append(jerrs, err.(jsonError)) }
            jts = append(jts, token)
		case r == '"':
            token, cursor, lineNum, err = tokenizeString(runeSource, cursor, lineNum)
            if err != nil { jerrs = append(jerrs, err.(jsonError)) }
            jts = append(jts, token)
		case unicode.IsDigit(r) || r == '+' || r == '-':
            token, cursor, lineNum, err = tokenizeNumber(runeSource, cursor, lineNum)
            if err != nil { jerrs = append(jerrs, err.(jsonError)) }
            jts = append(jts, token)
		case whiteSpace[r]:
            if r == '\n' { lineNum++ }
            cursor++
		default:
			var i = cursor
			for ; i < len(runeSource) && !delimiters[r]; i++ { r = runeSource[i] }
			jerrs = append(jerrs, jsonError{ typ: InvalidToken,  loc: location{lineNum, cursor } })
			cursor = i
		}
	}

	return jts, jerrs
}

func tokenizeBoolean(source []rune, start int, lineNum int) (jsonToken, int, int, error) {
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
    var line = lineNum

	for ; cursor < len(source); cursor++ {
		r = source[cursor]

        if r == '\n' { line++ }
		if delimiters[r] { break }

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
    var token jsonToken
    if stage != Legal {
        err = jsonError{ typ: InvalidToken, loc: location{ lineNum, start } }
        return token, cursor, line, err
    }
    token = makeToken(Boolean, string(source[start:cursor]), lineNum, start)
	return token, cursor, line, err
}

func tokenizeNull(source []rune, start int, lineNum int) (jsonToken, int, int, error) {
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
    var line = lineNum

	for ; cursor < len(source); cursor++ {
		r = source[cursor]

        if r == '\n' { line++ }
		if delimiters[r] { break }

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
    var token jsonToken
	if stage != Legal {
		err = jsonError{ typ: InvalidToken, loc: location{ lineNum, start } }
        return token, cursor, line, err
	}
    token = makeToken(Null, string(source[start:cursor]), lineNum, start)
	return token, cursor, line, err
}

func tokenizeNumber(source []rune, start int, lineNum int) (jsonToken, int, int, error) {
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
    var line = lineNum

	for ; cursor < len(source); cursor++ {
		r = source[cursor]

        if r == '\n' { line++ }
        if delimiters[r] { break}

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
    var token jsonToken
    if stage != Zero && stage != Integer && stage != Frac && stage != Exp {
        err = jsonError{ typ: InvalidToken, loc: location{ lineNum,start } }
        return token, cursor, line, err
    }
    token = makeToken(Number, string(source[start:cursor]), lineNum, start)
	return token, cursor, line, err
}

func tokenizeString(source []rune, start int, lineNum int) (jsonToken, int, int, error) {
	const (
		Initial 	State = iota
		Open
		Escape
		Unicode
		Hex
		HexHex
		HexHexHex
        Acc
		Panic
	)

	var r rune
	var stage = Initial
	var cursor = start
    var line = lineNum
    var isClose = false
    var before State

	for ; cursor < len(source); cursor++ {
        if stage != Panic { before = stage }

		r = source[cursor]
        if r == '\n' { line++ }
		if r == '"' && (stage != Initial && stage != Escape) {
            isClose = true
            if stage == Open { stage = Acc }
			break
		}

		switch stage {
		case Initial: 	if r == '"' { stage = Open } else { stage =  Panic }
		case Unicode: 	if isHex(r) { stage = Hex } else { stage =  Panic }
		case Hex: 		if isHex(r) { stage = HexHex } else { stage = Panic }
		case HexHex: 	if isHex(r) { stage = HexHexHex } else { stage = Panic }
		case HexHexHex: if isHex(r) { stage = Open; continue } else { stage = Panic }
		case Open:
			if r == '\\' { stage = Escape; continue }
			if r >= 0x0020 && r <= 0x10FFF && !unicode.IsControl(r) { stage = Open; continue }
			stage = Panic
		case Escape:
			if r == 'u' { stage = Unicode; continue }
			if isEscapable(r) { stage = Open; continue }
            stage = Panic
		case Panic:
			if r == '\\' { cursor++ } // skip next rune
			stage = Panic
		}
	}

	if cursor++; cursor >= len(source) {
		cursor = len(source)
	}

	var err error
    var token jsonToken
    if stage != Acc {
        var errTyp errorType
        if !isClose {
            errTyp = MissCloseQuote
        } else if before == Unicode || before == Hex || before == HexHex || before == HexHexHex {
            errTyp = InvalidUnicode
        } else if before == Escape {
            errTyp = InvalidEscape
        } else if before == Open {
            errTyp = InvalidChar
        }

        err = jsonError{ typ: errTyp, loc: location{ lineNum, start } }
        return token, cursor, line, err
	}
    token = makeToken(String, string(source[start:cursor]), lineNum, start)
	return token, cursor, line, err
}

func makeToken(typ tokenType, value string, lineNum int, pos int) jsonToken {
    return jsonToken{
        typ: typ,
        val: value,
        loc: location{ lineNum, pos },
    }
}

func isHex(r rune) bool {
	return unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f'
}

func isEscapable(r rune) bool {
	return r == '\\' || r == 'b' || r == 'f' || r == 'n' || r == 'r' || r == 't' || r == '"' || r == 'u'
}