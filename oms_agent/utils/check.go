package utils

import (
	"../log"
)

func CheckType(t interface{}){

}

func CheckError(err error) bool{
	isTrue := false
	if err != nil {
		log.Error.Println(err.Error())
		isTrue = true
	}
	return isTrue
}

func RaiseError(err error) {
	if err != nil {
		panic(err)
	}
}