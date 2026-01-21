package validator

import "github.com/cacack/gedcom-go/gedcom"

func validateV70Rules(doc *gedcom.Document) []error {
	deprecated := map[string]string{
		"AFN":   "deprecated in GEDCOM 7.0",
		"EMAIL": "deprecated in GEDCOM 7.0",
		"FAX":   "deprecated in GEDCOM 7.0",
		"RFN":   "deprecated in GEDCOM 7.0",
		"REFN":  "deprecated in GEDCOM 7.0",
		"RIN":   "deprecated in GEDCOM 7.0",
		"WWW":   "deprecated in GEDCOM 7.0",
	}
	return validateDeprecatedTags(doc, gedcom.Version70, deprecated)
}
