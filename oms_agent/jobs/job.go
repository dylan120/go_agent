package jobs

import (
	"../config"
	"../returners"
	"../transport"
	"../utils"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

func cmdJob(step *utils.Step, server transport.ServerChannel) {
	data, err := json.Marshal(step)
	if !utils.CheckError(err) {
		server.Publish(step.Minions, data)
		log.Info("sent msg")
	}

}

func fileJob(step *utils.Step, server transport.ServerChannel) {

}

func SqlJob(step *utils.Step, server transport.ServerChannel) {

}

func checkJobStatus(opts *config.MasterOptions, jid string, minions []string) bool {
	var (
		isSuccess = false
		isBreak   = false
		children  []string
	)
	//timeout := time.Duration(opts.TimeOut) * time.Second
	//timeoutAt := time.Now().Unix() + int64(opts.TimeOut)
	timeout := time.After(time.Duration(opts.TimeOut) * time.Second)
	zkClient, jobPath, _ := transport.JobRegister(opts, jid)
	for {
		//if time.Now().Unix() > timeoutAt {
		//	log.Errorf("minion time out %ds", opts.TimeOut)
		//	break
		//}
		_, _, eventChan, err := zkClient.ChildrenW(jobPath)
		if !utils.CheckError(err) {
			isBreak = true
			select {
			case <-eventChan:
				children, _, err = zkClient.Children(jobPath)
				log.Debug(len(minions))
				if !utils.CheckError(err) {
					isBreak = true
					if len(children) == len(minions) {
						isBreak = true
						isSuccess = true
					}
				}
			case <-timeout:
				log.Errorf("minion time out %ds", opts.TimeOut)
				time.Sleep(100 * time.Millisecond)
				isBreak = true
			}
		}
		if isBreak {
			break
		}
	}
	utils.CheckError(transport.JobDone(opts, jid))
	return isSuccess
}

func run(opts *config.MasterOptions, task *utils.Task, server transport.ServerChannel) {
	var (
		status         = utils.Running
		job            = task.Data
		taskInstanceID = task.InstanceID
		jid            string
	)

	startTime := time.Now().Unix()
	returners.UpdateTask(
		opts, task,
		startTime, startTime, false, status, true)

	isContinue := true
	log.Debug(task)
	for i, block := range job.Blocks {
		if !isContinue {
			break
		}
		steps := block.Steps
		for j, step := range steps {
			var events []*utils.Event
			if !isContinue {
				break
			}
			log.Debug(step)
			step.InstanceID = fmt.Sprintf("%s_%d_%d", taskInstanceID, i+1, j+1)
			jid = step.InstanceID
			if step.IsFinished {
				log.Warnf("[%s] pass step %s", jid, step.Name)
				continue
			}
			log.Infof("[%s] run step %s", jid, step.Name)
			startTime := time.Now().Unix()
			returners.UpdateStep(opts, taskInstanceID,
				jid, &step,
				startTime, startTime, false,
				status, true)
			for _, mid := range step.Minions {
				events = append(events, &utils.Event{
					//Function: "event",
					JID:      jid,
					MinionId: mid,
					Retcode:  utils.Wait,
					JobType:  step.Type,
				})
			}
			returners.UpdateMinion(opts, events, true)
			//returners.UpdateMinionStatus(opts, jid, step.Minions, utils.Wait, true)
			switch step.Type {
			case utils.CmdType:
				cmdJob(&step, server)
			case utils.FileType:
				fileJob(&step, server)
			case utils.SqlType:
				SqlJob(&step, server)
			}
			isSuccess := checkJobStatus(opts, jid, step.Minions)
			if isSuccess {
				if step.IsPause {
					status = utils.Stop
					isContinue = false
					log.Infof("[%s] step %s run successfully and stopped.", jid, step.Name)
				} else {
					status = utils.Success
					log.Infof("[%s] step %s successfully.", jid, step.Name)
				}
				step.IsFinished = true

			} else {
				status = utils.Failure
				isContinue = false
				log.Infof("[%s] step %s failed.", jid, step.Name)
			}

			returners.UpdateStep(opts, taskInstanceID,
				jid, &step,
				startTime, startTime, false,
				status, true)
		}
	}
	endTime := time.Now().Unix()

	returners.UpdateTask(
		opts, task,
		startTime, endTime, true, status, false)

}

func Start(opts *config.MasterOptions, server transport.ServerChannel) {
	broker := Broker{}
	err := json.Unmarshal([]byte(opts.JobBroker), &broker)
	utils.RaiseError(err)
	broker.Init()
	for {
		task, err := broker.Get()
		if err == nil {
			log.Debugf("receive job with id %s", task.InstanceID)
			go run(opts, &task, server)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
