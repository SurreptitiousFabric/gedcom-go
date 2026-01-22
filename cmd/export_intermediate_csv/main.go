package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cacack/gedcom-go/decoder"
	"github.com/cacack/gedcom-go/export/intermediatecsv"
)

func main() {
	inputPath := flag.String("input", "", "Path to GEDCOM file to export")
	outputDir := flag.String("output", "out", "Output directory for CSV files")
	includeSources := flag.Bool("include-sources", true, "Include sources/citations CSV files")
	includePlaces := flag.Bool("include-places", true, "Include places CSV file")
	includeGroups := flag.Bool("include-groups", true, "Include groups CSV files")
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Usage: export_intermediate_csv -input <gedcom_file> [-output <dir>] [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	input, err := os.Open(*inputPath) // #nosec G304 -- CLI tool accepts user-provided paths
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer input.Close()

	doc, err := decoder.Decode(input)
	if err != nil {
		log.Fatalf("decode: %v", err)
	}

	model := intermediatecsv.BuildIntermediateModel(doc)
	issues := intermediatecsv.ValidateIntermediate(model)

	options := intermediatecsv.Options{
		IncludeSources: *includeSources,
		IncludePlaces:  *includePlaces,
		IncludeGroups:  *includeGroups,
	}

	if err := intermediatecsv.WriteCSVBundle(model, issues, *outputDir, options); err != nil {
		log.Fatalf("export: %v", err)
	}

	fmt.Printf("Wrote CSV bundle to %s (persons=%d events=%d issues=%d)\n",
		*outputDir, len(model.Persons), len(model.Events), len(issues))
}
