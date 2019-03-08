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

func NewStep(function string,
	creator string, scriptParam string, timeOut int, minions []string,
	instanceID string) (*Step, error) {
	var err error
	if instanceID == "" {
		instanceID, err = GenInstanceID()
		if CheckError(err) {
			return nil, err
		}
	}
	return &Step{
		Function:    function,
		IsFinished:  false,
		BlockName:   function,
		Creator:     creator,
		Type:        1,
		ScriptParam: scriptParam,
		Name:        function,
		IsPause:     false,
		TimeOut:     timeOut,
		Minions:     minions,
		InstanceID:  instanceID,
	}, nil
}
