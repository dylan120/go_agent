package main

import (
	"../defaults"
	"../utils"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

func CheckAlive(step utils.Step, procDir string, resultChannel chan string, status *defaults.Status) {
	var (
		info utils.ProcessInfo
		//isAlive = false
		retcode int
		text    string
	)

	path := filepath.Join(procDir, step.ScriptParam)
	log.Debugf("check jid %s alive", step.ScriptParam)
	content, err := ioutil.ReadFile(path)
	if !utils.CheckError(err) {
		err = json.Unmarshal(content, &info)
		if !utils.CheckError(err) {
			proc, err := os.FindProcess(info.ProcessID)
			if !utils.CheckError(err) {
				//text = fmt.Sprintf("jid %s alive", step.ScriptParam)
				//retcode = defaults.Success
				err := proc.Signal(syscall.Signal(0))
				if !utils.CheckError(err) {
					text = fmt.Sprintf("jid %s alive", step.ScriptParam)
					retcode = defaults.Success
				} else {
					text = err.Error()
					retcode = defaults.Failure
				}

			} else {
				text = fmt.Sprintf("jid %s does not exist", step.ScriptParam)
				retcode = defaults.Failure
			}
		}
	} else {
		text = fmt.Sprintf("jid %s does not exist", step.ScriptParam)
		retcode = defaults.Failure
	}
	//msg := strconv.FormatBool(isAlive)
	//log.Debug(msg)
	//resultChannel <- msg
	//time.Sleep(100 * time.Millisecond)
	status.Set(retcode, text, true)
}
