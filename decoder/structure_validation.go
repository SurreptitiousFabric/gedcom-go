package decoder

import (
	"strconv"
	"strings"

	"github.com/cacack/gedcom-go/parser"
)

func validateStructure(lines []*parser.Line) []error {
	if len(lines) == 0 {
		return []error{&MissingHeaderError{Line: 0}}
	}

	var (
		hasHead bool
		hasTrlr bool
	)

	for _, line := range lines {
		if line.Level != 0 {
			continue
		}
		if line.Tag == "HEAD" {
			hasHead = true
		}
		if line.Tag == "TRLR" {
			hasTrlr = true
		}
	}

	var errs []error
	if !hasHead {
		first := lines[0]
		errs = append(errs, &MissingHeaderError{
			Line:    first.LineNumber,
			Context: formatLineContext(first),
		})
	}
	if !hasTrlr {
		last := lines[len(lines)-1]
		errs = append(errs, &MissingTrailerError{
			Line:    last.LineNumber,
			Context: formatLineContext(last),
		})
	}

	return errs
}

func formatLineContext(line *parser.Line) string {
	if line == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(strconv.Itoa(line.Level))
	b.WriteString(" ")
	if line.XRef != "" {
		b.WriteString(line.XRef)
		b.WriteString(" ")
	}
	b.WriteString(line.Tag)
	if line.Value != "" {
		b.WriteString(" ")
		b.WriteString(line.Value)
	}

	return b.String()
}
