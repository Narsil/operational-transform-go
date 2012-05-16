package ot

import (
    "testing"
)

func TestDocumentValidOps(t *testing.T){
    doc := NewDocument("Haha this is is some text")
    target := NewDocument("Hahaho this is some text")
    doc, err := doc.insert(4, "ho")
    doc, err = doc.delete(12, "is ")

    if err != nil{
        t.Errorf("Simple Operation failed (Error %q)", err)
    }
    if doc.text != target.text{
        t.Errorf("simple Operation failed %q != %q", doc.text, target.text)
    }
}

func TestDocumentConcurrent(t *testing.T){
    doc := NewDocument("Haha this is is some text")
    target := NewDocument("Hahaho this is some text")
    op1, check1 := doc.NewOperation(INSERT, 4, "ho")
    op2, check2 := doc.NewOperation(DELETE, 7, "is ")

    doc_left, err := doc.apply(op1, check1)
    doc_left, err = doc_left.apply(op2, check2)

    if err != nil{
        t.Errorf("Concurrent Operation failed (Error %q)", err)
    }
    if doc_left.text != target.text{
        t.Errorf("Concurrent Operation failed %q != %q", doc_left.text, target.text)
    }

}

func TestInvalidDeletes(t *testing.T){
    doc := NewDocument("Some doc")
    doc, err := doc.delete(0, "i")
    if _, ok := err.(InvalidComponentError); !ok{
        t.Error("Invalid DELETE did not raise error")
    }
}

func BenchmarkApplyInserts(b *testing.B){
    doc := NewDocument("Haha this is is some doc")
    for i:=0; i<b.N; i++{
        doc.insert(4, "ha")
    }
}

func BenchmarkApplyDeletes(b *testing.B){
    doc := NewDocument("Haha this is is some doc")
    for i:=0; i<b.N; i++{
        doc.delete(0, "H")
    }
}
