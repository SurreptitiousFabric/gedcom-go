package intermediatecsv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cacack/gedcom-go/gedcom"
)

func TestKeyGeneration(t *testing.T) {
	if got := personKey("@I1@"); got != "p_I1" {
		t.Fatalf("personKey mismatch: %s", got)
	}
	if got := groupKey("@F12@"); got != "g_F12" {
		t.Fatalf("groupKey mismatch: %s", got)
	}
	if got := eventKey("@I1@", "BIRT", 1); got != "e_I1_BIRT_1" {
		t.Fatalf("eventKey mismatch: %s", got)
	}
	if got := placeKey("Boston, MA"); got == "" || got[:2] != "w_" {
		t.Fatalf("placeKey prefix mismatch: %s", got)
	}
}

func TestDuplicateLinkHandling(t *testing.T) {
	model := &IntermediateModel{
		PersonEventLinks: []PersonEventLink{
			{PersonKey: "p_I1", Role: "SELF", EventKey: "e_I1_BIRT_1"},
			{PersonKey: "p_I1", Role: "SELF", EventKey: "e_I1_BIRT_1"},
		},
	}
	issues := ValidateIntermediate(model)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].IssueCode != "DUPLICATE_LINK" {
		t.Fatalf("expected DUPLICATE_LINK, got %s", issues[0].IssueCode)
	}
}

func TestUnresolvedPointerDetection(t *testing.T) {
	family := &gedcom.Family{
		XRef:    "@F1@",
		Husband: "@I999@",
	}
	record := &gedcom.Record{XRef: family.XRef, Type: gedcom.RecordTypeFamily, Entity: family}
	doc := &gedcom.Document{
		Records: []*gedcom.Record{record},
		XRefMap: map[string]*gedcom.Record{family.XRef: record},
	}

	model := BuildIntermediateModel(doc)
	issues := ValidateIntermediate(model)
	if len(issues) == 0 {
		t.Fatalf("expected unresolved pointer issue")
	}
	found := false
	for _, issue := range issues {
		if issue.IssueCode == "UNRESOLVED_POINTER" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected UNRESOLVED_POINTER issue")
	}
}

func TestBuildAndWriteIntermediateBundle(t *testing.T) {
	note := &gedcom.Note{XRef: "@N1@", Text: "note text"}
	source := &gedcom.Source{XRef: "@S1@", Title: "Register", Author: "Clerk", Publication: "Ledger", Text: "raw source"}
	birthDate, _ := gedcom.ParseDate("1 JAN 1900")
	occupationDate, _ := gedcom.ParseDate("1901")
	marriageDate, _ := gedcom.ParseDate("2 FEB 1920")
	individual1 := &gedcom.Individual{
		XRef:  "@I1@",
		Names: []*gedcom.PersonalName{{Full: "John /Doe/"}},
		Sex:   "M",
		Events: []*gedcom.Event{{
			Type:       gedcom.EventBirth,
			Date:       "1 JAN 1900",
			ParsedDate: birthDate,
			PlaceDetail: &gedcom.PlaceDetail{
				Name:        "Boston, MA",
				Form:        "City, State",
				Coordinates: &gedcom.Coordinates{Latitude: "N42.3601", Longitude: "W71.0589"},
			},
			SourceCitations: []*gedcom.SourceCitation{{
				SourceXRef: "@S1@",
				Page:       "p1",
				Data:       &gedcom.SourceCitationData{Text: "birth quote"},
			}},
		}},
		Attributes: []*gedcom.Attribute{{
			Type:       "OCCU",
			Value:      "Clerk",
			Date:       "1901",
			ParsedDate: occupationDate,
			Place:      "New York, NY",
		}},
		Notes: []string{"@N1@"},
		SourceCitations: []*gedcom.SourceCitation{{
			SourceXRef: "@S1@",
			Page:       "p2",
			Data:       &gedcom.SourceCitationData{Text: "person quote"},
		}},
	}
	individual2 := &gedcom.Individual{XRef: "@I2@", Names: []*gedcom.PersonalName{{Full: "Jane /Doe/"}}}
	individual3 := &gedcom.Individual{
		XRef:            "@I3@",
		Names:           []*gedcom.PersonalName{{Full: "Child /Doe/"}},
		ChildInFamilies: []gedcom.FamilyLink{{FamilyXRef: "@F1@"}},
	}
	family := &gedcom.Family{
		XRef:     "@F1@",
		Husband:  "@I1@",
		Wife:     "@I2@",
		Children: []string{"@I3@"},
		Events: []*gedcom.Event{{
			Type:       gedcom.EventMarriage,
			Date:       "2 FEB 1920",
			ParsedDate: marriageDate,
		}},
	}
	doc := buildDocument([]*gedcom.Individual{individual1, individual2, individual3}, []*gedcom.Family{family}, []*gedcom.Source{source}, []*gedcom.Note{note})

	model := BuildIntermediateModel(doc)
	issues := ValidateIntermediate(model)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
	if len(model.Persons) != 3 {
		t.Fatalf("expected 3 persons, got %d", len(model.Persons))
	}
	if len(model.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(model.Events))
	}
	if len(model.Places) != 2 {
		t.Fatalf("expected 2 places, got %d", len(model.Places))
	}

	outputDir := t.TempDir()
	options := Options{IncludeSources: true, IncludePlaces: true, IncludeGroups: true}
	if err := WriteCSVBundle(model, issues, outputDir, options); err != nil {
		t.Fatalf("WriteCSVBundle error: %v", err)
	}
	assertCSVExists(t, outputDir, "persons.csv", "key,gedcom_xref,title,sex,primary_name,notes_raw")
	assertCSVExists(t, outputDir, "events.csv", "key,gedcom_xref,title,type,subtype,when_value,place_key,gedcom_path,raw_date,raw_place")
	assertCSVExists(t, outputDir, "issues.csv", "severity,entity_type,entity_key,gedcom_xref,gedcom_path,issue_code,message,raw_value,suggested_action")
}

func TestInvalidDateIssue(t *testing.T) {
	individual := &gedcom.Individual{
		XRef: "@I1@",
		Events: []*gedcom.Event{{
			Type: gedcom.EventBirth,
			Date: "INVALID",
		}},
	}
	doc := buildDocument([]*gedcom.Individual{individual}, nil, nil, nil)

	model := BuildIntermediateModel(doc)
	issues := ValidateIntermediate(model)
	found := false
	for _, issue := range issues {
		if issue.IssueCode == "INVALID_DATE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected INVALID_DATE issue")
	}
}

func buildDocument(individuals []*gedcom.Individual, families []*gedcom.Family, sources []*gedcom.Source, notes []*gedcom.Note) *gedcom.Document {
	var records []*gedcom.Record
	xrefMap := make(map[string]*gedcom.Record)
	addRecord := func(xref string, recordType gedcom.RecordType, entity interface{}) {
		record := &gedcom.Record{XRef: xref, Type: recordType, Entity: entity}
		records = append(records, record)
		if xref != "" {
			xrefMap[xref] = record
		}
	}
	for _, individual := range individuals {
		addRecord(individual.XRef, gedcom.RecordTypeIndividual, individual)
	}
	for _, family := range families {
		addRecord(family.XRef, gedcom.RecordTypeFamily, family)
	}
	for _, source := range sources {
		addRecord(source.XRef, gedcom.RecordTypeSource, source)
	}
	for _, note := range notes {
		addRecord(note.XRef, gedcom.RecordTypeNote, note)
	}
	return &gedcom.Document{Records: records, XRefMap: xrefMap}
}

func assertCSVExists(t *testing.T, dir, name, header string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 || lines[0] != header {
		t.Fatalf("unexpected header for %s: %s", name, lines[0])
	}
}
