package utils

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"os"
)

func CheckFilePermission(path string) bool {
	var val os.FileMode = 0550
	isAccess := false
	//log.Info("size of a: ", unsafe.Sizeof(val))
	info, err := os.Stat(path)
	if !CheckError(err) {
		m := info.Mode()
		if m&val == val {
			isAccess = true
		}
	}
	return isAccess
}

func ReadFileByLine(file *os.File) <-chan string {
	ch := make(chan string)
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Error(err)
		}
		close(ch)
	}()
	return ch
}
