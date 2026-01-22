package encoder

import (
	"bytes"
	"fmt"

	"github.com/cacack/gedcom-go/gedcom"
)

func ExampleEncode() {
	doc := &gedcom.Document{
		Header: &gedcom.Header{
			Version:  "5.5",
			Encoding: "UTF-8",
		},
		Records: []*gedcom.Record{
			{
				XRef: "@I1@",
				Type: gedcom.RecordTypeIndividual,
				Tags: []*gedcom.Tag{
					{Level: 1, Tag: "NAME", Value: "Jane /Doe/"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, doc); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Print(buf.String())
	// Output:
	// 0 HEAD
	// 1 GEDC
	// 2 VERS 5.5
	// 1 CHAR UTF-8
	// 0 @I1@ INDI
	// 1 NAME Jane /Doe/
	// 0 TRLR
}
