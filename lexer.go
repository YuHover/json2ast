package json2ast

import (
    "errors"
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

type dfsState uint8 // status of DFA

// space in json
var whiteSpace = map[rune]bool{ ' ': true, '\t': true, '\r': true, '\n': true }
// delimiters
var delimiters = map[rune]bool {
    ' ': true, '\t': true, '\r': true, '\n': true,
    '{':true, '}':true, '[': true, ']': true,
    '"': true, ',': true, ':': true,
}

var fm = map[rune]func(*context) (jsonToken, error) {
    't': tokenizeBoolean,
    'f': tokenizeBoolean,
    'n': tokenizeNull,
    '"': tokenizeString,
    '+': tokenizeNumber,
    '-': tokenizeNumber,
    '0': tokenizeNumber,
    '1': tokenizeNumber,
    '2': tokenizeNumber,
    '3': tokenizeNumber,
    '4': tokenizeNumber,
    '5': tokenizeNumber,
    '6': tokenizeNumber,
    '7': tokenizeNumber,
    '8': tokenizeNumber,
    '9': tokenizeNumber,
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

func getNextRune(ctx *context) (rune, error) {
    var r rune
    if ctx.cursor < len(ctx.rs) {
        r = ctx.rs[ctx.cursor]
        ctx.cursor++
        ctx.colNum++
        return r, nil
    }
    return r, errors.New("overstep")
}

func back(ctx *context) {
    ctx.cursor--
    ctx.colNum--
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

    var doTokenizeSingle = func(r rune) {
        token := jsonToken{ tm[r], string(r), location{ctx.lineNum, ctx.colNum - 1 } }
        jts = append(jts, token)
    }

    var doTokenizeCall = func(r rune) {
        token, jerr := fm[r](&ctx)
        if jerr != nil { jerrs = append(jerrs, jerr.(jsonError)) }
        jts = append(jts, token)
    }

    var r rune
    var err error
    for r, err = getNextRune(&ctx); err == nil; r, err = getNextRune(&ctx) {
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
        case unicode.IsDigit(r) || r == '+' || r == '-':
            back(&ctx) // back to the first rune of the token
            doTokenizeCall(r)

        case whiteSpace[r]:
            if r == '\n' {
                ctx.lineNum++
                ctx.colNum = 1
            }
        default:
            var col = ctx.colNum - 1
            for r, err = getNextRune(&ctx); err == nil && !delimiters[r]; r, err = getNextRune(&ctx) { }
            jerr := jsonError{InvalidToken, location {ctx.lineNum,col } }
            jerrs = append(jerrs, jerr)
            if err == nil { back(&ctx) } // back to the delimiter
        }
    }

    return jts, jerrs
}

func tokenizeBoolean(ctx *context) (jsonToken, error) {
    const (
        Initial dfsState = iota
        T
        Tr
        Tru
        F
        Fa
        Fal
        Fals
        Acc
    )

    var start = ctx.cursor
    var col = ctx.colNum
    var stage = Initial
    var goPanic = false
    var r rune
    var err error

loop:
    for r, err = getNextRune(ctx); err == nil; r, err = getNextRune(ctx) {
        switch stage {
        case Initial:
            if r == 't' { stage = T; continue }
            if r == 'f' { stage = F; continue }
            goPanic = true
            break loop
        case T:     if r == 'r' { stage = Tr } else { goPanic = true; break loop }
        case Tr:	if r == 'u' { stage = Tru } else { goPanic = true; break loop }
        case Tru:	if r == 'e' { stage = Acc } else { goPanic = true; break loop }
        case F:		if r == 'a' { stage = Fa } else { goPanic = true; break loop }
        case Fa:	if r == 'l' { stage = Fal } else { goPanic = true; break loop }
        case Fal:	if r == 's' { stage = Fals } else { goPanic = true; break loop }
        case Fals:	if r == 'e' { stage = Acc } else { goPanic = true; break loop }
        case Acc:   if !delimiters[r] { goPanic = true }; break loop
        }
    }

    if stage == Acc && !goPanic {
        if err == nil { back(ctx) } // back to the delimiter
        var token = jsonToken {
            Typ: Boolean,
            Val: string(ctx.rs[start:ctx.cursor]),
            Loc: location{ctx.lineNum, col},
        }
        return token, nil
    }

    if goPanic {
        back(ctx) // back to the rune which caused the panic
        for r, err = getNextRune(ctx); err == nil && !delimiters[r]; r, err = getNextRune(ctx) { }
        if err == nil { back(ctx) } // back to the delimiter
    }
    var jerr = jsonError{ InvalidToken, location{ ctx.lineNum, col } }
    return jsonToken{}, jerr
}

func tokenizeNull(ctx *context) (jsonToken, error) {
    const (
        Initial dfsState = iota
        N
        Nu
        Nul
        Acc
    )

    var stage = Initial
    var start = ctx.cursor
    var col = ctx.colNum
    var goPanic = false
    var r rune
    var err error

loop:
    for r, err = getNextRune(ctx); err == nil; r, err = getNextRune(ctx) {
        switch stage {
        case Initial:	if r == 'n' { stage = N } else { goPanic = true; break loop }
        case N: 		if r == 'u' { stage = Nu } else { goPanic = true; break loop }
        case Nu: 		if r == 'l' { stage = Nul } else { goPanic = true; break loop }
        case Nul:		if r == 'l' { stage = Acc } else { goPanic = true; break loop }
        case Acc:       if !delimiters[r] { goPanic = true }; break loop
        }
    }

    if stage == Acc && !goPanic {
        if err == nil { back(ctx) } // back to the delimiter
        var token = jsonToken {
            Typ: Null,
            Val: string(ctx.rs[start:ctx.cursor]),
            Loc: location{ctx.lineNum, col},
        }
        return token, nil
    }

    if goPanic {
        back(ctx) // back to the rune which caused the panic
        for r, err = getNextRune(ctx); err == nil && !delimiters[r]; r, err = getNextRune(ctx) { }
        if err == nil { back(ctx) } // back to the delimiter
    }
    var jerr = jsonError{ InvalidToken, location{ ctx.lineNum, col } }
    return jsonToken{}, jerr
}

func tokenizeNumber(ctx *context) (jsonToken, error) {
    const (
        Initial dfsState = iota
        Neg
        Zero
        Integer
        Dot
        Frac
        E
        ESign
        Exp
    )

    var stage = Initial
    var start = ctx.cursor
    var col = ctx.colNum
    var goPanic = false
    var r rune
    var err error

loop:
    for r, err = getNextRune(ctx); err == nil; r, err = getNextRune(ctx) {
        switch stage {
        case Initial:
            if r == '0' { stage = Zero; continue }
            if r == '-' { stage = Neg; continue }
            if unicode.IsDigit(r) { stage = Integer; continue }
            goPanic = true
            break loop
        case Zero:
            if r == '.' { stage = Dot; continue }
            if r == 'e' || r == 'E' { stage = E; continue }
            if delimiters[r] { break loop } else { goPanic = true; break loop }
        case Neg:
            if r == '0' { stage = Zero; continue }
            if unicode.IsDigit(r) { stage = Integer; continue }
            goPanic = true
            break loop
        case Integer:
            if unicode.IsDigit(r) { stage = Integer; continue }
            if r == '.' { stage = Dot; continue }
            if r == 'e' || r == 'E' { stage = E; continue }
            if delimiters[r] { break loop } else { goPanic = true; break loop }
        case Dot:
            if unicode.IsDigit(r) { stage = Frac; continue }
            goPanic = true
            break loop
        case Frac:
            if unicode.IsDigit(r) { stage = Frac; continue }
            if r == 'e' || r == 'E' { stage = E; continue }
            if delimiters[r] { break loop } else { goPanic = true; break loop }
        case E:
            if r == '-' || r == '+' { stage = ESign; continue }
            if unicode.IsDigit(r) { stage = Exp; continue }
            goPanic = true
            break loop
        case ESign:
            if unicode.IsDigit(r) { stage = Exp; continue }
            goPanic = true
            break loop
        case Exp:
            if unicode.IsDigit(r) { stage = Exp; continue }
            if delimiters[r] { break loop } else { goPanic = true; break loop }
        }
    }

    if (stage == Zero || stage == Integer || stage == Frac || stage == Exp) && !goPanic {
        if err == nil { back(ctx) } // back to the delimiter
        var token = jsonToken {
            Typ: Number,
            Val: string(ctx.rs[start:ctx.cursor]),
            Loc: location{ctx.lineNum, col},
        }
        return token, nil
    }

    var errType ErrorType
    if stage == Dot && (err != nil || delimiters[r]) {
        errType = MissFracPart
    } else if (stage == E || stage == ESign) && (err != nil || delimiters[r]) {
        errType = MissExponentPart
    } else {
        errType = InvalidToken
    }

    if goPanic {
        back(ctx) // back to the rune which caused the panic
        for r, err = getNextRune(ctx); err == nil && !delimiters[r]; r, err = getNextRune(ctx) { }
        if err == nil { back(ctx) } // back to the delimiter
    }
    var jerr error = jsonError{ errType, location{ ctx.lineNum,col } }
    return jsonToken{}, jerr
}

func tokenizeString(ctx *context) (jsonToken, error) {
    const (
        Initial dfsState = iota
        Open
        Escape
        Unicode
        Hex
        HexHex
        HexHexHex
        Acc
    )

    var stage = Initial
    var start = ctx.cursor
    var col = ctx.colNum
    var isClose = false
    var goPanic = false
    var r rune
    var err error

loop:
    for r, err = getNextRune(ctx); err == nil; r, err = getNextRune(ctx) {
        switch stage {
        case Initial: 	if r == '"' { stage = Open } else { goPanic = true; break loop }
        case Unicode: 	if isHex(r) { stage = Hex } else { goPanic = true; break loop }
        case Hex: 		if isHex(r) { stage = HexHex } else { goPanic = true; break loop }
        case HexHex: 	if isHex(r) { stage = HexHexHex } else { goPanic = true; break loop }
        case HexHexHex: if isHex(r) { stage = Open } else { goPanic = true; break loop }
        case Open:
            if r == '"' { isClose = true; stage = Acc; break loop }
            if r == '\\' { stage = Escape; continue }
            if r >= 0x0020 && r <= 0x10FFF && !unicode.IsControl(r) { stage = Open; continue }
            goPanic = true
            break loop
        case Escape:
            if r == 'u' { stage = Unicode; continue }
            if isEscapable(r) { stage = Open; continue }
            goPanic = true
            break loop
        }
    }

    if stage == Acc {
        var token = jsonToken {
            Typ: String,
            Val: string(ctx.rs[start:ctx.cursor]),
            Loc: location{ctx.lineNum, col},
        }
        return token, nil
    }

    if goPanic {
        back(ctx) // back to the rune which caused the panic
        var pre rune = -1
        for r, err = getNextRune(ctx); err == nil; r, err = getNextRune(ctx) {
            if r == '\n' { break }
            if r == '"' && pre != '\\' {
                isClose = true
                break
            }
            pre = r
        }
        if r == '\n' { back(ctx) } // back to the newline
    }

    var errTyp ErrorType
    if !isClose {
        errTyp = MissCloseQuote
    } else if stage == Unicode || stage == Hex || stage == HexHex || stage == HexHexHex {
        errTyp = InvalidUnicode
    } else if stage == Escape {
        errTyp = InvalidEscape
    } else if stage == Open {
        errTyp = InvalidChar
    }
    var jerr = jsonError{ errTyp, location{ ctx.lineNum, col } }
    return jsonToken{}, jerr
}

func isHex(r rune) bool {
    return unicode.IsDigit(r) || r >= 'A' && r <= 'F' || r >= 'a' && r <= 'f'
}

func isEscapable(r rune) bool {
    return r == '\\' || r == 'b' || r == 'f' || r == 'n' || r == 'r' || r == 't' || r == '"' || r == 'u'
}