package validator

import (
	"fmt"
	"strings"

	"github.com/cacack/gedcom-go/decoder"
)

func ExampleValidator_Validate() {
	input := `0 HEAD
1 GEDC
2 VERS 5.5
0 @I1@ INDI
0 TRLR
`

	doc, err := decoder.Decode(strings.NewReader(input))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	v := New()
	errs := v.Validate(doc)
	fmt.Printf("issues: %d\n", len(errs))
	// Output: issues: 1
}
