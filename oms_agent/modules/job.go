package main

import (
	"../defaults"
	"../utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func CheckAlive(step utils.Step, procDir string, resultChannel chan string, status *defaults.Status) {
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
				status.Set(defaults.Failure, fmt.Sprintf("jid %s does not exist", step.InstanceID), true)
			} else {
				isAlive = true
				status.Set(defaults.Success, fmt.Sprintf("jid %s alive", step.InstanceID), true)
			}
		}
	}
	resultChannel <- strconv.FormatBool(isAlive)

}
