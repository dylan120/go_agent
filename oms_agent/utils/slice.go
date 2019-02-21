package utils

import (
	"log"
	"reflect"
)

func SliceExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)
	isExist := false
	if s.Kind() != reflect.Slice {
		log.Println("SliceExists() given a non-slice type")
	}
	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			isExist = true
		}
	}
	return isExist
}
