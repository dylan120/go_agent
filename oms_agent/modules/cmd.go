package main

import (
	"../defaults"
	"../utils"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func getShell() string {
	return os.Getenv("SHELL")
}

func isValidShell(shell string) bool {
	var (
		shells  = "/etc/shells"
		isValid = false
	)
	if _, err := os.Stat(shells); !os.IsNotExist(err) {
		file, err := os.Open(shells)
		defer file.Close()
		if !utils.CheckError(err) {
			ch := utils.ReadFileByLine(file)
			for c := range ch {
				if c == shell {
					isValid = true
					break
				}
			}
		}
	}
	if !isValid {
		log.Warnf("try to get invalid shell %s", shell)
	}
	return isValid
}

func getScriptInterruptor(scriptContent string, interruptorType string) string {
	scriptInterruptor := ""

	if strings.HasPrefix(scriptContent, "#!/") {
		lines := strings.Split(scriptContent, "\n")
		if strings.HasPrefix(lines[0], "#!/") {
			scriptInterruptor = strings.Replace(lines[0], "#!/", "", 1)
		}
	} else {
		if interruptorType == "shell" {
			scriptInterruptor = getShell()
		} else if interruptorType == "python" {
			scriptInterruptor = "/usr/bin/python"
		}

		if isValidShell(scriptInterruptor) {
			if _, err := os.Stat(scriptInterruptor); !os.IsNotExist(err) {
				if utils.CheckFilePermission(scriptInterruptor) {
				} else {
					log.Errorf("shell %s is not available", scriptInterruptor)
				}
			}
		}
	}

	return scriptInterruptor
}

func Run(step utils.Step, procDir string, resultChannel chan string, status *defaults.Status) {
	d, _ := json.Marshal(&step)
	log.Debug(string(d))
	var (
		//ch                chan *utils.Event
		account, scriptSuffix,
		scriptInterruptor, interruptorType string
		err error = nil
	)

	//event.Result = step.ScriptParam
	if step.Account != "" {
		account = step.Account
	} else {
		account = "root"
	}
	if _, err = user.Lookup(account); err == nil {
		scriptType := step.ScriptType
		projectID := step.ProjectID
		jid := step.InstanceID
		timeOut := step.TimeOut
		scriptParam := step.ScriptParam
		scriptContent := step.ScriptContent

		if scriptType == "2" { //shell script
			scriptSuffix = ".sh"
			interruptorType = "shell"
		} else if scriptType == "1" { // 1 python script
			scriptSuffix = ".py"
			interruptorType = "python"
		}
		scriptName := fmt.Sprintf("%s_%d%s", jid, projectID, scriptSuffix)
		scriptPath := filepath.Join("/tmp", scriptName)
		scriptInterruptor = getScriptInterruptor(scriptContent, interruptorType)

		if _, err = os.Stat(scriptPath); !os.IsNotExist(err) {
			os.Remove(scriptPath)
		}
		err = ioutil.WriteFile(scriptPath, []byte(scriptContent), 0500)
		if !utils.CheckError(err) {
			//run the script
			utils.IterJobResult(jid, procDir, scriptInterruptor, scriptPath, scriptParam, timeOut, resultChannel, status)
		}

	} else {
		log.Errorf("user %s  is not available.", account)
	}

}
