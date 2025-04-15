package judaism

import (
	"fmt"
	"testing"
)

// Test the sefaria index
func TestSefariaIndex(t *testing.T) {
	index := GetFullIndex()
	fmt.Printf("index: %v\n", index)
	if len(index) == 0 {
		t.FailNow()
	}
}

func TestSefariaText(t *testing.T) {
	text := GetText("Genesis 1", "en", "")
	if text.Ref == "" {
		t.FailNow()
	}
}

func TestSefariaTextLinks(t *testing.T) {
	text := GetText("Genesis 1", "en", "")
	links := GetLinks(text.Ref, "en", "")
	if len(links) == 0 {
		t.FailNow()
	}
}
