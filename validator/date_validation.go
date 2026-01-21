package validator

import (
	"fmt"
	"strings"

	"github.com/cacack/gedcom-go/gedcom"
)

func (v *Validator) validateDates(doc *gedcom.Document) {
	for _, record := range doc.Records {
		for _, tag := range record.Tags {
			if tag.Tag != "DATE" {
				continue
			}
			value := strings.TrimSpace(tag.Value)
			if value == "" {
				continue
			}
			parsed, err := gedcom.ParseDate(value)
			if err != nil {
				v.errors = append(v.errors, &ValidationError{
					Code:    "INVALID_DATE",
					Message: fmt.Sprintf("Invalid date %q", value),
					Line:    tag.LineNumber,
				})
				continue
			}
			if err := parsed.Validate(); err != nil {
				v.errors = append(v.errors, &ValidationError{
					Code:    "INVALID_DATE",
					Message: fmt.Sprintf("Invalid date %q: %s", value, err.Error()),
					Line:    tag.LineNumber,
				})
			}
		}
	}
}
