package jsonparser

import "fmt"

type errorType uint8

const (
    InvalidEscape   errorType = iota
    InvalidChar
    InvalidUnicode
    InvalidToken
    MissCloseQuote
)

type location struct {
    lineNum  int
    position    int
}

type jsonError struct {
    typ errorType
    loc location
}

func (jerr jsonError) Error() string {
    return fmt.Sprintf("[%d, %d], type: %v", jerr.loc.lineNum, jerr.loc.position, jerr.typ)
}