package utils

import (
	"github.com/satori/go.uuid"
	"path/filepath"
	"strconv"
)

func EventTag(prefix string, jid string, minionId string, seq int) string {
	return filepath.Join(prefix, jid, minionId, strconv.Itoa(seq))
}

func GenInstanceID() (string, error) {
	uuId, err := uuid.NewV4()
	return uuId.String(), err
}
