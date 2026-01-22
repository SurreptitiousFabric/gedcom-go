package decoder

import (
	"fmt"
	"strings"
)

func ExampleDecode() {
	input := `0 HEAD
1 GEDC
2 VERS 5.5
0 @I1@ INDI
1 NAME John /Smith/
0 TRLR
`

	doc, err := Decode(strings.NewReader(input))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("records: %d\n", len(doc.Records))
	// Output: records: 1
}
