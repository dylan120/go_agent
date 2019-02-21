package utils

import (
	"../defaults"
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os/exec"
	"syscall"
	"time"
)

func IterTaskResult(jid string, scriptInterruptor string,
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
