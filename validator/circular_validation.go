package validator

import (
	"fmt"

	"github.com/cacack/gedcom-go/gedcom"
)

func (v *Validator) validateCircularRelationships(doc *gedcom.Document) {
	for _, ind := range doc.Individuals() {
		if ind == nil || ind.XRef == "" {
			continue
		}
		if hasCircularAncestry(doc, ind, ind.XRef, make(map[string]bool), make(map[string]bool)) {
			v.errors = append(v.errors, &ValidationError{
				Code:    "CIRCULAR_REFERENCE",
				Message: fmt.Sprintf("Circular family relationship detected for %s", ind.XRef),
				XRef:    ind.XRef,
			})
		}
	}
}

func hasCircularAncestry(doc *gedcom.Document, current *gedcom.Individual, target string, visiting, visited map[string]bool) bool {
	if current == nil || current.XRef == "" {
		return false
	}
	if visiting[current.XRef] {
		return current.XRef == target
	}
	if visited[current.XRef] {
		return false
	}

	visiting[current.XRef] = true
	for _, parent := range current.Parents(doc) {
		if parent == nil {
			continue
		}
		if parent.XRef == target {
			return true
		}
		if hasCircularAncestry(doc, parent, target, visiting, visited) {
			return true
		}
	}
	visiting[current.XRef] = false
	visited[current.XRef] = true

	return false
}
