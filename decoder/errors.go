package decoder

import "fmt"

// DecodeErrors collects multiple decode-related errors.
type DecodeErrors struct {
	Errors []error
}

func (e *DecodeErrors) Error() string {
	if e == nil || len(e.Errors) == 0 {
		return "no decode errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d decode errors: %s", len(e.Errors), e.Errors[0].Error())
}

func (e *DecodeErrors) Unwrap() []error {
	return e.Errors
}

// BrokenXRefError reports a missing cross-reference target.
type BrokenXRefError struct {
	XRef       string
	Line       int
	Tag        string
	RecordXRef string
	Context    string
}

func (e *BrokenXRefError) Error() string {
	if e.RecordXRef != "" {
		if e.Context != "" {
			return fmt.Sprintf("line %d: broken reference %s in %s (record %s) (context: %q)", e.Line, e.XRef, e.Tag, e.RecordXRef, e.Context)
		}
		return fmt.Sprintf("line %d: broken reference %s in %s (record %s)", e.Line, e.XRef, e.Tag, e.RecordXRef)
	}
	if e.Context != "" {
		return fmt.Sprintf("line %d: broken reference %s in %s (context: %q)", e.Line, e.XRef, e.Tag, e.Context)
	}
	return fmt.Sprintf("line %d: broken reference %s in %s", e.Line, e.XRef, e.Tag)
}

// MissingHeaderError reports a missing HEAD record.
type MissingHeaderError struct {
	Line    int
	Context string
}

func (e *MissingHeaderError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("line %d: missing HEAD record (context: %q)", e.Line, e.Context)
	}
	return fmt.Sprintf("line %d: missing HEAD record", e.Line)
}

// MissingTrailerError reports a missing TRLR record.
type MissingTrailerError struct {
	Line    int
	Context string
}

func (e *MissingTrailerError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("line %d: missing TRLR record (context: %q)", e.Line, e.Context)
	}
	return fmt.Sprintf("line %d: missing TRLR record", e.Line)
}
