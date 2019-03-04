package utils

import "github.com/satori/go.uuid"

func GenInstanceID() (string, error) {
	uuId, err := uuid.NewV4()
	return uuId.String(), err
}
