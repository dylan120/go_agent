package main

import (
	"fmt"
	"reflect"
)

type a struct {
}

func test(i interface{}) {
	val := reflect.ValueOf(i)
	fmt.Println(val.Type())
}

func main() {
	t := a{}
	test(&t)
}
