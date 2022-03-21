package json2ast

import "errors"

type AstType uint8

const (
    Object AstType = iota
    Array
    Literal
)

type jsonLiteralAst jsonToken

type JsonAst struct {
    ObjectAst   map[string]JsonAst
    ArrayAst    []JsonAst
    LiteralAst  jsonLiteralAst
    Typ         AstType
}

var cursor = 0
var jts []jsonToken
var jerrs []jsonError

func Parser(source string) JsonAst {
    jts, jerrs = tokenize(source)

    if len(jerrs) != 0 || len(jts) == 0 {
        return JsonAst{} // fuck
    }
    cursor = 0
    return parseElement()
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

func parseElement() JsonAst {
    token, err := getToken()
    if err != nil {
        // ValueExpected
        // return JsonAst{}, errors.New("ValueExpected")
        return JsonAst{}
    }

    var ast JsonAst
    switch token.Typ {
    case LeftBrace:
        ast.ObjectAst = parseObject()
        ast.Typ = Object
    case LeftBracket:
        ast.ArrayAst = parseArray()
        ast.Typ = Array
    case String: fallthrough
    case Number: fallthrough
    case Boolean: fallthrough
    case Null:
        ast.LiteralAst = jsonLiteralAst(token)
        ast.Typ = Literal
    default: // TODO ERROR
    }

    return ast
}

func parseObject() map[string]JsonAst {
    token, err := getToken()
    if err != nil {
        return nil
        // return nil, errors.New("CommaOrClosingBraceExpected")
    }

    var objectAst = make(map[string]JsonAst, 0)

    if token.Typ == RightBrace {
        return objectAst
    }

    if token.Typ == String {
        var key = token.Val

        token, err = getToken()
        if err != nil {
            return nil
            // return nil, errors.New("ColonExpected")
        }

        if token.Typ == Colon {
            objectAst[key] = parseElement()

            others := parseObjMembers()
            for k, v := range others { // TODO duplicated key
                objectAst[k] = v
            }

            token, err = getToken()
            if err != nil {
                return nil
                // return nil, errors.New("CommaOrClosingBraceExpected")
            }

            if token.Typ == RightBrace {
                return objectAst
            }
        }
    }

    // TODO ERROR
    return nil
}

func parseObjMembers() map[string]JsonAst {
    token, err := getToken()
    if err != nil {
        return nil
        // return nil, errors.New("CommaOrClosingBraceExpected")
    }

    if token.Typ == RightBrace {
        cursor-- // ɛ
        return nil
    }

    if token.Typ == Comma {
        token, err = getToken()
        if err != nil {
            return nil
            // return nil, errors.New("PropertyExpected")
        }

        var objectAst = make(map[string]JsonAst, 0)
        if token.Typ == String {
            var key = token.Val

            token, err = getToken()
            if err != nil {
                return nil
                // return nil, errors.New("ColonExpected")
            }

            if token.Typ == Colon {
                objectAst[key] = parseElement()

                others := parseObjMembers()
                for k, v := range others { // TODO duplicated key
                    objectAst[k] = v
                }
                return objectAst
            }
        }
    }

    // TODO ERROR
    return nil
}

func parseArray() []JsonAst {
    token, err := getToken()
    if err != nil {
        return nil
        // return nil, errors.New("CommaOrClosingBracketExpected")
    }

    var arrayAst = make([]JsonAst, 0)

    if token.Typ == RightBracket {
        return arrayAst
    }

    if  token.Typ == LeftBrace || token.Typ == LeftBracket || token.Typ == String ||
        token.Typ == Number || token.Typ == Boolean || token.Typ == Null {
        cursor--
        arrayAst = append(arrayAst, parseElement())
        arrayAst = append(arrayAst, parseAryElements()...)

        token, err = getToken()
        if err != nil {
            return nil
            // return nil, errors.New("CommaOrClosingBracketExpected")
        }

        if token.Typ == RightBracket {
            return arrayAst
        }
    }

    // TODO ERROR
    return nil
}

func parseAryElements() []JsonAst {
    token, err := getToken()
    if err != nil {
        return nil
        // return nil, errors.New("CommaOrClosingBracketExpected")
    }

    if token.Typ == RightBracket {
        cursor-- // ɛ
        return nil
    }

    if token.Typ == Comma {
        var arrayAst = make([]JsonAst, 0)
        arrayAst = append(arrayAst, parseElement())
        arrayAst = append(arrayAst, parseAryElements()...)
        return arrayAst
    }

    // TODO ERROR
    return nil
}
