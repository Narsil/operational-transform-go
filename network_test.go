package ot

import (
	"testing"
    "net"
    "encoding/json"
)

func TestConnections(t *testing.T){
    doc := newTestDocument()
    // TODO Needs to be close in a proper way. Simple defer doc.Close()
    // Does not work atm.
    doc.Listen(":7776")

    conn, err := net.Dial("tcp", "localhost:7776")
    if err != nil{
        t.Errorf("Could not connect to document")
    }

    op := Operation{Component{
		Path: []string{"doc", "4"},
		Si:   "ho",
    }}

    remote := RemoteOperation{
        Op: op,
        Checksum: []byte(doc.Checksum()),
    }

    str, err := json.Marshal(remote)

    conn.Write(str)
    conn.Write([]byte{'0'})

    buf := make([]byte, 3)
    conn.Read(buf)

    if string(buf) != "ACK"{
        t.Errorf("Did not receive ACK")
    }

    result, err := doc.content.get([]string{"doc"})
    target := "Hahaho this is is some text"
    if result != target || err != nil{
        t.Errorf("RemoteOperation was applied incorrectly")
    }
}

func TestLinkedDocuments(t *testing.T){
    doc1 := newTestDocument()
    doc2 := newTestDocument()

    //TODO Listen connection from previous test is not correctly closed.
    doc1.Listen(":7777")
    doc2.Listen(":7778")
    doc1.hosts = append(doc1.hosts, "localhost:7778")
    doc2.hosts = append(doc2.hosts, "localhost:7777")

    op := Operation{Component{
		Path: []string{"doc", "4"},
		Si:   "ho",
    }}

    err, finished := doc1.Apply(op)
    if err != nil{
        t.Errorf("Could not apply op")
    }

    f := <-finished


    if !f{
        t.Errorf("Some remote operations failed")
    }

    result, err := doc2.content.get([]string{"doc"})
    target := "Hahaho this is is some text"
    if result != target || err != nil{
        t.Errorf("Connections between documents is not working got %q instead of %q", result, target)
    }
}
