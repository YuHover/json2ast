package json2ast

import (
    "testing"
)

func TestTokenizeBoolean(t *testing.T) {
	var jsonStrs = [...]string{
		"true", 	"false",
		"true ",	"false ",
		"true\t", 	"false\t",
		"true\r", 	"false\r",
		"true\n", 	"false\n",
		"true{", 	"false{",
		"true}}", 	"false}",
		"true[}", 	"false[",
		"true]}", 	"false]",
		"true\"}", 	"false\"",
		"true,}",	"false,",
		"true:}",	"false:",

		"tru", 		"fals}",
		"trua", 	"falsa]",
		"trve", 	"falve]",
		"truea", 	"falsea",
		"tr,ue,",	"fa,lse,",
	}

	for _, js := range jsonStrs {
        var ctx = context {
            rs:      []rune(js),
            cursor:  0,
            lineNum: 1,
            colNum:  1,
        }
		token, err := tokenizeBoolean(&ctx)
		t.Log(token, err, ctx)
	}
}

func TestTokenizeNull(t *testing.T) {
	var jsonStrs = [...]string{
		"null",
		"null ",
		"null\t",
		"null\r",
		"null\n",
		"null{",
		"null}",
		"null[",
		"null]",
		"null\"",
		"null,",
		"null:",

		"nul",
		"nula[",
		"nuLl",
		"nulll",
		"nu,ll,",
	}

	for _, js := range jsonStrs {
        var ctx = context {
            rs:      []rune(js),
            cursor:  0,
            lineNum: 1,
            colNum:  1,
        }
        token, err := tokenizeNull(&ctx)
        t.Log(token, err, ctx)
	}
}

func TestTokenizeNumber(t *testing.T) {
	validTests := []string{
		"0",
		"-0",
		"1",
		"-1",
		"0.1",
		"-0.1",
		"1234",
		"-1234",
		"12.34",
		"-12.34",
		"12E0",
		"12E1",
		"12e34",
		"12E-0",
		"12e+1",
		"12e-34",
		"-12E0",
		"-12E1",
		"-12e34",
		"-12E-0",
		"-12e+1",
		"-12e-34",
		"1.2E0",
		"1.2E1",
		"1.2e34",
		"1.2E-0",
		"1.2e+1",
		"1.2e-34",
		"-1.2E0",
		"-1.2E1",
		"-1.2e34",
		"-1.2E-0",
		"-1.2e+1",
		"-1.2e-34",
		"0E0",
		"0E1",
		"0e34",
		"0E-0",
		"0e+1",
		"0e-34",
		"-0E0",
		"-0E1",
		"-0e34",
		"-0E-0",
		"-0e+1",
		"-0e-34",
	}

	invalidTests := []string{
		"1.0.1",
		"1..1",
		"-1-2",
		"012a42",
		"01.2",
		"012",
		"12E12.12",
		"1e2e3",
		"1e+-2",
		"1e--23",
		"1e",
		"1e+",
		"1ea",
		"1a",
		"1.a",
		"1.",
		"01",
		"1.e1",
		"-",
		"+",
		"-,",
		"+123",
		"-1234.",
		"1.2e-",
		"1.33e+",
		".3",
		".34e-2",
	}

	for _, js := range validTests {
        var ctx = context {
            rs:      []rune(js),
            cursor:  0,
            lineNum: 1,
            colNum:  1,
        }
        token, err := tokenizeNumber(&ctx)
        t.Log(token, err, ctx)
	}

	for _, js := range invalidTests {
        var ctx = context {
            rs:      []rune(js),
            cursor:  0,
            lineNum: 1,
            colNum:  1,
        }
        token, err := tokenizeBoolean(&ctx)
        t.Log(token, err, ctx)
	}
}

func TestTokenizeString(t *testing.T) {
    validTests := []string {
        `""`,
        `"abc"`,
        `"\\\b\f\r\n\t\"\u0030\"\t\n\r\f\b\\"`,
        `"\\\b\f\r\n\t\"\u0030\"\t\n\r\f\b\\abc"`,
        `"abc\\\b\f\r\n\t\"\u0030\"\t\n\r\f\b\\"`,
        `"abc\\\b\f\r\n\t\"\u0048\"\t\n\r\f\b\\abc"`,
        `"\\\\\\\"\"\"\"\\\""`,
        `"""`,
    }

    invalidTests := []string {
        `"`,
        `"\"`,
        `"\\\b\z\\xyz"others`,
        `"xyz\u888xyz"others`,
        `"xyz\u"others`,
        `"xyz\uf"others`,
        `"xyz\uff"others`,
        `"xyz\ufff"others`,
        `"xyz\ufffothers`,
    }

    for _, js := range validTests {
        var ctx = context {
            rs:      []rune(js),
            cursor:  0,
            lineNum: 1,
            colNum:  1,
        }
        token, err := tokenizeString(&ctx)
        t.Log(token, err, ctx)
    }

    for _, js := range invalidTests {
        var ctx = context {
            rs:      []rune(js),
            cursor:  0,
            lineNum: 1,
            colNum:  1,
        }
        token, err := tokenizeString(&ctx)
        t.Log(token, err, ctx)
    }
}

func TestTokenize(t *testing.T) {
    validTests := `
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

    jtks, jerrs := tokenize(validTests)
    for _, v := range jtks {
        t.Logf("%+v\n", v)
    }
    t.Log(len(jerrs) == 0)

    invalidTest := `
    {
        "bool1": truefalse,
        "bool2": Fallse,
        "Null": nUll,
        "Number0": +0,
        "Number1": 12.,
        "Number2": 0e1.2,
        "String1": "abc",
        "String2": "\\\b\z\\xyz",
        "String3": "xyz\u",
        "String4": "xyz\uf",
        "String5": "xyz\uf",
        "String6": "xyz\uff",
        "String7": "xyz\ufff",
        "String8": "xyz\u0r00",
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
    jtks, jerrs = tokenize(invalidTest)
    for _, v := range jerrs {
        t.Logf("%+v\n", v)
    }
    for _, v := range jtks {
        t.Logf("%+v\n", v)
    }
}