package ot

import (
	"crypto/sha1"
	"fmt"
	"io"
	"strconv"
	"unsafe"
)

//Contains operations types
type OperationType int

const (
	INSERT = 1
	DELETE = 2
	SKIP   = 3
)

type Document struct {
	content   Dict
	checksums map[string]int
	ops       []Operation
}

type Component struct {
	path []string
	si   string
	sd   string
}

type Operation []Component

func pathEquals(strslice1, strslice2 []string) (b bool) {
	b = false
	if len(strslice1) != len(strslice2) {
		return
	}
	for i := 0; i < len(strslice1); i++ {
		el1 := strslice1[i]
		el2 := strslice2[i]
		if el1 != el2 {
			return
		}
	}
	b = true
	return
}

func hash(content Dict) string {
	h := sha1.New()
	for key, val := range content {
		io.WriteString(h, key)
		switch value := val.(type) {
		case Dict:
			io.WriteString(h, hash(value))
		case []interface{}:
			for _, el := range value {
				switch element := el.(type) {
				case Dict:
					io.WriteString(h, hash(element))
				case fmt.Stringer:
					io.WriteString(h, element.String())
				}

			}
		case fmt.Stringer:
			io.WriteString(h, value.String())
		case unsafe.Pointer:
			io.WriteString(h, *(*string)(value))
		}
	}
	return string(h.Sum(nil))
}

func NewDocument(content Dict) Document {
	h := hash(content)
	return Document{
		content: content,
		checksums: map[string]int{
			h: 0,
		},
	}
}

func transformPosition(oldpos int, comp Component) (newpos int) {
	newpos = oldpos
	compos := comp.position()
	if comp.si != "" {
		if compos <= oldpos {
			newpos += len(comp.si)
		}
	} else {
		if oldpos <= compos {
			newpos = oldpos
		} else if oldpos <= compos+len(comp.sd) {
			newpos = compos
		} else {
			newpos = oldpos - len(comp.sd)
		}
	}
	return

}

func (comp1 Component) transform(dest *Operation, comp2 Component) {
	pos1 := comp1.position()
	if comp1.si != "" { //Insert
		comp1.setPosition(transformPosition(pos1, comp2))
	} else { //Delete
		if comp2.si != "" { // Delete vs Insert
			deleted := comp1.sd
			if pos1 < comp2.position() {
				(*dest).append(Component{
					path: comp1.path,
					sd:   deleted[:comp2.position()-pos1]})
				deleted = deleted[comp2.position()-pos1:]
			}
			if deleted != "" {
				(*dest).append(Component{
					path: append(comp1.path[:len(comp1.path)-1], strconv.Itoa(pos1+len(comp2.si))),
					sd:   deleted,
				})

			}
		}
	}
	return
}

func (op Operation) append(comp Component) {
	op = append(op, comp)
}

func (op1 Operation) transform(op2 Operation) Operation {
	for _, comp2 := range op2 {
		for _, comp1 := range op1 {
			comp2path := comp2.path[:len(comp2.path)-1]
			comp1path := comp1.path[:len(comp1.path)-1]
			if pathEquals(comp1path, comp2path) {
				comp1.transform(&op1, comp2)
			}
		}
	}
	return op1
}

type InvalidComponentError struct {
	msg string
}

func (e InvalidComponentError) Error() string {
	return e.msg
}

func (doc Document) checksum() string {
	return hash(doc.content)
}

func (doc *Document) apply(op Operation, checksum string) (err error) {
	last_op_index := doc.checksums[checksum]
	if last_op_index != len(doc.ops) {
		transform_ops := doc.ops[last_op_index:]
		for i := 0; i < len(transform_ops); i++ {
			top := transform_ops[i]
			op = op.transform(top)
		}
	}
	content := doc.content
	for c := 0; c < len(op); c++ {
		comp := op[c]
		if comp.si != "" {
			index := comp.position()
			str, err := content.get(comp.path[:len(comp.path)-1])
			if err != nil {
				return InvalidComponentError{msg: str}
			}
			str = str[:index] + comp.si + str[index:]
			content.set(comp.path[:len(comp.path)-1], str)
		}
		if comp.sd != "" {
			str, err := content.get(comp.path[:len(comp.path)-1])
			if err != nil {
				return err
			}
			str_length := len(comp.sd)
			index := comp.position()
			deleted := str[index : index+str_length]
			if deleted != comp.sd {
				return InvalidComponentError{"Trying to delete '" + comp.sd + "' but found '" + deleted + "' instead"}
			}
			new_str := str[:index] + str[index+str_length:]
			content.set(comp.path[:len(comp.path)-1], new_str)
		}
	}
	doc.ops = append(doc.ops, op)
	doc.checksums[doc.checksum()] = len(doc.ops)
	return nil
}

func (comp Component) position() (pos int) {
	pos, _ = strconv.Atoi(comp.path[len(comp.path)-1])
	return
}
func (comp Component) setPosition(newpos int) {
	comp.path[len(comp.path)-1] = strconv.Itoa(newpos)
	return
}
