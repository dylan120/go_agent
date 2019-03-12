package utils

import (
	"bufio"
	"crypto/md5"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
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

const bufferSize = 65536

func MD5sum(filename string) (string, error) {
	if info, err := os.Stat(filename); err != nil {
		return "", err
	} else if info.IsDir() {
		return "", nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	for buf, reader := make([]byte, bufferSize), bufio.NewReader(file); ; {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		hash.Write(buf[:n])
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum, nil
}
