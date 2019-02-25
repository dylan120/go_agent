package utils

import (
	"../defaults"
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type processInfo struct {
	jid       string   `json:"jid"`
	processID int      `json:"process_id"`
	cmd       []string `json:"cmd"`
}

func writeProcInfo(procDir string, procInfo processInfo) {
	path := filepath.Join(procDir, procInfo.jid)

	if _, err := os.Stat(procDir); os.IsNotExist(err) {
		os.Mkdir(procDir, 0555)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	defer f.Close()
	if !CheckError(err) {
		data, err := json.Marshal(procInfo)
		log.Debug(string(data))
		if !CheckError(err) {
			f.Write(data)
		}
	}

}

func IterJobResult(jid string, procDir string, scriptInterruptor string,
	script string, scriptParams string, timeout int, resultChannel chan string, status *defaults.Status) {
	nowTimestamp := time.Now().Unix()
	timeOutAt := nowTimestamp + int64(timeout)
	cmd := exec.Command(scriptInterruptor, script, scriptParams)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		//Pdeathsig: syscall.SIGTERM,
		Setsid: true,
	}

	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	err := cmd.Start()
	log.Info(cmd.Process.Pid)

	info := processInfo{
		jid:       jid,
		processID: cmd.Process.Pid,
		cmd:       []string{scriptInterruptor, script, scriptParams},
	}
	writeProcInfo(procDir, info)

	if !CheckError(err) {
		go func() {
			scanner := bufio.NewScanner(stdout)
			lines := 0
			results := ""
			for scanner.Scan() {
				if status.IsFinished {
					log.Warnf("job with id %s result chan closed,", jid)
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

			if !status.IsFinished {
				if results != "" {
					resultChannel <- results
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
			} else {
				cmd.Process.Signal(syscall.SIGTERM)
			}
			log.Debugf("job with id %s done!", jid)
		}()

		go func() {
			if time.Now().Unix() > timeOutAt {
				desc := fmt.Sprintf("job with id %s timeout with %d", jid, timeout)
				log.Warnf(desc)
				cmd.Process.Signal(syscall.SIGTERM)
				status.Set(defaults.TimeOut, desc, true)
			}
		}()
	}

}
