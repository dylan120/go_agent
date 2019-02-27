package main

import (
	"../defaults"
	"../utils"
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
			nowTimestamp := time.Now().Unix()
			timeOutAt := nowTimestamp + int64(timeOut)
			cmd := exec.Command(scriptInterruptor, scriptPath, scriptParam)
			cmd.SysProcAttr = &syscall.SysProcAttr{
				//Pdeathsig: syscall.SIGTERM,
				Setsid: true,
			}

			stdout, _ := cmd.StdoutPipe()
			cmd.Stderr = cmd.Stdout
			err := cmd.Start()
			if !utils.CheckError(err) {
				info := utils.ProcessInfo{
					JID:       jid,
					ProcessID: cmd.Process.Pid,
					Cmd:       []string{scriptInterruptor, scriptPath, scriptParam},
				}
				utils.WriteProcInfo(procDir, info)
				//utils.IterJobResult(jid, procDir, stdout, timeOut, resultChannel, status)

				scanner := bufio.NewScanner(stdout)
				lines := 0
				results := ""
				for scanner.Scan() {
					if time.Now().Unix() > timeOutAt {
						desc := fmt.Sprintf("job with id %s timeout with %d", jid, timeOut)
						log.Warnf(desc)
						cmd.Process.Signal(syscall.SIGTERM)
						close(resultChannel)
						status.Set(defaults.TimeOut, desc, true)
						break
					}
					lines += 1
					if results != "" {
						results = fmt.Sprintf("%s\n%s", results, scanner.Text())
					} else {
						results = scanner.Text()
					}

					if lines == 50 { //TODO maybe having another better way to do this
						status.Set(defaults.Run, "", false)
						resultChannel <- results
						time.Sleep(time.Duration(rand.Float64()) * time.Second)
						lines = 0
						results = ""
					}
				}

				if err := cmd.Wait(); err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok {
						if stat, ok := exitErr.Sys().(syscall.WaitStatus); ok {
							status.Set(stat.ExitStatus(), err.Error(), true)
						}
					} else {
						status.Set(defaults.Failure, fmt.Sprintf("cmd wait: %v", err), true)
					}
				} else {
					status.Set(defaults.Success, "", true)
				}

				log.Debugf("job with id %s done!", jid)

			}
		}

	} else {
		log.Errorf("user %s  is not available.", account)
	}

}
