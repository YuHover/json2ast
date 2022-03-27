package json2ast

import "fmt"

type ErrorType uint8

const (
    InvalidEscape ErrorType = iota
    InvalidChar
    InvalidUnicode
    InvalidToken
    MissCloseQuote
    MissFracPart
    MissExponentPart

    ValueExpected
    TrailingComma
    CommaExpected
    PropertyOrClosingBraceExpected
    PropertyExpected
    ColonExpected
    CommaOrClosingBraceExpected
    CommaOrClosingBracketExpected
    EndOfJsonExpected
)

var descriptions = map[ErrorType]string {
    InvalidEscape: "InvalidEscape",
    InvalidChar: "InvalidChar",
    InvalidUnicode: "InvalidUnicode",
    InvalidToken: "InvalidToken",
    MissCloseQuote: "MissCloseQuote",
    MissFracPart: "MissFracPart",
    MissExponentPart: "MissExponentPart",

    ValueExpected: "ValueExpected",
    TrailingComma: "TrailingComma",
    CommaExpected: "CommaExpected",
    PropertyOrClosingBraceExpected: "PropertyOrClosingBraceExpected",
    PropertyExpected: "PropertyExpected",
    ColonExpected: "ColonExpected",
    CommaOrClosingBraceExpected: "CommaOrClosingBraceExpected",
    CommaOrClosingBracketExpected: "CommaOrClosingBracketExpected",
    EndOfJsonExpected: "EndOfJsonExpected",
}

type location struct {
    lineNum  int
    position    int
}

type jsonError struct {
    typ ErrorType
    loc location
}

func (jerr jsonError) Error() string {
    return fmt.Sprintf("[%d, %d], type: %s", jerr.loc.lineNum, jerr.loc.position, descriptions[jerr.typ])
}