package json2ast

import (
    "errors"
)

type AstType uint8

const (
    Object AstType = iota
    Array
    Literal
)

type nonTerminal uint8

const (
    parser nonTerminal = iota
    element
    array
    elements
    object
    members
)

var firstSet = map[nonTerminal]map[tokenType]bool {
    element: {LeftBrace: true, LeftBracket: true, String: true, Number: true, Boolean: true, Null: true},
    array: {LeftBrace: true, LeftBracket: true, RightBracket: true ,String: true, Number: true, Boolean: true, Null: true},
    elements: {RightBracket: true, Comma: true},
    object: {RightBrace: true, String: true},
    members: {RightBrace: true, Comma: true},
}

type literalAst jsonToken

type JsonAst struct {
    ObjectAst   map[string]JsonAst
    ArrayAst    []JsonAst
    LiteralAst  literalAst
    Typ         AstType
}

var cursor int
var jts []jsonToken
var jerrs []jsonError

func Parser(source string) (JsonAst, []jsonError) {
    jts, jerrs = tokenize(source)

    if len(jerrs) != 0 || len(jts) == 0 {
        return JsonAst{}, jerrs
    }

    cursor = 0
    ast := parseElement(parser)
    // expected end of json
    if cursor < len(jts) {
        jerrs = append(jerrs, jsonError{EndOfJsonExpected, jts[cursor].Loc})
    }

    if len(jerrs) != 0 {
        ast = JsonAst{}
    }

    return ast, jerrs
}

func getToken() (jsonToken, error) {
    var token jsonToken
    if cursor < len(jts) {
        token = jts[cursor]
        cursor++
        return token, nil
    }

    return token, errors.New("overstep")
}

func goPanic(nT nonTerminal, syncTokenTypes ...tokenType) (bool, bool, jsonToken) {
    var sync = false
    var first = false
    var token jsonToken
    var err error
    for token, err = getToken(); err == nil; token, err = getToken() {
        if firstSet[nT][token.Typ] {
            first = true
            break
        }
        for _, stt := range syncTokenTypes {
            if token.Typ == stt {
                sync = true
                break // break the closest for
            }
        }
        if sync { break }
    }

    return sync, first, token
}

func parseElement(caller nonTerminal) JsonAst {
    token, err := getToken()
    if err != nil {
        jerrs = append(jerrs, jsonError{ValueExpected, location{-1, -1}})
        return JsonAst{}
    }

    if firstSet[element][token.Typ] {
        return doParseElement(token)
    }

    if caller == parser {
        jerrs = append(jerrs, jsonError{ValueExpected, token.Loc})
        _, first, token := goPanic(element)
        if first { _ = doParseElement(token) }
        return JsonAst{}
    }

    if caller == elements {
        if token.Typ == RightBrace || token.Typ == Colon {
            jerrs = append(jerrs, jsonError{ValueExpected, token.Loc})
            sync, first, token := goPanic(element, RightBracket, Comma)
            if sync { cursor-- } // cursor go back to the Comma or RightBracket
            if first { _ = doParseElement(token) }
            return JsonAst{}
        }

        cursor-- // cursor go back to the Comma or RightBracket
        if token.Typ == Comma { // Comma
            jerrs = append(jerrs, jsonError{ValueExpected, token.Loc})
        } else { // RightBracket
            jerrs = append(jerrs, jsonError{TrailingComma, jts[cursor-1].Loc}) // Trailing comma
        }
        return JsonAst{}
    }

    if caller == object || caller == members {
        jerrs = append(jerrs, jsonError{ValueExpected, token.Loc})
        if token.Typ == RightBracket || token.Typ == Colon {
            sync, first, token := goPanic(element, RightBrace, Comma)
            if sync { cursor-- } // cursor go back to the Comma or RightBrace
            if first { _ = doParseElement(token) }
            return JsonAst{}
        }

        cursor-- // cursor go back to the Comma or RightBrace
        return JsonAst{}
    }

    return JsonAst{}
}

func doParseElement(token jsonToken) JsonAst {
    var ast JsonAst

    if token.Typ == LeftBrace {
        ast.ObjectAst = parseObject()
        ast.Typ = Object
    } else if token.Typ == LeftBracket {
        ast.ArrayAst = parseArray()
        ast.Typ = Array
    } else {
        ast.LiteralAst = literalAst(token)
        ast.Typ = Literal
    }

    return ast
}

func parseObject() map[string]JsonAst {
    token, err := getToken()
    if err != nil {
        jerrs = append(jerrs, jsonError{PropertyOrClosingBraceExpected, location{-1, -1}})
        return nil
    }

    if token.Typ == RightBrace {
        return map[string]JsonAst{}
    }
    if token.Typ == String {
        var objAst = doParseMember(token)
        if isNextRightBrace() { return objAst }
        return nil
    }

    jerrs = append(jerrs, jsonError{PropertyOrClosingBraceExpected, token.Loc})
    cursor-- // for this token may also be sync tokens Comma or Colon
    sync, first, token := goPanic(object, Comma, Colon)
    if first {
        if token.Typ == String {
            _ = doParseMember(token)
            _ = isNextRightBrace()
        } // else RightBrace
        return nil
    }
    if sync {
        if token.Typ == Comma {
            cursor-- // cursor go back to the Comma
        } else { // Colon
            parseElement(object)
        }
        _ = parseObjMembers()
        _ = isNextRightBrace()
        return nil
    }

    return nil
}

func parseObjMembers() map[string]JsonAst {
    token, err := getToken()
    if err != nil {
        jerrs = append(jerrs, jsonError{CommaOrClosingBraceExpected, location{-1, -1}})
        return nil
    }

    if firstSet[members][token.Typ] {
        return doParseObjMembers(token)
    }

    if token.Typ == String {
        jerrs = append(jerrs, jsonError{CommaExpected, token.Loc})
        _ = doParseMember(token)
        return nil
    }

    jerrs = append(jerrs, jsonError{CommaOrClosingBraceExpected, token.Loc})
    sync, first, token := goPanic(members, String)
    if first {
        _ = doParseObjMembers(token)
        return nil
    }
    if sync {
        _ = doParseMember(token)
        return nil
    }

    return nil
}

func doParseObjMembers(token jsonToken) map[string]JsonAst {
    if token.Typ == RightBrace {
        cursor-- // ɛ
        return map[string]JsonAst{}
    }

    if token.Typ == Comma {
        token, err := getToken()
        if err != nil {
            jerrs = append(jerrs, jsonError{PropertyExpected, location{-1, -1}})
            return nil
        }

        if token.Typ == String {
            return doParseMember(token)
        }

        if token.Typ == Colon {
            cursor-- // pretend to insert a String
            jerrs = append(jerrs, jsonError{PropertyExpected, token.Loc})
            _ = doParseMember(jsonToken{Val: "dummy"})
            return nil
        }

        if token.Typ == Comma {
            jerrs = append(jerrs, jsonError{PropertyExpected, token.Loc})
            cursor-- // cursor go back to the Comma
            _ = parseObjMembers()
            return nil
        }

        if token.Typ == RightBrace {
            cursor-- // cursor go back to the RightBrace
            jerrs = append(jerrs, jsonError{TrailingComma, jts[cursor-1].Loc}) // Trailing comma
            return nil
        }

        // other cases
        jerrs = append(jerrs, jsonError{PropertyExpected, token.Loc})
        sync, first, token := goPanic(members, String, Colon)
        if first {
            cursor-- // cursor go back to the Comma or RightBrace
            _ = parseObjMembers()
            return nil
        }
        if sync {
            if token.Typ == String {
                _ = doParseMember(token)
                return nil
            } else { // Colon
                cursor-- // pretend to insert a string key
                _ = doParseMember(jsonToken{Val: "dummy"})
                return nil
            }
        }
        return nil
    }

    return nil
}

func doParseMember(token jsonToken) map[string]JsonAst {
    var objAst = make(map[string]JsonAst)
    var key = token.Val

    token, err := getToken()
    if err != nil {
        jerrs = append(jerrs, jsonError{ColonExpected, location{-1, -1}})
        return nil
    }

    if token.Typ == Colon {
        objAst[key] = parseElement(object)
        others := parseObjMembers()
        for k, v := range others {
            objAst[k] = v
        }
        return objAst
    }

    jerrs = append(jerrs, jsonError{ColonExpected, token.Loc})

    if token.Typ == Comma {
        cursor-- // cursor go back to the Comma
        _ = parseObjMembers()
        return nil
    }

    if token.Typ != Colon {
        cursor-- // pretend to insert a Colon
        parseElement(object)
        _ = parseObjMembers()
        return nil
    }

    return nil
}

func isNextRightBrace() bool {
    token, err := getToken()
    if err != nil || token.Typ != RightBrace {
        if err != nil {
            jerrs = append(jerrs, jsonError{CommaOrClosingBraceExpected, location{-1, -1}})
        } else {
            jerrs = append(jerrs, jsonError{CommaOrClosingBraceExpected, token.Loc})
        }
        return false
    }
    return true
}

func parseArray() []JsonAst {
    token, err := getToken()
    if err != nil {
        jerrs = append(jerrs, jsonError{CommaOrClosingBracketExpected, location{-1, -1}})
        return nil
    }

    if firstSet[array][token.Typ] {
        return doParseArray(token)
    }

    jerrs = append(jerrs, jsonError{ValueExpected, token.Loc})
    cursor-- // for this token may also be sync token Comma
    sync, first, token := goPanic(array, Comma)
    if first {
        _ = doParseArray(token)
        return nil
    }
    if sync {
        cursor-- // cursor go back to the Comma
        _ = parseAryElements()
        _ = isNextRightBracket()
        return nil
    }

    return nil
}

func doParseArray(token jsonToken) []JsonAst {
    var arrayAst = make([]JsonAst, 0)

    if token.Typ == RightBracket {
        return arrayAst
    }

    cursor-- // cursor back to the "element"
    arrayAst = append(arrayAst, parseElement(array))
    arrayAst = append(arrayAst, parseAryElements()...)

    if isNextRightBracket() {
        return arrayAst
    }
    return nil
}

func isNextRightBracket() bool {
    token, err := getToken()
    if err != nil || token.Typ != RightBracket {
        if err != nil {
            jerrs = append(jerrs, jsonError{CommaOrClosingBracketExpected, location{-1, -1}})
        } else {
            jerrs = append(jerrs, jsonError{CommaOrClosingBracketExpected, token.Loc})
        }
        return false
    }
    return true
}

func parseAryElements() []JsonAst {
    token, err := getToken()
    if err != nil {
        jerrs = append(jerrs, jsonError{CommaOrClosingBracketExpected, location{-1, -1}})
        return nil
    }

    if token.Typ == RightBracket {
        cursor-- // ɛ
        return nil
    }

    if token.Typ == Comma {
        var arrayAst = make([]JsonAst, 0)
        arrayAst = append(arrayAst, parseElement(elements))
        arrayAst = append(arrayAst, parseAryElements()...)
        return arrayAst
    }

    jerrs = append(jerrs, jsonError{CommaOrClosingBracketExpected, token.Loc})

    if firstSet[element][token.Typ] {
        cursor-- // cursor back to the "element"
        _ = parseElement(elements)
        _ = parseAryElements()
        return nil
    }

    if token.Typ == RightBrace || token.Typ == Colon {
        _, first, token := goPanic(elements)
        if first {
            if token.Typ == RightBracket {
                cursor-- // cursor back to the RightBracket
                return nil
            } else { // Comma
                _ = parseElement(elements)
                _ = parseAryElements()
                return nil
            }
        }
    }

    return nil
}
