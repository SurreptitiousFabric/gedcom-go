package validator

import (
	"fmt"

	"github.com/cacack/gedcom-go/gedcom"
)

func (v *Validator) validateVersionSpecific(doc *gedcom.Document) {
	if doc.Header == nil {
		return
	}
	var errs []error
	switch doc.Header.Version {
	case gedcom.Version55:
		errs = validateV55Rules(doc)
	case gedcom.Version551:
		errs = validateV551Rules(doc)
	case gedcom.Version70:
		errs = validateV70Rules(doc)
	default:
		return
	}
	v.errors = append(v.errors, errs...)
}

func validateDeprecatedTags(doc *gedcom.Document, version gedcom.Version, deprecated map[string]string) []error {
	var errs []error
	for _, record := range doc.Records {
		if record == nil {
			continue
		}
		if reason, ok := deprecated[string(record.Type)]; ok {
			errs = append(errs, &ValidationError{
				Code:    "DEPRECATED_TAG",
				Message: fmt.Sprintf("Tag %s is not valid in GEDCOM %s: %s", record.Type, version, reason),
				Line:    record.LineNumber,
				XRef:    record.XRef,
			})
		}
		for _, tag := range record.Tags {
			if reason, ok := deprecated[tag.Tag]; ok {
				errs = append(errs, &ValidationError{
					Code:    "DEPRECATED_TAG",
					Message: fmt.Sprintf("Tag %s is not valid in GEDCOM %s: %s", tag.Tag, version, reason),
					Line:    tag.LineNumber,
					XRef:    record.XRef,
				})
			}
		}
	}
	return errs
}
