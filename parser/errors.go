package parser

import "fmt"

// ParseError represents an error that occurred during parsing.
// It includes line number and context for better error reporting.
type ParseError struct {
	// Line is the line number where the error occurred (1-based)
	Line int

	// Message describes what went wrong
	Message string

	// Context provides the actual line content that caused the error
	Context string

	// Err is the underlying error, if any
	Err error
}

// InvalidTagError reports a tag format violation.
type InvalidTagError struct {
	Tag    string
	Reason string
}

func (e *InvalidTagError) Error() string {
	return fmt.Sprintf("invalid tag %q: %s", e.Tag, e.Reason)
}

// InvalidLevelError reports a level format violation.
type InvalidLevelError struct {
	Raw    string
	Reason string
}

func (e *InvalidLevelError) Error() string {
	return fmt.Sprintf("invalid level %q: %s", e.Raw, e.Reason)
}

// LevelMismatchError reports a malformed level jump.
type LevelMismatchError struct {
	Previous int
	Current  int
}

func (e *LevelMismatchError) Error() string {
	return fmt.Sprintf("level jump from %d to %d", e.Previous, e.Current)
}

// InvalidXRefError reports a malformed cross-reference identifier.
type InvalidXRefError struct {
	XRef   string
	Reason string
}

func (e *InvalidXRefError) Error() string {
	return fmt.Sprintf("invalid xref %q: %s", e.XRef, e.Reason)
}

func (e *ParseError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("line %d: %s (context: %q)", e.Line, e.Message, e.Context)
	}
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// newParseError creates a new ParseError with the given details.
func newParseError(line int, message, context string) error {
	return &ParseError{
		Line:    line,
		Message: message,
		Context: context,
	}
}

// wrapParseError wraps an existing error with parse context.
func wrapParseError(line int, message, context string, err error) error {
	return &ParseError{
		Line:    line,
		Message: message,
		Context: context,
		Err:     err,
	}
}

func enrichParseError(err error, prevLine, currentLine string) error {
	parseErr, ok := err.(*ParseError)
	if !ok {
		return err
	}

	context := currentLine
	if prevLine != "" {
		context = fmt.Sprintf("prev: %s | line: %s", prevLine, currentLine)
	}

	return &ParseError{
		Line:    parseErr.Line,
		Message: parseErr.Message,
		Context: context,
		Err:     parseErr.Err,
	}
}
