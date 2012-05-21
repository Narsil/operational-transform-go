package ot

import (
	"strconv"
)

func DocumentFromLines(lines []string) (doc Document) {
	doc = Document{
		checksums: make(map[string]int),
		content:   Dict{},
		ops:       make([]Operation, 0),
	}
	for i, line := range lines {
		str := strconv.Itoa(i)
		doc.content[str] = line
	}
	return doc
}
