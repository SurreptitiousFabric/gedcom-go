# Validation Policy

## Deprecated Tags

The validator applies version-specific checks for tags that are deprecated or not valid
for a given GEDCOM version. These checks run only when the parsed document header
has a recognized version (5.5, 5.5.1, or 7.0). If the header is missing or unknown,
version-specific checks are skipped.

Each deprecated tag results in a validation error with code `DEPRECATED_TAG` and a
message of the form:

"Tag <TAG> is not valid in GEDCOM <VERSION>: <REASON>"

Current deprecated tag lists:

- GEDCOM 5.5 and 5.5.1:
  - UID (introduced in GEDCOM 7.0)
  - CREA (introduced in GEDCOM 7.0)
  - MIME (introduced in GEDCOM 7.0)

- GEDCOM 7.0:
  - AFN (deprecated in GEDCOM 7.0)
  - RFN (deprecated in GEDCOM 7.0)
  - REFN (deprecated in GEDCOM 7.0)
  - RIN (deprecated in GEDCOM 7.0)
  - EMAIL (deprecated in GEDCOM 7.0)
  - FAX (deprecated in GEDCOM 7.0)
  - WWW (deprecated in GEDCOM 7.0)

## Updating the Lists

When you add or remove deprecated tags:

- Update the maps in:
  - `validator/v55_rules.go`
  - `validator/v551_rules.go`
  - `validator/v70_rules.go`
- Adjust tests in `validator/validator_test.go`.
- Update `testdata/validator/deprecated-tag-70.ged` when GEDCOM 7.0 tags change.

