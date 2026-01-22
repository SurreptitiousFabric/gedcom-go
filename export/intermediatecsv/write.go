package intermediatecsv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// WriteCSVBundle writes the intermediate CSV bundle to outputDir.
func WriteCSVBundle(model *IntermediateModel, issues []Issue, outputDir string, options Options) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	if err := writePersons(filepath.Join(outputDir, "persons.csv"), model.Persons); err != nil {
		return err
	}
	if err := writeEvents(filepath.Join(outputDir, "events.csv"), model.Events); err != nil {
		return err
	}
	if options.IncludePlaces {
		if err := writePlaces(filepath.Join(outputDir, "places.csv"), model.Places); err != nil {
			return err
		}
	}
	if options.IncludeGroups {
		if err := writeGroups(filepath.Join(outputDir, "groups.csv"), model.Groups); err != nil {
			return err
		}
	}
	if err := writePersonEventLinks(filepath.Join(outputDir, "person_event_links.csv"), model.PersonEventLinks); err != nil {
		return err
	}
	if err := writePersonParentLinks(filepath.Join(outputDir, "person_parent_links.csv"), model.PersonParentLinks); err != nil {
		return err
	}
	if options.IncludeGroups {
		if err := writeGroupPersonLinks(filepath.Join(outputDir, "group_person_links.csv"), model.GroupPersonLinks); err != nil {
			return err
		}
	}
	if options.IncludeSources {
		if err := writeSources(filepath.Join(outputDir, "sources.csv"), model.Sources); err != nil {
			return err
		}
		if err := writeCitations(filepath.Join(outputDir, "citations.csv"), model.Citations); err != nil {
			return err
		}
		if err := writeEntityCitationLinks(filepath.Join(outputDir, "entity_citation_links.csv"), model.EntityCitationLinks); err != nil {
			return err
		}
	}
	if err := writeIssues(filepath.Join(outputDir, "issues.csv"), issues); err != nil {
		return err
	}

	return nil
}

func writePersons(path string, rows []PersonRow) error {
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.Key, row.GedcomXRef, row.Title, row.Sex, row.PrimaryName, row.NotesRaw})
	}
	return writeCSV(path, []string{"key", "gedcom_xref", "title", "sex", "primary_name", "notes_raw"}, data)
}

func writeEvents(path string, rows []EventRow) error {
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{
			row.Key,
			row.GedcomXRef,
			row.Title,
			row.Type,
			row.Subtype,
			row.WhenValue,
			row.PlaceKey,
			row.GedcomPath,
			row.RawDate,
			row.RawPlace,
		})
	}
	return writeCSV(path, []string{"key", "gedcom_xref", "title", "type", "subtype", "when_value", "place_key", "gedcom_path", "raw_date", "raw_place"}, data)
}

func writePlaces(path string, rows []PlaceRow) error {
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.Key, row.RawPlace, row.NormalizedPlace, row.Lat, row.Lon, row.NotesRaw})
	}
	return writeCSV(path, []string{"key", "raw_place", "normalized_place", "lat", "lon", "notes_raw"}, data)
}

func writeGroups(path string, rows []GroupRow) error {
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.Key, row.GedcomXRef, row.Title, row.Type, row.NotesRaw})
	}
	return writeCSV(path, []string{"key", "gedcom_xref", "title", "type", "notes_raw"}, data)
}

func writePersonEventLinks(path string, rows []PersonEventLink) error {
	rows = dedupePersonEventLinks(rows)
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.PersonKey, row.Role, row.EventKey, strconv.Itoa(row.Seq), row.GedcomPath})
	}
	return writeCSV(path, []string{"person_key", "role", "event_key", "seq", "gedcom_path"}, data)
}

func writePersonParentLinks(path string, rows []PersonParentLink) error {
	rows = dedupePersonParentLinks(rows)
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.ChildPersonKey, row.ParentType, row.ParentPersonKey, row.GedcomPath})
	}
	return writeCSV(path, []string{"child_person_key", "parent_type", "parent_person_key", "gedcom_path"}, data)
}

func writeGroupPersonLinks(path string, rows []GroupPersonLink) error {
	rows = dedupeGroupPersonLinks(rows)
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.GroupKey, row.PersonKey, row.Role, strconv.Itoa(row.Seq)})
	}
	return writeCSV(path, []string{"group_key", "person_key", "role", "seq"}, data)
}

func writeSources(path string, rows []SourceRow) error {
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.Key, row.Title, row.Author, row.Publication, row.Date, row.RawSource})
	}
	return writeCSV(path, []string{"key", "title", "author", "publication", "date", "raw_source"}, data)
}

func writeCitations(path string, rows []CitationRow) error {
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.Key, row.SourceKey, row.Detail, row.Quote, row.GedcomPath})
	}
	return writeCSV(path, []string{"key", "source_key", "detail", "quote", "gedcom_path"}, data)
}

func writeEntityCitationLinks(path string, rows []EntityCitationLink) error {
	rows = dedupeEntityCitationLinks(rows)
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{row.EntityType, row.EntityKey, row.CitationKey, strconv.Itoa(row.Seq)})
	}
	return writeCSV(path, []string{"entity_type", "entity_key", "citation_key", "seq"}, data)
}

func writeIssues(path string, rows []Issue) error {
	rows = append([]Issue{}, rows...)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Severity != rows[j].Severity {
			return rows[i].Severity < rows[j].Severity
		}
		if rows[i].EntityType != rows[j].EntityType {
			return rows[i].EntityType < rows[j].EntityType
		}
		if rows[i].EntityKey != rows[j].EntityKey {
			return rows[i].EntityKey < rows[j].EntityKey
		}
		return rows[i].IssueCode < rows[j].IssueCode
	})
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{
			row.Severity,
			row.EntityType,
			row.EntityKey,
			row.GedcomXRef,
			row.GedcomPath,
			row.IssueCode,
			row.Message,
			row.RawValue,
			row.SuggestedAction,
		})
	}
	return writeCSV(path, []string{"severity", "entity_type", "entity_key", "gedcom_xref", "gedcom_path", "issue_code", "message", "raw_value", "suggested_action"}, data)
}

func writeCSV(path string, header []string, rows [][]string) error {
	f, err := os.Create(path) // #nosec G304 -- CLI tool accepts user-provided paths
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header for %s: %w", path, err)
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row for %s: %w", path, err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush %s: %w", path, err)
	}
	return nil
}

func dedupePersonEventLinks(rows []PersonEventLink) []PersonEventLink {
	seen := make(map[string]struct{})
	result := make([]PersonEventLink, 0, len(rows))
	for _, row := range rows {
		key := row.PersonKey + "|" + row.Role + "|" + row.EventKey
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, row)
	}
	return result
}

func dedupePersonParentLinks(rows []PersonParentLink) []PersonParentLink {
	seen := make(map[string]struct{})
	result := make([]PersonParentLink, 0, len(rows))
	for _, row := range rows {
		key := row.ChildPersonKey + "|" + row.ParentType + "|" + row.ParentPersonKey
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, row)
	}
	return result
}

func dedupeGroupPersonLinks(rows []GroupPersonLink) []GroupPersonLink {
	seen := make(map[string]struct{})
	result := make([]GroupPersonLink, 0, len(rows))
	for _, row := range rows {
		key := row.GroupKey + "|" + row.PersonKey + "|" + row.Role
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, row)
	}
	return result
}

func dedupeEntityCitationLinks(rows []EntityCitationLink) []EntityCitationLink {
	seen := make(map[string]struct{})
	result := make([]EntityCitationLink, 0, len(rows))
	for _, row := range rows {
		key := row.EntityType + "|" + row.EntityKey + "|" + row.CitationKey
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, row)
	}
	return result
}
