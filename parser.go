package gtfs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"log"
)

type Parser string

type ParseError struct {
	Message    string
	LineNumber int
	FileName   string
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("ParseError in file %v at line %d: %v", pe.FileName, pe.LineNumber, pe.Message)
}

type settableThroughField interface {
	setField(fieldName string, value string)
}

func fieldsSetter(model settableThroughField, fieldKeys, fieldValues []string) {
	for i, key := range fieldKeys {
		model.setField(key, fieldValues[i])
	}
}

func (p *Parser) parse(r io.Reader, recordHandler func(k, v []string)) error {

	lineNumber := 1

	reader := bufio.NewReader(r)

	firstline, isPrefix, err := reader.ReadLine()
	if err != nil {
		perr := &ParseError{Message: err.Error()}
		perr.FileName = string(*p)
		perr.LineNumber = lineNumber
		return perr
	} else if isPrefix {
		return errors.New(fmt.Sprintf("First line too long (not handled yet, oups): \"%v\"", p))
	}

	fieldKeys, perr := p.parseLine(firstline)
	if perr != nil {
		perr.FileName = string(*p)
		perr.LineNumber = lineNumber
		return perr
	}

	line, isPrefix, err := reader.ReadLine()
	for err == nil {
		if err != nil {
			panic(err)
		} else if isPrefix {
			return errors.New(fmt.Sprintf("First line too long (not handled yet, oups): \"%v\"", p))
		}

		lineNumber = lineNumber + 1

		fieldValues, perr := p.parseLine(line)
		if perr != nil {
			perr.FileName = string(*p)
			perr.LineNumber = lineNumber
			log.Println(perr)
			// return perr
		} else {
			lengthdiff := len(fieldKeys) - len(fieldValues)
			if lengthdiff != 0 && lengthdiff > 0 {
				for lengthdiff > 0 {
					fieldValues = append(fieldValues, "")
					lengthdiff = lengthdiff - 1
				}
			}
			recordHandler(fieldKeys, fieldValues)
		}

		line, isPrefix, err = reader.ReadLine()
	}

	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (p *Parser) parseLine(line []byte) (tokens []string, err *ParseError) {
	reader := bufio.NewReader(strings.NewReader(string(line)))
	tokens = make([]string, 0, 10)
	var previousRune rune
	field := ""
	startedWithQuote := false
	charIndex := 0
	quoteCount := 0
	rune, size, error := reader.ReadRune()
	for {
		if error != nil || size == 0 {
			if error == io.EOF || size == 0 { // EOF is the gracious end of Read. Same for ReadRune? Seems like size==0 is replacing that
				return append(tokens, field), nil
			}
			return nil, &ParseError{Message: error.Error()}
		}

		switch rune {
		case '\t', '\n', '\r':
			var char string
			if rune == '\t' {
				char = "\\t"
			} else if rune == '\n' {
				char = "\\n"
			} else if rune == '\r' {
				char = "\\r"
			}
			return nil, &ParseError{Message: fmt.Sprintf("Found illegal character '%v' at char %d", char, charIndex)}
		case '"':
			if field == "" && !startedWithQuote {
				startedWithQuote = true
				break
			}

			if field != "" && !startedWithQuote {
				log.Println(fmt.Sprintf("Unexpected quote (\") found at char %d", charIndex))
				field = field + string(rune)
				// return nil, &ParseError{Message: fmt.Sprintf("Unexpected quote (\") found at char %d", charIndex)}
			}

			if quoteCount == 1 {
				quoteCount = 0
			} else {
				quoteCount = 1
			}

			if startedWithQuote && previousRune == '"' {
				field = field + string(rune)
			}
			break
		case ',':
			if (startedWithQuote && quoteCount == 1) || !startedWithQuote {
				tokens = append(tokens, field)
				field = ""
				startedWithQuote = false
				quoteCount = 0
			} else {
				field = field + string(rune)
			}
			break
		case ' ':
			if field != "" {
				field = field + string(rune)
			}
			break
		default:
			field = field + string(rune)
			break
		}
		previousRune = rune
		charIndex += 1
		rune, size, error = reader.ReadRune()
	}

	return tokens, nil
}
