package ot

import (
	"testing"
)

func TestDocumentFromLines(t *testing.T) {
	strings := []string{"These", "are", "strings"}
	doc := DocumentFromLines(strings)
	if len(doc.content) == 0 {
		t.Errorf("Failed to use DocumentFromLines")
	}
}
