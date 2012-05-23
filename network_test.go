package ot

import (
	"testing"
    "net"
    "encoding/json"
)

func TestConnections(t *testing.T){
    doc := newTestDocument()
    doc.Listen(":7777")

    conn, err := net.Dial("tcp", "localhost:7777")
    if err != nil{
        t.Errorf("Could not connect to document")
    }
    comp := Component{
		Path: []string{"doc", "4"},
		Si:   "ho",
    }
	op1 := Operation{comp}
    remote := RemoteOperation{
        Op: op1,
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
