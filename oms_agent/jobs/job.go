package jobs

import (
	"../config"
	"../defaults"
	"../returners"
	"../transport"
	"../utils"
	"encoding/json"
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
)

func cmdJob(step *utils.Step, server transport.ServerChannel) {
	data, err := json.Marshal(step)
	if !utils.CheckError(err) {
		server.Publish(step.Minions, data)
	}
}

func fileJob(step *utils.Step, opts *config.MasterOptions, server transport.ServerChannel) {
	log.Info("run file job")
	fileSource := step.FileSource
	if len(fileSource) == 0 {
		log.Error(defaults.InValidFileArg, errors.New("file source is null"))
	}
	for _, fs := range fileSource {
		//step.FileSource[i].MD5sum
		srcFile := strings.TrimSpace(fs.File)
		if fs.Mode == "web_agent" {
			log.Debugf("init local file transfer.")
			// mtgt = get_masters(web_minions, step['minions'], oms_client.opts)
			utils.MakeTorrent(
				opts.BtAnnouce,
				srcFile,
				step.InstanceID,
			)

		} else {

		}
	}
}

func SqlJob(step *utils.Step, server transport.ServerChannel) {

}

func checkJobAlive(
	opts *config.MasterOptions, jid string,
	server transport.ServerChannel, minions []string) bool {
	var (
		isSuccess = false
		isBreak   = false
	)
	timeoutAt := time.Now().Unix() + int64(opts.TimeOut)
	context, _ := zmq.NewContext()
	defer context.Term()
	eventSubSock, _ := context.NewSocket(zmq.SUB)
	defer eventSubSock.Close()
	eventSubSock.Connect("ipc://" + filepath.Join(opts.SockDir, "event_publish.ipc"))
	eventSubSock.SetSubscribe("")

	for {
		//timeout := time.After(time.Duration(opts.TimeOut) * time.Second)
		if time.Now().Unix() > timeoutAt {
			log.Debugf("jid %s timeout %d", jid, opts.TimeOut)
			break
		}
		var (
			//step = utils.Step{
			//	Function:    "job.checkalive",
			//	IsFinished:  false,
			//	BlockName:   "CheckAlive",
			//	Creator:     "agent",
			//	Type:        1,
			//	ScriptParam: jid,
			//	Name:        "CheckAlive",
			//	IsPause:     false,
			//	TimeOut:     opts.TimeOut,
			//	Minions:     minions,
			//	InstanceID:  instanceID,
			//}
			runningMinion = make(map[string]int)
			doneMinion    = make(map[string]int)
		)

		step, err := utils.NewStep(
			"job.checkalive",
			"agent",
			jid,
			opts.TimeOut,
			minions,
			"")
		prefix := strings.Join([]string{utils.JobTagPrefix, step.InstanceID}, "/")

		if utils.CheckError(err) {
			break
		}

		data, err := json.Marshal(step)
		if utils.CheckError(err) {
			break
		}
		server.Publish(step.Minions, data)
		for {
			if time.Now().Unix() > timeoutAt {
				break
			}
			msg, err := eventSubSock.RecvBytes(0)
			if !utils.CheckError(err) {
				event := utils.Event{}
				load := utils.Load{}
				payLoad := utils.Payload{}

				err = utils.UnPackPayload(msg, &payLoad)
				if err == nil {
					err = json.Unmarshal(payLoad.Data, &load)
					if !utils.CheckError(err) {
						err = json.Unmarshal(load.Data, &event)
						if !utils.CheckError(err) {
							log.Debug(event.Tag)
							if strings.HasPrefix(event.Tag, prefix) {
								log.Debugf("receive event data: %s", load.Data)
								if event.Function == "job.checkalive" && event.Params == jid {
									if event.Retcode == defaults.Success {
										runningMinion[event.MinionId] = 1
										timeoutAt = time.Now().Unix() + int64(opts.TimeOut)
										log.Debugf("job %s is still running!", jid)

									} else if event.Retcode == defaults.Failure {
										doneMinion[event.MinionId] = 1
									}

									if len(runningMinion) == len(minions) {
										break
									}

									if len(doneMinion) == len(minions) {
										log.Debugf("job %s in all minion done!", jid)
										isBreak = true
										break
									}
								}
							}
						}
					}
				}
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}

		if isBreak {
			log.Debug("stop check minions alive")
			break
		}
		time.Sleep(2 * time.Second)
	}
	time.Sleep(1500 * time.Millisecond)
	isSuccess = returners.CheckJobStatus(opts, jid)
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
					Function: step.Function,
					JID:      jid,
					MinionId: mid,
					Retcode:  defaults.Wait,
					JobType:  step.Type,
				})
			}
			returners.UpdateMinion(opts, events, true)
			switch step.Type {
			case utils.CmdType:
				cmdJob(&step, server)
			case utils.FileType:
				fileJob(&step, opts, server)
			case utils.SqlType:
				SqlJob(&step, server)
			}
			time.Sleep(time.Second)
			isSuccess := checkJobAlive(opts, jid, server, step.Minions)
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
