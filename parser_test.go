package json2ast

import (
    "encoding/json"
    "reflect"
    "strings"
    "testing"
)

func TestParser(t *testing.T) {
    validTests := []string {
    `
    {
        "bool1": true,
        "bool2": false,
        "Null": null,
        "Number0": 0,
        "Number1": -0,
        "Number2": 1,
        "Number3": -1,
        "Number4": 0.1,
        "Number5": -0.1,
        "Number6": 1234,
        "Number7": -1234,
        "Number8": 12.34,
        "Number9": -12.34,
        "Number10": 12E0,
        "Number11": 12E1,
        "Number12": 12e34,
        "Number13": 12E-0,
        "Number14": 12e+1,
        "Number15": 12e-34,
        "Number16": -12E0,
        "Number17": -12E1,
        "Number18": -12e34,
        "Number19": -12E-0,
        "Number20": -12e+1,
        "Number21": -12e-34,
        "Number22": 1.2E0,
        "Number23": 1.2E1,
        "Number24": 1.2e34,
        "Number25": 1.2E-0,
        "Number26": 1.2e+1,
        "Number27": 1.2e-34,
        "Number28": -1.2E0,
        "Number29": -1.2E1,
        "Number30": -1.2e34,
        "Number31": -1.2E-0,
        "Number32": -1.2e+1,
        "Number33": -1.2e-34,
        "Number34": 0E0,
        "Number35": 0E1,
        "Number36": 0e34,
        "Number37": 0E-0,
        "Number38": 0e+1,
        "Number39": 0e-34,
        "Number40": -0E0,
        "Number41": -0E1,
        "Number42": -0e34,
        "Number43": -0E-0,
        "Number44": -0e+1,
        "Number45": -0e-34,
        "String0": "",
        "String1": "abc",
        "String2": "\\\b\f\r\n\t\"\u0030\"\t\n\r\f\b\\",
        "String3": "\\\b\f\r\n\t\"\u0030\"\t\n\r\f\b\\abc",
        "String4": "abc\\\b\f\r\n\t\"\u0030\"\t\n\r\f\b\\",
        "String5": "abc\\\b\f\r\n\t\"\u0048\"\t\n\r\f\b\\abc",
        "String6": "\\\\\\\"\"\"\"\\\"",
        "Array0": [
            {"type": "home", "number": "212 555-1234"},
            {"type": "fax", "number": "646 555-4567"}
        ],
        "object0": {
            "streetAddress": "21 2nd Street",
            "city": "New York",
            "state": "NY",
            "postalCode": "10021"
        }
    }`,
    `123`,
    `null`,
    `true`,
    `false`,
    `"string"`,
    `[123, null, true, false, "string", {"type": "home", "number": "212 555-1234"}]`,
    }
    for _, vt := range validTests {
        if !json.Valid([]byte(vt)) {
            t.Fatalf("valid test `%s` is invalid", vt)
        }

        ast, jerrs := Parser(vt)
        if len(jerrs) != 0 {
            t.Fatalf("build AST for `%s` failed", vt)
        }

        jsonText := ast2JsonText(ast)
        if !json.Valid([]byte(jsonText)) {
            t.Fatalf("json text `%s` is invalid, corresponding test is `%s`", jsonText, vt)
        }

        if ast.Typ == Object {
            vtObj := make(map[string]interface{})
            astObj := make(map[string]interface{})
            doTestParser(t, vt, vtObj, jsonText, astObj)
            continue
        }

        if ast.Typ == Array {
            vtObj := make([]interface{}, 0)
            astObj := make([]interface{}, 0)
            doTestParser(t, vt, vtObj, jsonText, astObj)
            continue
        }

        if ast.Typ == Literal {
            if jsonText != vt {
                t.Fatalf("json object of `%s` is not equal to `%s`", jsonText, vt)
            }
            continue
        }

        t.Fatalf("unsupported AST type: %d", ast.Typ)
    }

    invalidTests := []string {
        `{{[], "k1":{[]{:123}, "k2":{][{, "k3":123}, "k4":{1 true null false: 123}, "k5":{, "k6":123}, "k7":{:123}}`,
        `{"k1":123 "k2":123, "k3":123, "k4":{"k5":123 {][ 123 true false null:, "k6": 123 }}`,
        `{"k1":[}:}:, 123], "k2":[:}:}, 123], "k3":[,123,123], "k4":[}:"v1"], "k5":[:}123], "k6":[:}true], "k7":[:}[]]}`,
        `{"k1":[}}{"k2":123}], "k3":[}}{"k4":123}], "k5":[}}null, 123]}`,
        `{"k1":[123}:}:], "k2":[123}:}:, 123], "k3":[123:}:}{"k4":123}, 123], "k5":[123}:}:[123,123]], "k6":[123 123]}`,
        `{"k1":[123 "v1"], "k2":[123 true], "k3":[123 false], "k4":[123 null]}`,
        `{"k1":[123,], "k2":[123,}:}::,123], "k3":[123,:}:{}], "k4":[123,:}:[123]], "k5":[123,:}:"v1"], "k6":[123,:}:123]}`,
        `{"k1":[123,}:false], "k2":[123,}:true], "k3":[123,}:null]}, "k4":[123,,456]}`,
        `{"k1":{"k2"::]:]}, "k3":}}`,
        `{"k1":]:], "k2":123, "k3":{"k4":,}}`,
        `,:}]{"k1":123}{},:[] "k1" 123 null false true`,
    }

    for _, ivt := range invalidTests {
        if json.Valid([]byte(ivt)) {
            t.Fatalf("invalid test `%s` is valid", ivt)
        }

        _, jerrs := Parser(ivt)
        t.Logf("errors of `%s` :", ivt)
        for _, jerr := range jerrs {
            t.Log(jerr)
        }
        t.Log()
    }
}

func doTestParser(t *testing.T, vt string, vtObj interface{}, jsonText string, astObj interface{}) {
    err := json.Unmarshal([]byte(vt), &vtObj)
    if err != nil {
        t.Fatalf("unmarshal valid test `%s` failed", vt)
    }

    err = json.Unmarshal([]byte(jsonText), &astObj)
    if err != nil {
        t.Fatalf("unmarshal json text `%s` failed, corresponding test is `%s`", jsonText, vt)
    }

    if !reflect.DeepEqual(astObj, vtObj) {
        t.Fatalf("json object of `%s` is not equal to `%s`", jsonText, vt)
    }
}

func ast2JsonText(ast JsonAst) string {
    var sb strings.Builder
    switch ast.Typ {
    case Object:
        var i = 0
        sb.WriteRune('{')
        for k, v := range ast.ObjectAst {
            sb.WriteString(k)
            sb.WriteString(": ")
            sb.WriteString(ast2JsonText(v))
            if i != len(ast.ObjectAst)-1 { sb.WriteString(", ") }
            i++
        }
        sb.WriteRune('}')
    case Array:
        sb.WriteRune('[')
        for i, v := range ast.ArrayAst {
            sb.WriteString(ast2JsonText(v))
            if i != len(ast.ArrayAst)-1 { sb.WriteString(", ") }
        }
        sb.WriteRune(']')
    case Literal:
        sb.WriteString(ast.LiteralAst.Val)
    }

    return sb.String()
}
