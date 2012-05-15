package main

import (
    "fmt"
)

const (
    INSERT = 1
    DELETE = 2
)

type Component struct{
    code int
    position int
    text string
}

type Operation []Component

func (op Operation) apply(str string) string{
    new_str := str
    for c:=0; c<len(op); c++{
        comp:=op[c]
        switch (comp.code){
        case INSERT:
            new_str = new_str[:comp.position] + comp.text + new_str[comp.position:]
        case DELETE:
            text_length := len(comp.text)
            current_text := new_str[comp.position:comp.position+text_length]
            if current_text != comp.text {
                fmt.Println("ERREUR")
            }
            new_str = new_str[:comp.position] + new_str[comp.position+text_length:]
        }
    }
    return new_str
}

func main(){
    str := "Haha this is is some text"
    op := Operation{Component{INSERT, 4, "ha"}, Component{DELETE, 12, "is "}}
    new_str := op.apply(str)
    fmt.Println(new_str)
}
