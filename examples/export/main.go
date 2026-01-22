// Example: Export GEDCOM data to JSON files.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/cacack/gedcom-go/decoder"
	"github.com/cacack/gedcom-go/gedcom"
)

type placeKey struct {
	name string
	form string
	lat  string
	long string
}

type placeOutput struct {
	Name        string `json:"name"`
	Form        string `json:"form,omitempty"`
	Latitude    string `json:"latitude,omitempty"`
	Longitude   string `json:"longitude,omitempty"`
	Occurrences int    `json:"occurrences"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <gedcom_file> [output_dir]")
		fmt.Println("Example: go run main.go ../../testdata/gedcom-5.5/minimal.ged ./out")
		os.Exit(1)
	}

	input := os.Args[1]
	outDir := "out"
	if len(os.Args) > 2 {
		outDir = os.Args[2]
	}

	f, err := os.Open(input) // #nosec G304 -- CLI tool accepts user-provided paths
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer f.Close()

	doc, err := decoder.Decode(f)
	if err != nil {
		log.Fatalf("decode: %v", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	individuals := doc.Individuals()
	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].XRef < individuals[j].XRef
	})
	families := doc.Families()
	sort.Slice(families, func(i, j int) bool {
		return families[i].XRef < families[j].XRef
	})
	sources := doc.Sources()
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].XRef < sources[j].XRef
	})

	places := collectPlaces(individuals, families)

	if err := writeJSON(filepath.Join(outDir, "individuals.json"), individuals); err != nil {
		log.Fatalf("write individuals: %v", err)
	}
	if err := writeJSON(filepath.Join(outDir, "families.json"), families); err != nil {
		log.Fatalf("write families: %v", err)
	}
	if err := writeJSON(filepath.Join(outDir, "sources.json"), sources); err != nil {
		log.Fatalf("write sources: %v", err)
	}
	if err := writeJSON(filepath.Join(outDir, "places.json"), places); err != nil {
		log.Fatalf("write places: %v", err)
	}

	fmt.Printf("Wrote %d individuals, %d families, %d sources, %d places to %s\n",
		len(individuals), len(families), len(sources), len(places), outDir)
}

func collectPlaces(individuals []*gedcom.Individual, families []*gedcom.Family) []placeOutput {
	places := make(map[placeKey]*placeOutput)

	addPlace := func(name, form string, coords *gedcom.Coordinates) {
		if name == "" {
			return
		}
		lat := ""
		long := ""
		if coords != nil {
			lat = coords.Latitude
			long = coords.Longitude
		}
		key := placeKey{name: name, form: form, lat: lat, long: long}
		if existing, ok := places[key]; ok {
			existing.Occurrences++
			return
		}
		places[key] = &placeOutput{
			Name:        name,
			Form:        form,
			Latitude:    lat,
			Longitude:   long,
			Occurrences: 1,
		}
	}

	addEventPlaces := func(events []*gedcom.Event) {
		for _, event := range events {
			if event == nil {
				continue
			}
			if event.PlaceDetail != nil {
				addPlace(event.PlaceDetail.Name, event.PlaceDetail.Form, event.PlaceDetail.Coordinates)
				continue
			}
			addPlace(event.Place, "", nil)
		}
	}

	for _, individual := range individuals {
		if individual == nil {
			continue
		}
		addEventPlaces(individual.Events)
		for _, attr := range individual.Attributes {
			if attr == nil {
				continue
			}
			addPlace(attr.Place, "", nil)
		}
	}

	for _, family := range families {
		if family == nil {
			continue
		}
		addEventPlaces(family.Events)
	}

	result := make([]placeOutput, 0, len(places))
	for _, place := range places {
		result = append(result, *place)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		if result[i].Form != result[j].Form {
			return result[i].Form < result[j].Form
		}
		if result[i].Latitude != result[j].Latitude {
			return result[i].Latitude < result[j].Latitude
		}
		return result[i].Longitude < result[j].Longitude
	})
	return result
}

func writeJSON(path string, value any) error {
	f, err := os.Create(path) // #nosec G304 -- CLI tool accepts user-provided paths
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
