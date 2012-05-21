package ot

import (
    "testing"
)

func TestDictHashing(t *testing.T){
    dict:= Dict{
        "doc": "Haha this is is some text",
    }
    h:= hash(dict)
    target := "\xf7\xf0)칊\xbe\x97\x90t\xa3\xabE\xb7M\xbd\x9a\xf0-B"
    if h != target{
        t.Errorf("Hash failed, got %x, instead of %x", h, target)
    }

    dict= Dict{
        "titi": "toto",
    }
    h= hash(dict)
    target = "\xf7眨\xeb\v1\xeeM]l\x18\x14\x16f\u007f\xfe\xe5(\xed"
    if h != target{
        t.Errorf("Hash failed, got %x, instead of %x", h, target)
    }
}

func TestDictGetting(t *testing.T){
    dict:= Dict{
        "doc": "Haha this is is some text",
    }
    str, err := dict.get([]string{"doc"})
    if err != nil{
        t.Errorf("Cannot correctly get from dict")
    }
    if str != "Haha this is is some text"{
        t.Errorf("Cannot correctly get from dict")
    }
}

func TestDictSetting(t *testing.T){
    dict:= Dict{
        "doc": "Haha this is is some text",
    }
    dict.set([]string{"doc"}, "New string")
    str, err := dict.get([]string{"doc"})
    if err != nil{
        t.Errorf("Cannot properly set to dict (%s)", err)
    }
    if str != "New string"{
        t.Errorf("Cannot properly set to dict")
    }
}
