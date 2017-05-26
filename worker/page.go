package worker

import (
	"bytes"
	"unicode"
)

// Link is a wikipedia link. e.g. "Mike Tyson".
type Link string

// Compare compares a link against another one.
func (l Link) Compare(l2 Link) bool {
	b1 := l.buildBytes(l)
	b2 := l.buildBytes(l2)
	return bytes.Compare(b1, b2) == 0
}

func (l Link) buildBytes(link Link) []byte {
	result := []byte{}
	for _, b := range link {
		r, ok := l.normalizeRune(b)
		if !ok {
			continue
		}
		result = append(result, byte(r))
	}
	return result
}

func (l Link) normalizeRune(b int32) (rune, bool) {
	r := unicode.ToLower(rune(b))
	// TODO: this does not work with only number links, fix it
	if !unicode.IsLetter(r) {
		return 0, false
	}
	return r, true
}

// Page represents a crawled wikipedia page with Name and embedded Links.
type Page struct {
	Name  Link
	Links map[Link]uint64
}

// Has return true if item exists in a page.
func (p Page) Has(item Link) bool {
	if item.Compare(p.Name) {
		return true
	}

	for l := range p.Links {
		if item.Compare(l) {
			return true
		}
	}

	return false
}
