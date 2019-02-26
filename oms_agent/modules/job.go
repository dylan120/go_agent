package main

import (
	"../defaults"
	"../utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CheckAlive(step utils.Step, procDir string, resultChannel chan string, status *defaults.Status) bool {
	var (
		info    utils.ProcessInfo
		isAlive = false
	)

	path := filepath.Join(procDir, step.InstanceID)
	content, err := ioutil.ReadFile(path)
	if !utils.CheckError(err) {
		err := json.Unmarshal(content, &info)
		if !utils.CheckError(err) {
			_, err := os.FindProcess(info.ProcessID)
			if err != nil {
				fmt.Printf("Failed to find process: %s\n", err)
			} else {
				isAlive = true
			}
		}
	}
	return isAlive
}
