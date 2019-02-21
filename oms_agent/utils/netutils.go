package utils

import (
	"net"
	"time"
)

func CheckAlive(dip string, dport string) bool {
	var isAlive bool
	isAlive = false
	dest := net.JoinHostPort(dip, dport)
	conn, err := net.DialTimeout("tcp", dest, 5*time.Second)

	if !CheckError(err) {
		defer conn.Close()
		isAlive = true
	}
	return isAlive
}
