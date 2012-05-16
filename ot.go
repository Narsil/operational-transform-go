package ot

import (
    "crypto/sha1"
    "io"
)

//Contains operations types
type OperationType int
const (
    INSERT = 1
    DELETE = 2
    SKIP = 3
)

type Document struct{
    text string
    checksums map[string]int
    ops []Operation
}

type Component struct{
    componentType OperationType
    length int
    str string
}

type Operation []Component

func NewDocument(text string) Document{
    h := sha1.New()
    io.WriteString(h, text)
    return Document{
        text:text,
        checksums:map[string]int{
            string(h.Sum(nil)):0,
        },
    }
}

func (op1 Operation)transform(op2 Operation) Operation{
    return op1
}

func (doc Document)insert(pos int, str string) (Document, error){
    return doc.do(INSERT, pos, str)
}

func (doc Document)delete(pos int, str string) (Document, error){
    return doc.do(DELETE, pos, str)
}

type InvalidComponentError struct {
    msg string
}

func (e InvalidComponentError) Error() string{
    return e.msg
}

func (doc Document) checksum() string{
    h := sha1.New()
    io.WriteString(h, doc.text)
    return string(h.Sum(nil))
}

func (doc Document) NewOperation(typeOp OperationType, pos int, str string) (Operation, string){
    op := Operation{
        Component{
            componentType:SKIP,
            length:pos,
        },
        Component{
            componentType:typeOp,
            str:str,
        },
    }
    return op, doc.checksum()
}

func (doc Document) do(typeOp OperationType, pos int, str string) (Document, error){
    op, check := doc.NewOperation(typeOp, pos, str)
    return doc.apply(op, check)
}

func (doc Document) apply(op Operation, checksum string) (Document, error){
    last_op_index := doc.checksums[checksum]
    if last_op_index != len(doc.ops){
        transform_ops := doc.ops[last_op_index:]
        for i:=0; i<len(transform_ops); i++{
            top := transform_ops[i]
            op = op.transform(top)
        }
    }
    position := 0
    str := doc.text
    new_str := ""
    for c:=0; c<len(op); c++{
        comp:=op[c]
        switch (comp.componentType){
        case INSERT:
            new_str += comp.str
        case DELETE:
            str_length := len(comp.str)
            deleted := str[position:position+str_length]
            if deleted != comp.str{
                return doc, InvalidComponentError{"Trying to delete inexistent str"}
            }
            position += str_length
        case SKIP:
            new_str += str[position:position+comp.length]
            position += comp.length
        }
    }
    new_str += str[position:]

    doc.text = new_str
    doc.ops = append(doc.ops, op)
    h := sha1.New()
    io.WriteString(h, doc.text)
    doc.checksums[string(h.Sum(nil))] = len(doc.ops)
    return doc, nil
}
