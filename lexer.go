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
	Typ tokenType
	Val string
    Loc location
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

var fm = map[rune]func(*context)(jsonToken, error) {
    't': tokenizeBoolean,
    'f': tokenizeBoolean,
    'n': tokenizeNull,
    '"': tokenizeString,
    '+': tokenizeNumber,
    '-': tokenizeNumber,
}

var tm = map[rune] tokenType {
    '{': LeftBrace,
    '}': RightBrace,
    '[': LeftBracket,
    ']': RightBracket,
    ':': Colon,
    ',': Comma,
}

type context struct {
    rs      []rune
    cursor  int
    lineNum int
    colNum  int
}

func init() {
    for r := '0'; r <= '9'; r++ {
        fm[r] = tokenizeNumber
    }
}

// 将json字符串解析成token流
func tokenize(source string) ([]jsonToken, []jsonError) {
	var ctx = context {
        rs:      []rune(source),
        cursor:  0,
        lineNum: 1,
        colNum:  1,
    }

	var jts []jsonToken
    var jerrs []jsonError

	var r rune
	var token jsonToken
	var err error

    var doTokenizeSingle = func(r rune) {
        token = jsonToken{ tm[r], string(r), location{ctx.lineNum, ctx.colNum } }
        jts = append(jts, token)
        ctx.cursor++
        ctx.colNum++
    }
    
    var doTokenizeCall = func(r rune) {
        token, err = fm[r](&ctx)
        if err != nil { jerrs = append(jerrs, err.(jsonError)) }
        jts = append(jts, token)
    }

	for ctx.cursor < len(ctx.rs) {
        r = ctx.rs[ctx.cursor]
		switch {
		case r == '{': fallthrough
		case r == '}': fallthrough
		case r == '[': fallthrough
		case r == ']': fallthrough
		case r == ':': fallthrough
		case r == ',': doTokenizeSingle(r)

        case r == 't' || r == 'f':  fallthrough
		case r == 'n':              fallthrough
		case r == '"':              fallthrough
		case unicode.IsDigit(r) || r == '+' || r == '-': doTokenizeCall(r)

        case whiteSpace[r]:
            ctx.cursor++
            ctx.colNum++
            if r == '\n' {
                ctx.lineNum++
                ctx.colNum = 1
            }
		default:
			var col = ctx.colNum
			for ; ctx.cursor < len(ctx.rs) && !delimiters[r]; ctx.cursor++ {
                r = ctx.rs[ctx.cursor]
                ctx.colNum++
            }
            err = jsonError{InvalidToken, location {ctx.lineNum,col } }
			jerrs = append(jerrs, err.(jsonError))
		}
	}

	return jts, jerrs
}

func tokenizeBoolean(ctx *context) (jsonToken, error) {
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
	var start = ctx.cursor
    var col = ctx.colNum

	for ; ctx.cursor < len(ctx.rs); ctx.cursor++ {
		r = ctx.rs[ctx.cursor]
		if delimiters[r] { break }
        ctx.colNum++

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

    var err error
    var token jsonToken
    if stage != Legal {
        err = jsonError{ InvalidToken, location{ ctx.lineNum, col } }
        return token, err
    }
    token = jsonToken {
        Typ: Boolean,
        Val: string(ctx.rs[start:ctx.cursor]),
        Loc: location{ctx.lineNum, col},
    }
	return token, err
}

func tokenizeNull(ctx *context) (jsonToken, error) {
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
	var start = ctx.cursor
    var col = ctx.colNum

	for ; ctx.cursor < len(ctx.rs); ctx.cursor++ {
		r = ctx.rs[ctx.cursor]
		if delimiters[r] { break }
        ctx.colNum++

		switch stage {
		case Initial:	if r == 'n' { stage = N } else { stage = Panic }
		case N: 		if r == 'u' { stage = Nu } else { stage = Panic }
		case Nu: 		if r == 'l' { stage = Nul } else { stage = Panic }
		case Nul:		if r == 'l' { stage = Legal } else { stage = Panic }
		case Legal: 	stage = Panic
		case Panic: 	stage = Panic
		}
	}

	var err error
    var token jsonToken
	if stage != Legal {
		err = jsonError{ InvalidToken, location{ ctx.lineNum, col } }
        return token, err
	}
    token = jsonToken {
        Typ: Null,
        Val: string(ctx.rs[start:ctx.cursor]),
        Loc: location{ctx.lineNum, col},
    }
	return token, err
}

func tokenizeNumber(ctx *context) (jsonToken, error) {
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
	var start = ctx.cursor
    var col = ctx.colNum

	for ; ctx.cursor < len(ctx.rs); ctx.cursor++ {
		r = ctx.rs[ctx.cursor]
        if delimiters[r] { break }
        ctx.colNum++

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
        err = jsonError{ InvalidToken, location{ ctx.lineNum,col } }
        return token, err
    }
    token = jsonToken {
        Typ: Number,
        Val: string(ctx.rs[start:ctx.cursor]),
        Loc: location{ctx.lineNum, col},
    }
	return token, err
}

func tokenizeString(ctx *context) (jsonToken, error) {
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
    var start = ctx.cursor
    var line = ctx.lineNum
    var col = ctx.colNum
    var isClose = false
    var before State

	for ; ctx.cursor < len(ctx.rs); ctx.cursor++ {
        if stage != Panic { before = stage }

        ctx.colNum++
        r = ctx.rs[ctx.cursor]

        if r == '\n' {
            ctx.lineNum++
            ctx.colNum = 1
        }

		if r == '"' && (stage != Initial && stage != Escape) {
            isClose = true
            if stage == Open { stage = Acc }
            ctx.cursor++
			break
		}

		switch stage {
		case Initial: 	if r == '"' { stage = Open } else { stage =  Panic }
		case Unicode: 	if isHex(r) { stage = Hex } else { stage =  Panic }
		case Hex: 		if isHex(r) { stage = HexHex } else { stage = Panic }
		case HexHex: 	if isHex(r) { stage = HexHexHex } else { stage = Panic }
		case HexHexHex: if isHex(r) { stage = Open } else { stage = Panic }
		case Open:
			if r == '\\' { stage = Escape; continue }
			if r >= 0x0020 && r <= 0x10FFF && !unicode.IsControl(r) { stage = Open; continue }
			stage = Panic
		case Escape:
			if r == 'u' { stage = Unicode; continue }
			if isEscapable(r) { stage = Open; continue }
            stage = Panic
		case Panic:
			if r == '\\' { ctx.cursor++ } // skip next rune
			stage = Panic
		}
	}

	var err error
    var token jsonToken
    if stage != Acc {
        var errTyp ErrorType
        if !isClose {
            errTyp = MissCloseQuote
        } else if before == Unicode || before == Hex || before == HexHex || before == HexHexHex {
            errTyp = InvalidUnicode
        } else if before == Escape {
            errTyp = InvalidEscape
        } else if before == Open {
            errTyp = InvalidChar
        }

        err = jsonError{ errTyp, location{ line, col } }
        return token, err
	}
    token = jsonToken {
        Typ: String,
        Val: string(ctx.rs[start:ctx.cursor]),
        Loc: location{line, col},
    }
	return token, err
}

func isHex(r rune) bool {
	return unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f'
}

func isEscapable(r rune) bool {
	return r == '\\' || r == 'b' || r == 'f' || r == 'n' || r == 'r' || r == 't' || r == '"' || r == 'u'
}