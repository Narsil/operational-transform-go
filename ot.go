package ot

import (
	"crypto/sha1"
	"fmt"
	"io"
	"strconv"
	"unsafe"
    "net"
)

//This is document type. Dict is a very flexible struct to contain python like
//dict structure. For now only supports string inserts and deletes. Checksums
//is a mapping between the checksum of a document and the index within the ops
//array. It is used when receiving remote ops that where built against old versions
//of the doc to transform received operation against the operation that occured
//locally in the meantime.
type Document struct {
	content   Dict
	checksums map[string]int
	ops       []Operation
    hosts []string
    listen net.Listener
}

//An operation is a list of components. To build a complex operation use
// op.Append(component).
type Operation []Component

//Components contains a Path ["doc", "toto", "0"], that is the list of keys
//to descend to underlying strings and apply inserts and deletes contained 
//within either si or Sd.
type Component struct {
	Path []string
	Si   string
	Sd   string
}

//Compares two strings to see if they are the same Path.
func PathEquals(strslice1, strslice2 []string) (b bool) {
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

//hashes a Dict, to produce checksums used within Document struct. hashes reflects
//the whole dict, both values and keys to be unique for each document.
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

//Returns a new document containing initialized map, from dict.
func NewDocument(content Dict) (doc Document) {
	h := hash(content)
    doc = Document{
		content: content,
		checksums: map[string]int{
			h: 0,
		},
	}
    return
}

//Given the old position of an insert operation returns its new position
//when transforming against another component.
func transformPosition(oldpos int, comp Component) (newpos int) {
	newpos = oldpos
	compos := comp.position()
	if comp.Si != "" {
		if compos <= oldpos {
			newpos += len(comp.Si)
		}
	} else {
		if oldpos <= compos {
			newpos = oldpos
		} else if oldpos <= compos+len(comp.Sd) {
			newpos = compos
		} else {
			newpos = oldpos - len(comp.Sd)
		}
	}
	return

}

//Transforms a component against another one. We use dest to accumulate 
//components because the transform of a component may result in several
//components.
func (comp1 Component) transform(dest *Operation, comp2 Component) {
	pos1 := comp1.position()
	if comp1.Si != "" { //Insert
		comp1.setPosition(transformPosition(pos1, comp2))
	} else { //Delete
		if comp2.Si != "" { // Delete vs Insert
			deleted := comp1.Sd
			if pos1 < comp2.position() {
				(*dest).Append(Component{
					Path: comp1.Path,
					Sd:   deleted[:comp2.position()-pos1]})
				deleted = deleted[comp2.position()-pos1:]
			}
			if deleted != "" {
				(*dest).Append(Component{
					Path: append(comp1.Path[:len(comp1.Path)-1], strconv.Itoa(pos1+len(comp2.Si))),
					Sd:   deleted,
				})

			}
		}
	}
	return
}

//Appends a new component to an operation. If many components already exists
//within op, it will try to compress them in as few components as possible.
func (op Operation) Append(comp Component) {
	op = append(op, comp)
}

//transforms an operation against another one. This basically transform every
//component against every other component
func (op1 Operation) transform(op2 Operation) Operation {
	for _, comp2 := range op2 {
		for _, comp1 := range op1 {
			comp2Path := comp2.Path[:len(comp2.Path)-1]
			comp1Path := comp1.Path[:len(comp1.Path)-1]
			if PathEquals(comp1Path, comp2Path) {
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

//Returns the hash of the document. Used to determine version of doc (but 
//timeline agnostic as it only depends on the content.
func (doc Document) Checksum() string {
	return hash(doc.content)
}

//Applies an operation.checksum argument represents what
//checksum the document was built against. It is useful when receiving
//remote ops to know how to tranform received op against local ops.
//Apply func is in network.go
func (doc *Document) applyNoRemote(op Operation, checksum string) (err error) {
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
		if comp.Si != "" {
			index := comp.position()
			str, err := content.get(comp.Path[:len(comp.Path)-1])
			if err != nil {
				return InvalidComponentError{msg: str}
			}
			str = str[:index] + comp.Si + str[index:]
			content.set(comp.Path[:len(comp.Path)-1], str)
		}
		if comp.Sd != "" {
			str, err := content.get(comp.Path[:len(comp.Path)-1])
			if err != nil {
				return err
			}
			str_length := len(comp.Sd)
			index := comp.position()
			deleted := str[index : index+str_length]
			if deleted != comp.Sd {
				return InvalidComponentError{"Trying to delete '" + comp.Sd + "' but found '" + deleted + "' instead"}
			}
			new_str := str[:index] + str[index+str_length:]
			content.set(comp.Path[:len(comp.Path)-1], new_str)
		}
	}
	doc.ops = append(doc.ops, op)
	doc.checksums[doc.Checksum()] = len(doc.ops)
	return nil
}

//Returns the position at which a component is operating.
func (comp Component) position() (pos int) {
	pos, _ = strconv.Atoi(comp.Path[len(comp.Path)-1])
	return
}

//Sets position of component.
func (comp Component) setPosition(newpos int) {
	comp.Path[len(comp.Path)-1] = strconv.Itoa(newpos)
	return
}

//HelperFunction to create a new component. To be used in all cases because
//struct members are all private.
func NewInsertComponent(Path []string, str string) (comp Component){
    comp.Path = Path
    comp.Si = str
    return
}

//HelperFunction to create a new component. To be used in all cases because
//struct members are all private.
func NewDeleteComponent(Path []string, str string) (comp Component){
    comp.Path = Path
    comp.Sd = str
    return
}
