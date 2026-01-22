package intermediatecsv

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/cacack/gedcom-go/gedcom"
)

// Options controls which optional CSV files are written.
type Options struct {
	IncludeSources bool
	IncludePlaces  bool
	IncludeGroups  bool
}

// IntermediateModel is the normalized export-ready view of a GEDCOM document.
type IntermediateModel struct {
	Persons                []PersonRow
	Events                 []EventRow
	Places                 []PlaceRow
	Groups                 []GroupRow
	PersonEventLinks       []PersonEventLink
	PersonParentLinks      []PersonParentLink
	GroupPersonLinks       []GroupPersonLink
	Sources                []SourceRow
	Citations              []CitationRow
	EntityCitationLinks    []EntityCitationLink
	Issues                 []Issue
	PlaceKeyByRaw          map[string]string
	PlaceByKey             map[string]PlaceRow
	SourceKeyByXRef        map[string]string
	CitationKeyByIdentity  map[string]string
	EventKeyByGedcomPath   map[string]string
	PersonKeyByXRef        map[string]string
	GroupKeyByXRef         map[string]string
	SourceCitationIdentity map[string]SourceCitationIdentity
}

// PersonRow maps to persons.csv.
type PersonRow struct {
	Key         string
	GedcomXRef  string
	Title       string
	Sex         string
	PrimaryName string
	NotesRaw    string
}

// EventRow maps to events.csv.
type EventRow struct {
	Key        string
	GedcomXRef string
	Title      string
	Type       string
	Subtype    string
	WhenValue  string
	PlaceKey   string
	GedcomPath string
	RawDate    string
	RawPlace   string
}

// PlaceRow maps to places.csv.
type PlaceRow struct {
	Key             string
	RawPlace        string
	NormalizedPlace string
	Lat             string
	Lon             string
	NotesRaw        string
}

// GroupRow maps to groups.csv.
type GroupRow struct {
	Key        string
	GedcomXRef string
	Title      string
	Type       string
	NotesRaw   string
}

// PersonEventLink maps to person_event_links.csv.
type PersonEventLink struct {
	PersonKey  string
	Role       string
	EventKey   string
	Seq        int
	GedcomPath string
}

// PersonParentLink maps to person_parent_links.csv.
type PersonParentLink struct {
	ChildPersonKey  string
	ParentType      string
	ParentPersonKey string
	GedcomPath      string
}

// GroupPersonLink maps to group_person_links.csv.
type GroupPersonLink struct {
	GroupKey  string
	PersonKey string
	Role      string
	Seq       int
}

// SourceRow maps to sources.csv.
type SourceRow struct {
	Key         string
	Title       string
	Author      string
	Publication string
	Date        string
	RawSource   string
}

// CitationRow maps to citations.csv.
type CitationRow struct {
	Key        string
	SourceKey  string
	Detail     string
	Quote      string
	GedcomPath string
}

// EntityCitationLink maps to entity_citation_links.csv.
type EntityCitationLink struct {
	EntityType  string
	EntityKey   string
	CitationKey string
	Seq         int
}

// Issue maps to issues.csv.
type Issue struct {
	Severity        string
	EntityType      string
	EntityKey       string
	GedcomXRef      string
	GedcomPath      string
	IssueCode       string
	Message         string
	RawValue        string
	SuggestedAction string
}

// SourceCitationIdentity captures the uniqueness identity for a citation.
type SourceCitationIdentity struct {
	SourceXRef string
	Detail     string
	Quote      string
	Path       string
}

// BuildIntermediateModel constructs a normalized model for CSV export.
func BuildIntermediateModel(doc *gedcom.Document) *IntermediateModel {
	model := &IntermediateModel{
		PlaceKeyByRaw:          make(map[string]string),
		PlaceByKey:             make(map[string]PlaceRow),
		SourceKeyByXRef:        make(map[string]string),
		CitationKeyByIdentity:  make(map[string]string),
		EventKeyByGedcomPath:   make(map[string]string),
		PersonKeyByXRef:        make(map[string]string),
		GroupKeyByXRef:         make(map[string]string),
		SourceCitationIdentity: make(map[string]SourceCitationIdentity),
	}
	if doc == nil {
		return model
	}

	ensureXRefMap(doc)

	individuals := doc.Individuals()
	families := doc.Families()
	allNotes := buildNoteLookup(doc)
	allSources := doc.Sources()

	for _, source := range allSources {
		if source == nil {
			continue
		}
		key := sourceKey(source)
		model.SourceKeyByXRef[source.XRef] = key
		model.Sources = append(model.Sources, SourceRow{
			Key:         key,
			Title:       source.Title,
			Author:      source.Author,
			Publication: source.Publication,
			Date:        sourceDate(source),
			RawSource:   source.Text,
		})
	}

	for _, individual := range individuals {
		if individual == nil {
			continue
		}
		personKey := personKey(individual.XRef)
		model.PersonKeyByXRef[individual.XRef] = personKey
		primaryName := primaryName(individual)
		title := primaryName
		notesRaw := collectNotes(individual.Notes, allNotes, &model.Issues, "person", personKey, individual.XRef, gedcomPathForPerson(individual.XRef, "NOTE"))
		model.Persons = append(model.Persons, PersonRow{
			Key:         personKey,
			GedcomXRef:  individual.XRef,
			Title:       title,
			Sex:         individual.Sex,
			PrimaryName: primaryName,
			NotesRaw:    notesRaw,
		})

		eventCounter := make(map[string]int)
		for _, event := range individual.Events {
			if event == nil {
				continue
			}
			model.addEventForIndividual(individual, event, eventCounter)
		}
		for _, attribute := range individual.Attributes {
			if attribute == nil {
				continue
			}
			model.addAttributeForIndividual(individual, attribute, eventCounter)
		}

		for _, familyLink := range individual.ChildInFamilies {
			if familyLink.FamilyXRef == "" {
				continue
			}
			if doc.GetFamily(familyLink.FamilyXRef) == nil {
				model.Issues = append(model.Issues, Issue{
					Severity:        "ERROR",
					EntityType:      "person",
					EntityKey:       personKey,
					GedcomXRef:      individual.XRef,
					GedcomPath:      gedcomPathForPerson(individual.XRef, "FAMC"),
					IssueCode:       "UNRESOLVED_POINTER",
					Message:         "Child-in-family reference points to missing family record",
					RawValue:        familyLink.FamilyXRef,
					SuggestedAction: "Verify the family xref or add the missing family record.",
				})
			}
		}

		model.collectCitationsForPerson(individual)
	}

	for _, family := range families {
		if family == nil {
			continue
		}
		groupKey := groupKey(family.XRef)
		model.GroupKeyByXRef[family.XRef] = groupKey
		notesRaw := collectNotes(family.Notes, allNotes, &model.Issues, "group", groupKey, family.XRef, gedcomPathForGroup(family.XRef, "NOTE"))
		model.Groups = append(model.Groups, GroupRow{
			Key:        groupKey,
			GedcomXRef: family.XRef,
			Title:      "",
			Type:       string(gedcom.RecordTypeFamily),
			NotesRaw:   notesRaw,
		})

		eventCounter := make(map[string]int)
		for _, event := range family.Events {
			if event == nil {
				continue
			}
			model.addEventForFamily(family, event, eventCounter)
		}

		model.collectGroupLinks(family)
		model.collectParentLinks(family)
		model.collectCitationsForFamily(family)
	}

	sort.Slice(model.Persons, func(i, j int) bool { return model.Persons[i].Key < model.Persons[j].Key })
	sort.Slice(model.Events, func(i, j int) bool { return model.Events[i].Key < model.Events[j].Key })
	sort.Slice(model.Places, func(i, j int) bool { return model.Places[i].Key < model.Places[j].Key })
	sort.Slice(model.Groups, func(i, j int) bool { return model.Groups[i].Key < model.Groups[j].Key })
	sort.Slice(model.PersonEventLinks, func(i, j int) bool {
		if model.PersonEventLinks[i].PersonKey != model.PersonEventLinks[j].PersonKey {
			return model.PersonEventLinks[i].PersonKey < model.PersonEventLinks[j].PersonKey
		}
		if model.PersonEventLinks[i].EventKey != model.PersonEventLinks[j].EventKey {
			return model.PersonEventLinks[i].EventKey < model.PersonEventLinks[j].EventKey
		}
		return model.PersonEventLinks[i].Role < model.PersonEventLinks[j].Role
	})
	sort.Slice(model.PersonParentLinks, func(i, j int) bool {
		if model.PersonParentLinks[i].ChildPersonKey != model.PersonParentLinks[j].ChildPersonKey {
			return model.PersonParentLinks[i].ChildPersonKey < model.PersonParentLinks[j].ChildPersonKey
		}
		if model.PersonParentLinks[i].ParentPersonKey != model.PersonParentLinks[j].ParentPersonKey {
			return model.PersonParentLinks[i].ParentPersonKey < model.PersonParentLinks[j].ParentPersonKey
		}
		return model.PersonParentLinks[i].ParentType < model.PersonParentLinks[j].ParentType
	})
	sort.Slice(model.GroupPersonLinks, func(i, j int) bool {
		if model.GroupPersonLinks[i].GroupKey != model.GroupPersonLinks[j].GroupKey {
			return model.GroupPersonLinks[i].GroupKey < model.GroupPersonLinks[j].GroupKey
		}
		if model.GroupPersonLinks[i].PersonKey != model.GroupPersonLinks[j].PersonKey {
			return model.GroupPersonLinks[i].PersonKey < model.GroupPersonLinks[j].PersonKey
		}
		return model.GroupPersonLinks[i].Role < model.GroupPersonLinks[j].Role
	})
	sort.Slice(model.Sources, func(i, j int) bool { return model.Sources[i].Key < model.Sources[j].Key })
	sort.Slice(model.Citations, func(i, j int) bool { return model.Citations[i].Key < model.Citations[j].Key })
	sort.Slice(model.EntityCitationLinks, func(i, j int) bool {
		if model.EntityCitationLinks[i].EntityType != model.EntityCitationLinks[j].EntityType {
			return model.EntityCitationLinks[i].EntityType < model.EntityCitationLinks[j].EntityType
		}
		if model.EntityCitationLinks[i].EntityKey != model.EntityCitationLinks[j].EntityKey {
			return model.EntityCitationLinks[i].EntityKey < model.EntityCitationLinks[j].EntityKey
		}
		return model.EntityCitationLinks[i].CitationKey < model.EntityCitationLinks[j].CitationKey
	})

	return model
}

// ValidateIntermediate checks for data issues and returns any issues found.
func ValidateIntermediate(model *IntermediateModel) []Issue {
	if model == nil {
		return nil
	}
	issues := append([]Issue{}, model.Issues...)
	issues = append(issues, detectDuplicateLinks(model.PersonEventLinks, func(link PersonEventLink) string {
		return link.PersonKey + "|" + link.Role + "|" + link.EventKey
	}, "person_event_link")...)
	issues = append(issues, detectDuplicateLinks(model.PersonParentLinks, func(link PersonParentLink) string {
		return link.ChildPersonKey + "|" + link.ParentType + "|" + link.ParentPersonKey
	}, "person_parent_link")...)
	issues = append(issues, detectDuplicateLinks(model.GroupPersonLinks, func(link GroupPersonLink) string {
		return link.GroupKey + "|" + link.PersonKey + "|" + link.Role
	}, "group_person_link")...)
	issues = append(issues, detectDuplicateLinks(model.EntityCitationLinks, func(link EntityCitationLink) string {
		return link.EntityType + "|" + link.EntityKey + "|" + link.CitationKey
	}, "entity_citation_link")...)
	return issues
}

func detectDuplicateLinks[T any](links []T, keyFn func(T) string, entityType string) []Issue {
	seen := make(map[string]struct{})
	var issues []Issue
	for _, link := range links {
		key := keyFn(link)
		if _, ok := seen[key]; ok {
			issues = append(issues, Issue{
				Severity:        "WARN",
				EntityType:      entityType,
				EntityKey:       key,
				IssueCode:       "DUPLICATE_LINK",
				Message:         "Duplicate link encountered; keeping first occurrence",
				SuggestedAction: "Remove duplicate links or consolidate the source records.",
			})
			continue
		}
		seen[key] = struct{}{}
	}
	return issues
}

func (m *IntermediateModel) addEventForIndividual(individual *gedcom.Individual, event *gedcom.Event, counter map[string]int) {
	ownerXRef := individual.XRef
	tag := string(event.Type)
	if tag == "" {
		tag = "EVEN"
	}
	index := nextIndex(counter, tag)
	key := eventKey(ownerXRef, tag, index)
	path := gedcomPathForPersonEvent(ownerXRef, tag, index)
	m.EventKeyByGedcomPath[path] = key

	whenValue, ok := normalizeDate(event.ParsedDate)
	if !ok && event.Date != "" {
		m.Issues = append(m.Issues, invalidDateIssue("event", key, ownerXRef, path, event.Date))
	}

	rawPlace, placeKey := m.placeFromEvent(event)
	m.Events = append(m.Events, EventRow{
		Key:        key,
		GedcomXRef: ownerXRef,
		Title:      eventTitle(event, tag),
		Type:       tag,
		Subtype:    event.EventTypeDetail,
		WhenValue:  whenValue,
		PlaceKey:   placeKey,
		GedcomPath: path,
		RawDate:    event.Date,
		RawPlace:   rawPlace,
	})
	if personKey := m.personKey(ownerXRef); personKey != "" {
		m.PersonEventLinks = append(m.PersonEventLinks, PersonEventLink{
			PersonKey:  personKey,
			Role:       "SELF",
			EventKey:   key,
			Seq:        1,
			GedcomPath: path,
		})
	}
	m.collectCitationsForEvent(key, ownerXRef, path, event.SourceCitations)
}

func (m *IntermediateModel) addAttributeForIndividual(individual *gedcom.Individual, attribute *gedcom.Attribute, counter map[string]int) {
	ownerXRef := individual.XRef
	tag := attribute.Type
	if tag == "" {
		tag = "ATTR"
	}
	index := nextIndex(counter, tag)
	key := eventKey(ownerXRef, tag, index)
	path := gedcomPathForPersonEvent(ownerXRef, tag, index)
	m.EventKeyByGedcomPath[path] = key

	whenValue, ok := normalizeDate(attribute.ParsedDate)
	if !ok && attribute.Date != "" {
		m.Issues = append(m.Issues, invalidDateIssue("event", key, ownerXRef, path, attribute.Date))
	}

	rawPlace, placeKey := m.placeFromAttribute(attribute)
	m.Events = append(m.Events, EventRow{
		Key:        key,
		GedcomXRef: ownerXRef,
		Title:      attribute.Value,
		Type:       tag,
		Subtype:    "",
		WhenValue:  whenValue,
		PlaceKey:   placeKey,
		GedcomPath: path,
		RawDate:    attribute.Date,
		RawPlace:   rawPlace,
	})
	if personKey := m.personKey(ownerXRef); personKey != "" {
		m.PersonEventLinks = append(m.PersonEventLinks, PersonEventLink{
			PersonKey:  personKey,
			Role:       "SELF",
			EventKey:   key,
			Seq:        1,
			GedcomPath: path,
		})
	}
	m.collectCitationsForEvent(key, ownerXRef, path, attribute.SourceCitations)
}

func (m *IntermediateModel) addEventForFamily(family *gedcom.Family, event *gedcom.Event, counter map[string]int) {
	ownerXRef := family.XRef
	tag := string(event.Type)
	if tag == "" {
		tag = "EVEN"
	}
	index := nextIndex(counter, tag)
	key := eventKey(ownerXRef, tag, index)
	path := gedcomPathForFamilyEvent(ownerXRef, tag, index)
	m.EventKeyByGedcomPath[path] = key

	whenValue, ok := normalizeDate(event.ParsedDate)
	if !ok && event.Date != "" {
		m.Issues = append(m.Issues, invalidDateIssue("event", key, ownerXRef, path, event.Date))
	}

	rawPlace, placeKey := m.placeFromEvent(event)
	m.Events = append(m.Events, EventRow{
		Key:        key,
		GedcomXRef: ownerXRef,
		Title:      eventTitle(event, tag),
		Type:       tag,
		Subtype:    event.EventTypeDetail,
		WhenValue:  whenValue,
		PlaceKey:   placeKey,
		GedcomPath: path,
		RawDate:    event.Date,
		RawPlace:   rawPlace,
	})

	m.addFamilyEventLinks(family, key, path)
	m.collectCitationsForEvent(key, ownerXRef, path, event.SourceCitations)
}

func (m *IntermediateModel) addFamilyEventLinks(family *gedcom.Family, eventKey, path string) {
	seq := 1
	addLink := func(xref, role string) {
		if xref == "" {
			return
		}
		personKey := m.personKey(xref)
		if personKey == "" {
			m.Issues = append(m.Issues, Issue{
				Severity:        "ERROR",
				EntityType:      "group",
				EntityKey:       m.groupKey(family.XRef),
				GedcomXRef:      family.XRef,
				GedcomPath:      path,
				IssueCode:       "UNRESOLVED_POINTER",
				Message:         "Family event references missing person",
				RawValue:        xref,
				SuggestedAction: "Verify the person xref or add the missing individual record.",
			})
			return
		}
		m.PersonEventLinks = append(m.PersonEventLinks, PersonEventLink{
			PersonKey:  personKey,
			Role:       role,
			EventKey:   eventKey,
			Seq:        seq,
			GedcomPath: path,
		})
		seq++
	}

	addLink(family.Husband, "HUSB")
	addLink(family.Wife, "WIFE")
	for _, child := range family.Children {
		addLink(child, "CHIL")
	}
}

func (m *IntermediateModel) collectGroupLinks(family *gedcom.Family) {
	groupKey := m.groupKey(family.XRef)
	seq := 1
	addGroupLink := func(xref, role string) {
		if xref == "" {
			return
		}
		personKey := m.personKey(xref)
		if personKey == "" {
			m.Issues = append(m.Issues, Issue{
				Severity:        "ERROR",
				EntityType:      "group",
				EntityKey:       groupKey,
				GedcomXRef:      family.XRef,
				GedcomPath:      gedcomPathForGroup(family.XRef, role),
				IssueCode:       "UNRESOLVED_POINTER",
				Message:         "Family member reference points to missing person",
				RawValue:        xref,
				SuggestedAction: "Verify the person xref or add the missing individual record.",
			})
			return
		}
		m.GroupPersonLinks = append(m.GroupPersonLinks, GroupPersonLink{
			GroupKey:  groupKey,
			PersonKey: personKey,
			Role:      role,
			Seq:       seq,
		})
		seq++
	}

	addGroupLink(family.Husband, "HUSB")
	addGroupLink(family.Wife, "WIFE")
	for _, child := range family.Children {
		addGroupLink(child, "CHIL")
	}
}

func (m *IntermediateModel) collectParentLinks(family *gedcom.Family) {
	groupPath := gedcomPathForGroup(family.XRef, "CHIL")
	addParent := func(childXRef, parentXRef, parentType string) {
		if childXRef == "" || parentXRef == "" {
			return
		}
		childKey := m.personKey(childXRef)
		if childKey == "" {
			m.Issues = append(m.Issues, Issue{
				Severity:        "ERROR",
				EntityType:      "group",
				EntityKey:       m.groupKey(family.XRef),
				GedcomXRef:      family.XRef,
				GedcomPath:      groupPath,
				IssueCode:       "UNRESOLVED_POINTER",
				Message:         "Family child reference points to missing person",
				RawValue:        childXRef,
				SuggestedAction: "Verify the child xref or add the missing individual record.",
			})
			return
		}
		parentKey := m.personKey(parentXRef)
		if parentKey == "" {
			m.Issues = append(m.Issues, Issue{
				Severity:        "ERROR",
				EntityType:      "group",
				EntityKey:       m.groupKey(family.XRef),
				GedcomXRef:      family.XRef,
				GedcomPath:      groupPath,
				IssueCode:       "UNRESOLVED_POINTER",
				Message:         "Family parent reference points to missing person",
				RawValue:        parentXRef,
				SuggestedAction: "Verify the parent xref or add the missing individual record.",
			})
			return
		}
		m.PersonParentLinks = append(m.PersonParentLinks, PersonParentLink{
			ChildPersonKey:  childKey,
			ParentType:      parentType,
			ParentPersonKey: parentKey,
			GedcomPath:      groupPath,
		})
	}

	for _, child := range family.Children {
		addParent(child, family.Husband, "HUSB")
		addParent(child, family.Wife, "WIFE")
	}
}

func (m *IntermediateModel) collectCitationsForPerson(individual *gedcom.Individual) {
	if individual == nil {
		return
	}
	personKey := m.personKey(individual.XRef)
	path := gedcomPathForPerson(individual.XRef, "SOUR")
	m.collectCitations(personKey, "person", individual.XRef, path, individual.SourceCitations)
}

func (m *IntermediateModel) collectCitationsForFamily(family *gedcom.Family) {
	if family == nil {
		return
	}
	groupKey := m.groupKey(family.XRef)
	path := gedcomPathForGroup(family.XRef, "SOUR")
	m.collectCitations(groupKey, "group", family.XRef, path, family.SourceCitations)
}

func (m *IntermediateModel) collectCitationsForEvent(eventKey, ownerXRef, path string, citations []*gedcom.SourceCitation) {
	m.collectCitations(eventKey, "event", ownerXRef, path, citations)
}

func (m *IntermediateModel) collectCitations(entityKey, entityType, ownerXRef, path string, citations []*gedcom.SourceCitation) {
	if len(citations) == 0 {
		return
	}
	seq := 1
	for _, citation := range citations {
		if citation == nil {
			continue
		}
		identity := SourceCitationIdentity{
			SourceXRef: citation.SourceXRef,
			Detail:     citation.Page,
			Quote:      citationQuote(citation),
			Path:       path,
		}
		identityKey := citationIdentityKey(identity)
		citationKey, ok := m.CitationKeyByIdentity[identityKey]
		if !ok {
			citationKey = citationKeyFromIdentity(identity)
			m.CitationKeyByIdentity[identityKey] = citationKey
			m.SourceCitationIdentity[citationKey] = identity
			m.Citations = append(m.Citations, CitationRow{
				Key:        citationKey,
				SourceKey:  m.sourceKey(citation.SourceXRef, ownerXRef, path),
				Detail:     citation.Page,
				Quote:      citationQuote(citation),
				GedcomPath: path,
			})
		}
		m.EntityCitationLinks = append(m.EntityCitationLinks, EntityCitationLink{
			EntityType:  entityType,
			EntityKey:   entityKey,
			CitationKey: citationKey,
			Seq:         seq,
		})
		seq++
	}
}

func (m *IntermediateModel) sourceKey(sourceXRef, ownerXRef, path string) string {
	if sourceXRef == "" {
		return ""
	}
	key := m.SourceKeyByXRef[sourceXRef]
	if key == "" {
		m.Issues = append(m.Issues, Issue{
			Severity:        "ERROR",
			EntityType:      "citation",
			EntityKey:       sourceXRef,
			GedcomXRef:      ownerXRef,
			GedcomPath:      path,
			IssueCode:       "UNRESOLVED_POINTER",
			Message:         "Source citation references missing source record",
			RawValue:        sourceXRef,
			SuggestedAction: "Verify the source xref or add the missing source record.",
		})
		return ""
	}
	return key
}

func (m *IntermediateModel) placeFromEvent(event *gedcom.Event) (string, string) {
	if event == nil {
		return "", ""
	}
	if event.PlaceDetail != nil && event.PlaceDetail.Name != "" {
		return m.placeFromDetail(event.PlaceDetail)
	}
	if event.Place != "" {
		return event.Place, m.placeKey(event.Place, "", nil)
	}
	return "", ""
}

func (m *IntermediateModel) placeFromAttribute(attribute *gedcom.Attribute) (string, string) {
	if attribute == nil {
		return "", ""
	}
	if attribute.Place == "" {
		return "", ""
	}
	return attribute.Place, m.placeKey(attribute.Place, "", nil)
}

func (m *IntermediateModel) placeFromDetail(detail *gedcom.PlaceDetail) (string, string) {
	if detail == nil || detail.Name == "" {
		return "", ""
	}
	return detail.Name, m.placeKey(detail.Name, detail.Form, detail.Coordinates)
}

func (m *IntermediateModel) placeKey(rawPlace, form string, coords *gedcom.Coordinates) string {
	key, ok := m.PlaceKeyByRaw[rawPlace]
	if ok {
		return key
	}
	key = placeKey(rawPlace)
	m.PlaceKeyByRaw[rawPlace] = key
	m.PlaceByKey[key] = PlaceRow{
		Key:             key,
		RawPlace:        rawPlace,
		NormalizedPlace: rawPlace,
		Lat:             coordinatesLatitude(coords),
		Lon:             coordinatesLongitude(coords),
		NotesRaw:        "",
	}
	m.Places = append(m.Places, m.PlaceByKey[key])
	return key
}

func (m *IntermediateModel) personKey(xref string) string {
	if xref == "" {
		return ""
	}
	return m.PersonKeyByXRef[xref]
}

func (m *IntermediateModel) groupKey(xref string) string {
	if xref == "" {
		return ""
	}
	return m.GroupKeyByXRef[xref]
}

func ensureXRefMap(doc *gedcom.Document) {
	if doc == nil || doc.XRefMap != nil {
		return
	}
	xrefMap := make(map[string]*gedcom.Record)
	for _, record := range doc.Records {
		if record == nil || record.XRef == "" {
			continue
		}
		xrefMap[record.XRef] = record
	}
	doc.XRefMap = xrefMap
}

func buildNoteLookup(doc *gedcom.Document) map[string]*gedcom.Note {
	notes := make(map[string]*gedcom.Note)
	if doc == nil {
		return notes
	}
	for _, note := range doc.Notes() {
		if note == nil || note.XRef == "" {
			continue
		}
		notes[note.XRef] = note
	}
	return notes
}

func primaryName(individual *gedcom.Individual) string {
	if individual == nil {
		return ""
	}
	if len(individual.Names) > 0 {
		name := individual.Names[0]
		if name == nil {
			return ""
		}
		if name.Full != "" {
			return name.Full
		}
		parts := []string{name.Given, name.Surname}
		return strings.TrimSpace(strings.Join(parts, " "))
	}
	return ""
}

func collectNotes(noteXRefs []string, noteLookup map[string]*gedcom.Note, issues *[]Issue, entityType, entityKey, gedcomXRef, gedcomPath string) string {
	if len(noteXRefs) == 0 {
		return ""
	}
	var parts []string
	for _, noteXRef := range noteXRefs {
		if noteXRef == "" {
			continue
		}
		note := noteLookup[noteXRef]
		if note == nil {
			*issues = append(*issues, Issue{
				Severity:        "ERROR",
				EntityType:      entityType,
				EntityKey:       entityKey,
				GedcomXRef:      gedcomXRef,
				GedcomPath:      gedcomPath,
				IssueCode:       "UNRESOLVED_POINTER",
				Message:         "Note reference points to missing note record",
				RawValue:        noteXRef,
				SuggestedAction: "Verify the note xref or add the missing note record.",
			})
			continue
		}
		parts = append(parts, note.FullText())
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func eventTitle(event *gedcom.Event, fallback string) string {
	if event == nil {
		return ""
	}
	if event.Description != "" {
		return event.Description
	}
	if event.EventTypeDetail != "" {
		return event.EventTypeDetail
	}
	return fallback
}

func sourceDate(source *gedcom.Source) string {
	if source == nil {
		return ""
	}
	if source.CreationDate != nil && source.CreationDate.Date != "" {
		return source.CreationDate.Date
	}
	if source.ChangeDate != nil {
		return source.ChangeDate.Date
	}
	return ""
}

func citationQuote(citation *gedcom.SourceCitation) string {
	if citation == nil || citation.Data == nil {
		return ""
	}
	return citation.Data.Text
}

func citationIdentityKey(identity SourceCitationIdentity) string {
	parts := []string{identity.SourceXRef, identity.Detail, identity.Quote, identity.Path}
	return strings.Join(parts, "|")
}

func citationKeyFromIdentity(identity SourceCitationIdentity) string {
	return "c_" + stableHash(citationIdentityKey(identity))
}

func personKey(xref string) string {
	if xref == "" {
		return ""
	}
	return "p_" + trimXRef(xref)
}

func groupKey(xref string) string {
	if xref == "" {
		return ""
	}
	return "g_" + trimXRef(xref)
}

func sourceKey(source *gedcom.Source) string {
	if source == nil {
		return ""
	}
	if source.XRef != "" {
		return "s_" + trimXRef(source.XRef)
	}
	seed := strings.Join([]string{source.Title, source.Author, source.Publication, source.Text}, "|")
	return "s_" + stableHash(seed)
}

func eventKey(ownerXRef, tag string, index int) string {
	if ownerXRef == "" {
		ownerXRef = "UNKNOWN"
	}
	if tag == "" {
		tag = "EVEN"
	}
	return fmt.Sprintf("e_%s_%s_%d", trimXRef(ownerXRef), tag, index)
}

func placeKey(rawPlace string) string {
	if rawPlace == "" {
		return ""
	}
	return "w_" + stableHash(rawPlace)
}

func stableHash(value string) string {
	hash := sha1.Sum([]byte(value))
	return hex.EncodeToString(hash[:])
}

func trimXRef(xref string) string {
	return strings.Trim(xref, "@")
}

func nextIndex(counter map[string]int, tag string) int {
	counter[tag]++
	return counter[tag]
}

func normalizeDate(date *gedcom.Date) (string, bool) {
	if date == nil || date.IsPhrase || date.Year == 0 {
		return "", false
	}
	start := formatDatePart(date)
	if start == "" {
		return "", false
	}
	modifier := strings.TrimSpace(date.Modifier.String())
	if date.EndDate != nil {
		end := formatDatePart(date.EndDate)
		if end == "" {
			return "", false
		}
		if modifier != "" {
			return fmt.Sprintf("%s %s/%s", modifier, start, end), true
		}
		return fmt.Sprintf("%s/%s", start, end), true
	}
	if modifier != "" {
		return fmt.Sprintf("%s %s", modifier, start), true
	}
	return start, true
}

func formatDatePart(date *gedcom.Date) string {
	if date == nil || date.Year == 0 {
		return ""
	}
	year := date.Year
	if date.IsBC {
		return fmt.Sprintf("-%04d", year)
	}
	if date.Month == 0 {
		return fmt.Sprintf("%04d", year)
	}
	if date.Day == 0 {
		return fmt.Sprintf("%04d-%02d", year, date.Month)
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, date.Month, date.Day)
}

func invalidDateIssue(entityType, entityKey, gedcomXRef, gedcomPath, rawDate string) Issue {
	return Issue{
		Severity:        "WARN",
		EntityType:      entityType,
		EntityKey:       entityKey,
		GedcomXRef:      gedcomXRef,
		GedcomPath:      gedcomPath,
		IssueCode:       "INVALID_DATE",
		Message:         "Date could not be normalized",
		RawValue:        rawDate,
		SuggestedAction: "Review the date value for GEDCOM compliance.",
	}
}

func gedcomPathForPerson(xref, tag string) string {
	return fmt.Sprintf("INDI(%s).%s", xref, tag)
}

func gedcomPathForPersonEvent(xref, tag string, index int) string {
	return fmt.Sprintf("INDI(%s).%s[%d]", xref, tag, index)
}

func gedcomPathForFamilyEvent(xref, tag string, index int) string {
	return fmt.Sprintf("FAM(%s).%s[%d]", xref, tag, index)
}

func gedcomPathForGroup(xref, tag string) string {
	return fmt.Sprintf("FAM(%s).%s", xref, tag)
}

func coordinatesLatitude(coords *gedcom.Coordinates) string {
	if coords == nil {
		return ""
	}
	return coords.Latitude
}

func coordinatesLongitude(coords *gedcom.Coordinates) string {
	if coords == nil {
		return ""
	}
	return coords.Longitude
}
