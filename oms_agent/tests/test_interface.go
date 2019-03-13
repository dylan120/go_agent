package main

import (
	"fmt"
	"reflect"
)

type a struct {
}

func test(i interface{}) {
	val := reflect.ValueOf(i)
	fmt.Println(val.Type().Name())
}

func main() {
	test(&a{})
}
