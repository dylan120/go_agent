package utils

import (
	log "github.com/sirupsen/logrus"
	"reflect"
)

func Generate(fc interface{}, abort <-chan struct{}, args ...interface{}) <-chan string {
	ch := make(chan string)

	go func(fc interface{}, args ...interface{}) {
		funcName := GetFunctionName(fc)
		funcValue := reflect.ValueOf(fc)
		if funcValue.Kind() != reflect.Func {
			log.Error(funcName + " is not a func type")
		} else {
			input := make([]reflect.Value, len(args))
			for i, arg := range args {
				input[i] = reflect.ValueOf(arg)
			}
			//values := funcValue.Call(input)
			//txt := values[0].Interface().(string)
		}

	}(fc, args...)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ch:
				<-ch
			case <-abort:
				break
			}
		}
	}()
	return ch
}

//func Generate(function interface{}) {
//
//}
