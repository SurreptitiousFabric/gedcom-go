package decoder

import (
	"strings"

	"github.com/cacack/gedcom-go/gedcom"
)

func validateXRefs(doc *gedcom.Document) []error {
	if doc == nil {
		return nil
	}

	var errs []error
	for _, record := range doc.Records {
		for _, tag := range record.Tags {
			if !isXRefValue(tag.Value) {
				continue
			}
			if _, ok := doc.XRefMap[tag.Value]; ok {
				continue
			}
			errs = append(errs, &BrokenXRefError{
				XRef:       tag.Value,
				Line:       tag.LineNumber,
				Tag:        tag.Tag,
				RecordXRef: record.XRef,
				Context:    strings.TrimSpace(tag.Tag + " " + tag.Value),
			})
		}
	}

	return errs
}

func isXRefValue(value string) bool {
	if !strings.HasPrefix(value, "@") || !strings.HasSuffix(value, "@") {
		return false
	}
	if len(value) < 3 {
		return false
	}
	if value == "@VOID@" {
		return false
	}
	return true
}
