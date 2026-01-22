package decoder

import (
	"strings"

	"github.com/cacack/gedcom-go/parser"
)

func validateStrictTags(lines []*parser.Line) []error {
	var errs []error
	for _, line := range lines {
		if line == nil {
			continue
		}
		if strings.HasPrefix(line.Tag, "_") {
			errs = append(errs, &NonStandardTagError{
				Line:    line.LineNumber,
				Tag:     line.Tag,
				Context: formatLineContext(line),
			})
		}
	}
	return errs
}
