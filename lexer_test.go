package jsonparser

import "testing"

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
		token, cursor, err := tokenizeBoolean([]rune(js), 0)
		t.Log(token, err, cursor)
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
		token, cursor, err := tokenizeNull([]rune(js), 0)
		t.Log(token, err, cursor)
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
		token, cursor, err := tokenizeNumber([]rune(js), 0)
		t.Log(js, token, err, cursor)
	}

	for _, js := range invalidTests {
		token, cursor, err := tokenizeNumber([]rune(js), 0)
		t.Log(js, token, err, cursor)
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
        token, cursor, err := tokenizeString([]rune(js), 0)
        t.Log(js, token, err, cursor)
        // t.Logf("%#v, %#v, %#v, %#v,", js, token, err, cursor)
    }

    for _, js := range invalidTests {
        token, cursor, err := tokenizeString([]rune(js), 0)
        t.Log(js, token, err, cursor)
    }
}