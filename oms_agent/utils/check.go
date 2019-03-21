package utils

import (
	log "github.com/sirupsen/logrus"
)

func CheckType(t interface{}) {

}

func CheckError(err error) bool {
	isTrue := false
	if err != nil {
		log.Error(err.Error())
		isTrue = true
	}
	return isTrue
}

func RaiseError(err error) {
	if err != nil {
		panic(err)
	}
}
