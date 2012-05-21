package ot

import (
	"fmt"
	"strings"
	"unsafe"
)

type InvalidPathError struct {
	msg string
}

func (err InvalidPathError) Error() string {
	return err.msg
}

type Dict map[string]interface{}

func (content Dict) get(path []string) (inner string, err error) {
	dict := content
	var object interface{}
	for _, key := range path {
		object = dict[key]
		switch value := object.(type) {
		case Dict:
			dict = value
		case string:
			inner = value
		case unsafe.Pointer:
			inner = *(*string)(value)
		}

	}
	if inner == "" {
		err = InvalidPathError{"Path ['" + strings.Join(path, "', '") + "'] is not present in document"}
	}
	return
}

func (content Dict) set(path []string, str string) (err error) {
	dict := content
	var key string
	for i := 0; i < len(path)-1; i++ {
		key = path[i]
		switch value := dict[key].(type) {
		case Dict:
			dict = value
		default:
			err = InvalidPathError{"Path is not valid for current dict"}

		}
	}
	key = path[len(path)-1]
	pointer := unsafe.Pointer(&str)
	dict[key] = pointer
	return
}

func (content Dict) String() (str string) {
	for k, v := range content {
		str += k + ": "
		switch value := v.(type) {
		case string:
			str += value
		case fmt.Stringer:
			str += value.String()
		case unsafe.Pointer:
			str += *(*string)(value)
		}
		str += "\n"
	}
	return
}
