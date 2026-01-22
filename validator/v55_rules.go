package validator

import "github.com/cacack/gedcom-go/gedcom"

func validateV55Rules(doc *gedcom.Document) []error {
	deprecated := map[string]string{
		"UID":  "introduced in GEDCOM 7.0",
		"CREA": "introduced in GEDCOM 7.0",
		"MIME": "introduced in GEDCOM 7.0",
	}
	return validateDeprecatedTags(doc, gedcom.Version55, deprecated)
}
