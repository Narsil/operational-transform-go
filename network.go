package ot

import (
    "net"
    "encoding/json"
    "fmt"
    "bufio"
)

type RemoteOperation struct {
    Op Operation
    Checksum []byte
}

func (doc Document)Listen(port string){
    ln , _ :=  net.Listen("tcp", port)
    go func(){
        for {
            conn, err := ln.Accept()
            if err != nil{
                continue
            }
            go doc.handleConnection(conn)
        }
    }()
    return
}

func (doc Document)handleConnection(c net.Conn){
    jsonstring, err := bufio.NewReader(c).ReadBytes('0')
    if err != nil{
        fmt.Println("Reading remote error", err)
    }
    remote := new(RemoteOperation)
    err = json.Unmarshal(jsonstring[:len(jsonstring)-1], &remote)
    if err != nil{
        //Eat lemons
        fmt.Println("Unnmarshling error", err)
    }
    err = doc.Apply(remote.Op, string(remote.Checksum))
    if err != nil{
        fmt.Println("Could not apply remote op")
    }
    c.Write([]byte("ACK"))
    return
}

