// Package parser provides low-level GEDCOM line parsing functionality.
//
// This package handles the tokenization and parsing of individual GEDCOM lines,
// converting them into Line structures with level, tag, value, and cross-reference
// information. It supports all standard GEDCOM formats and provides detailed error
// reporting with line numbers.
//
// Example usage:
//
//	p := parser.NewParser(reader)
//	for {
//	    line, err := p.ParseLine()
//	    if err == io.EOF {
//	        break
//	    }
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Level %d: %s = %s\n", line.Level, line.Tag, line.Value)
//	}
package parser

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// MaxNestingDepth is the maximum allowed nesting depth to prevent stack overflow.
const MaxNestingDepth = 100
const maxTagLength = 31

// Parser parses GEDCOM files into Line structures.
type Parser struct {
	lineNumber int
	lastLevel  int
}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{
		lineNumber: 0,
		lastLevel:  -1,
	}
}

// Reset resets the parser state for reuse.
func (p *Parser) Reset() {
	p.lineNumber = 0
	p.lastLevel = -1
}

// ParseLine parses a single GEDCOM line.
// GEDCOM line format: LEVEL [XREF] TAG [VALUE]
// Examples:
//
//	0 HEAD
//	0 @I1@ INDI
//	1 NAME John /Smith/
//	2 GIVN John
func (p *Parser) ParseLine(input string) (*Line, error) {
	p.lineNumber++

	// Trim line endings (CRLF, LF, CR)
	line := strings.TrimRight(input, "\r\n")

	// Empty or whitespace-only lines are invalid
	if strings.TrimSpace(line) == "" {
		return nil, newParseError(p.lineNumber, "empty line", input)
	}

	// Split into parts
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, newParseError(p.lineNumber, "line must have at least level and tag (expected a tag like HEAD, INDI, FAM, or SOUR)", line)
	}

	// Parse level (first part)
	level, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, wrapParseError(p.lineNumber, "invalid level number", line, &InvalidLevelError{
			Raw:    parts[0],
			Reason: "not a number",
		})
	}

	if level < 0 {
		return nil, wrapParseError(p.lineNumber, "level cannot be negative", line, &InvalidLevelError{
			Raw:    parts[0],
			Reason: "negative",
		})
	}

	// Check nesting depth
	if level > MaxNestingDepth {
		return nil, wrapParseError(p.lineNumber, "maximum nesting depth exceeded", line, &InvalidLevelError{
			Raw:    parts[0],
			Reason: "exceeds max depth",
		})
	}

	if p.lastLevel >= 0 && level > p.lastLevel+1 {
		return nil, wrapParseError(p.lineNumber, "level jump exceeds one", line, &LevelMismatchError{
			Previous: p.lastLevel,
			Current:  level,
		})
	}

	// Parse XRef and Tag
	var xref, tag string
	var valueStartIdx int

	// Check if second part is an XRef (starts with @ and ends with @)
	if strings.HasPrefix(parts[1], "@") && strings.HasSuffix(parts[1], "@") {
		xref = parts[1]
		if err := validateXRef(xref); err != nil {
			return nil, wrapParseError(p.lineNumber, err.Error(), line, err)
		}
		if len(parts) < 3 {
			return nil, newParseError(p.lineNumber, "line with xref must have a tag (expected a tag like INDI, FAM, or SOUR)", line)
		}
		tag = parts[2]
		valueStartIdx = 3
	} else {
		tag = parts[1]
		valueStartIdx = 2
	}

	if err := validateTag(tag); err != nil {
		message := err.Error()
		if _, ok := err.(*InvalidTagError); ok {
			message = message + " (expected A-Z, 0-9, underscore, max length 31)"
		}
		return nil, wrapParseError(p.lineNumber, message, line, err)
	}

	// Parse value (everything after the tag)
	var value string
	if valueStartIdx < len(parts) {
		// Find the position in the original line where the value starts
		// We need to preserve original spacing in the value
		tagPos := strings.Index(line, tag)
		if tagPos >= 0 {
			afterTag := tagPos + len(tag)
			if afterTag < len(line) {
				value = strings.TrimLeft(line[afterTag:], " ")
			}
		}
	}

	p.lastLevel = level

	return &Line{
		Level:      level,
		Tag:        tag,
		Value:      value,
		XRef:       xref,
		LineNumber: p.lineNumber,
	}, nil
}

// Parse reads a GEDCOM file from a reader and returns all parsed lines.
// Supports all line ending styles: LF (Unix), CRLF (Windows), CR (old Macintosh).
func (p *Parser) Parse(r io.Reader) ([]*Line, error) {
	p.Reset()

	scanner := bufio.NewScanner(r)
	// Use custom split function that handles CR, LF, and CRLF line endings
	scanner.Split(scanGEDCOMLines)
	var lines []*Line
	var prevLine string

	for scanner.Scan() {
		text := scanner.Text()
		line, err := p.ParseLine(text)
		if err != nil {
			return nil, enrichParseError(err, prevLine, text)
		}
		lines = append(lines, line)
		prevLine = text
	}

	if err := scanner.Err(); err != nil {
		return nil, wrapParseError(p.lineNumber, "error reading input", "", err)
	}

	return lines, nil
}

// ParseWithRecovery parses lines and continues after errors, returning both lines and errors.
func (p *Parser) ParseWithRecovery(r io.Reader) ([]*Line, []error) {
	p.Reset()

	scanner := bufio.NewScanner(r)
	scanner.Split(scanGEDCOMLines)
	var (
		lines    []*Line
		errs     []error
		prevLine string
	)

	for scanner.Scan() {
		text := scanner.Text()
		line, err := p.ParseLine(text)
		if err != nil {
			errs = append(errs, enrichParseError(err, prevLine, text))
			continue
		}
		lines = append(lines, line)
		prevLine = text
	}

	if err := scanner.Err(); err != nil {
		errs = append(errs, wrapParseError(p.lineNumber, "error reading input", "", err))
	}

	return lines, errs
}

func validateTag(tag string) error {
	if tag == "" {
		return &InvalidTagError{Tag: tag, Reason: "empty"}
	}
	if len(tag) > maxTagLength {
		return &InvalidTagError{Tag: tag, Reason: "too long"}
	}
	for _, r := range tag {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return &InvalidTagError{Tag: tag, Reason: "contains invalid characters"}
	}
	return nil
}

func validateXRef(xref string) error {
	if len(xref) <= 2 {
		return &InvalidXRefError{XRef: xref, Reason: "empty"}
	}
	if strings.Count(xref, "@") != 2 {
		return &InvalidXRefError{XRef: xref, Reason: "must start and end with @"}
	}
	return nil
}

// scanGEDCOMLines is a split function for bufio.Scanner that handles
// all GEDCOM line ending styles: LF, CRLF, and CR (old Macintosh).
// This is based on bufio.ScanLines but adds CR-only support.
func scanGEDCOMLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Look for CR or LF
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			// Found LF - this could be standalone or part of CRLF
			return i + 1, data[0:i], nil
		}
		if data[i] == '\r' {
			// Found CR - check if followed by LF (CRLF)
			if i+1 < len(data) {
				if data[i+1] == '\n' {
					// CRLF - return line without either terminator
					return i + 2, data[0:i], nil
				}
				// CR alone - return line
				return i + 1, data[0:i], nil
			}
			// CR at end of data - need more data to determine if CRLF
			if !atEOF {
				return 0, nil, nil
			}
			// At EOF with CR - treat as line ending
			return i + 1, data[0:i], nil
		}
	}

	// If we're at EOF, return remaining data as final line
	if atEOF {
		return len(data), data, nil
	}

	// Request more data
	return 0, nil, nil
}
