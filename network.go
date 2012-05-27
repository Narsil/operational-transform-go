package ot

import (
    "net"
    "encoding/json"
    "bufio"
    "log"
)

//Contains an operation, and the version it was built against so that remote
//documents can properly transform before applying.
type RemoteOperation struct {
    Op Operation
    Checksum []byte
}

func (doc *Document)sendRemote(op Operation, checksum string, finished chan bool){
    remote := RemoteOperation{
        Op: op,
        Checksum: []byte(checksum),
    }
    str, err := json.Marshal(remote)
    if err != nil{
        log.Fatal("Could not marshal op")
    }

    sent_hosts := make(chan int, len(doc.hosts))
    for _, host := range doc.hosts{
        go func(){
            conn, err := net.Dial("tcp", host)
            if err != nil{
                log.Fatal("Dialing error", err)
            }
            conn.Write(str)
            conn.Write([]byte{'0'})
            buf := make([]byte, 3)
            conn.Read(buf)

            if string(buf) != "ACK"{
                log.Fatal("Did not receive ACK")
            }
            sent_hosts <- 1
        }()
    }
    go func(){
        i := 0
        for _, _ = range doc.hosts{
            i += <-sent_hosts
        }
        if i == len(doc.hosts){
            finished<-true
        }else{
            finished<-false
        }
    }()

    return
}

func (doc *Document)Listen(port string){
    ln , err :=  net.Listen("tcp", port)
    doc.listen = ln
    if err != nil{
        log.Fatal("Listening error", err)
    }
    go func(){
        for {
            conn, err := ln.Accept()
            if err != nil{
                log.Fatal("Cannot accept connection:", err)
            }
            go doc.handleConnection(conn)
        }
    }()
    return
}

func (doc *Document)Close(){
    doc.listen.Close()
}

func (doc *Document)Connect(host string){
    doc.hosts = append(doc.hosts, host)
    _, err := net.Dial("tcp", host)
    if err != nil{
        log.Fatal(err)
    }
    return
}

func (doc *Document)handleConnection(c net.Conn){
    for {
        jsonstring, err := bufio.NewReader(c).ReadBytes('0')
        if err != nil{
            log.Fatal("Reading remote error ", err)
        }
        remote := new(RemoteOperation)
        err = json.Unmarshal(jsonstring[:len(jsonstring)-1], &remote)
        if err != nil{
            //Eat lemons
            log.Fatal("Unnmarshling error", err)
        }
        err = doc.applyNoRemote(remote.Op, string(remote.Checksum))
        if err != nil{
            log.Fatal("Could not apply remote op")
        }
        c.Write([]byte("ACK"))
    }
    return
}

//Applies an operation to a document. This one will apply only local ops to
//the last version of the document. It will automatically send the op to 
//connected documents.
func (doc *Document) Apply(op Operation) (err error, finished chan bool) {
    checksum := doc.Checksum()
    err = doc.applyNoRemote(op, checksum)
    if err == nil{
        finished = make(chan bool, 1)
        doc.sendRemote(op, checksum, finished)
    }
    return
}

