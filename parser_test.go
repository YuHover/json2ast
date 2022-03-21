package json2ast

import (
    "fmt"
    "testing"
)

func TestParser(t *testing.T) {
    validTests := `
    {
        "bool1": true,
        "bool2": false,
        "Null": null,
        "Number0": 0,
        "String0": "",
        "Array0": [
            {"type": "home", "number": "212 555-1234"},
            {
                "type": "fax",
                "number": "646 555-4567"
            }
        ],
        "object0":
        {
            "streetAddress": "21 2nd Street",
            "city": "New York",
            "state": "NY",
            "postalCode": "10021"
        }
    }
    `
    ast := Parser(validTests)
    travel(ast)
}

func travel(ast JsonAst) {
    switch ast.Typ {
    case Object:
        fmt.Printf("{")
        for k, v := range ast.ObjectAst {
            fmt.Printf("%s: ", k)
            travel(v)
        }
        fmt.Printf("}, ")
    case Array:
        fmt.Printf("[")
        for _, v := range ast.ArrayAst {
            travel(v)
        }
        fmt.Printf("], ")
    case Literal:
        fmt.Printf("%s, ", ast.LiteralAst.Val)
    }
}
