package ot

import (
    "strconv"
)

func DocumentFromLines(lines []string) (doc Document){
    doc = Document{
        checksums:make(map[string]int),
        ops:make([]Operation, 0),
    }
    for i, line := range lines{
        doc.content[strconv.Itoa(i)] = line
    }
    return

}
