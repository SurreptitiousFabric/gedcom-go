package validator

import (
	"fmt"

	"github.com/cacack/gedcom-go/gedcom"
)

func (v *Validator) validateXRefFormats(doc *gedcom.Document) {
	for _, record := range doc.Records {
		if record.XRef == "" {
			continue
		}
		if isStandardXRef(record.XRef) {
			continue
		}
		v.errors = append(v.errors, &ValidationError{
			Code:    "NON_STANDARD_XREF",
			Message: fmt.Sprintf("Non-standard XRef format %s", record.XRef),
			Line:    record.LineNumber,
			XRef:    record.XRef,
		})
	}
}

func isStandardXRef(xref string) bool {
	if len(xref) < 3 {
		return false
	}
	if xref[0] != '@' || xref[len(xref)-1] != '@' {
		return false
	}
	for i := 1; i < len(xref)-1; i++ {
		b := xref[i]
		if b >= 'A' && b <= 'Z' {
			continue
		}
		if b >= 'a' && b <= 'z' {
			continue
		}
		if b >= '0' && b <= '9' {
			continue
		}
		return false
	}
	return true
}
