package utils

import (
	"errors"
)

var (
	RowNoFound = errors.New("row no found")
	AuthFailure = errors.New("auth failed")
	MissMatchPubPkey = errors.New("public key miss matched")
	ConnectFailed = errors.New("connect failed")
)
