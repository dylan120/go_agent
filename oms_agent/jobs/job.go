package jobs

import (
	"../config"
	"../returners"
	"../transport"
	"../utils"
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
)

func subscribeEvent(opts *config.MasterOptions, prefix string, eventChan chan utils.Event, timeoutAt int64) {
	var (
		//eventChan = make(chan utils.Event)
		err   error
		event utils.Event
	)
	context, _ := zmq.NewContext()
	defer context.Term()
	eventSubSock, _ := context.NewSocket(zmq.SUB)
	defer eventSubSock.Close()
	eventSubSock.Connect("ipc://" + filepath.Join(opts.SockDir, "event_publish.ipc"))
	eventSubSock.SetSubscribe("")
	for {
		if time.Now().Unix() > timeoutAt {
			close(eventChan)
			break
		}
		msg, _ := eventSubSock.RecvBytes(0)
		load := utils.Load{}
		payLoad := utils.Payload{}

		err = utils.UnPackPayload(msg, &payLoad)
		if err == nil {
			err = json.Unmarshal(payLoad.Data, &load)
			if !utils.CheckError(err) {
				if strings.HasPrefix(event.Tag, prefix) {
					err = json.Unmarshal(load.Data, &event)
					if !utils.CheckError(err) {
						log.Debugf("receive event data: %s", event)
						if strings.HasPrefix(event.Tag, "/job") {
							eventChan <- event
						}
					}
				}
			}
		}
	}
}

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

func checkJobStatus(
	opts *config.MasterOptions, jid string,
	server transport.ServerChannel, minions []string) bool {
	var (
		isSuccess = false
		isBreak   = false
		//children  []string
	)
	//timeout := time.Duration(opts.TimeOut) * time.Second
	timeoutAt := time.Now().Unix() + int64(opts.TimeOut)
	//zkClient, jobPath, _ := transport.JobRegister(opts, jid)

	for {
		//timeout := time.After(time.Duration(opts.TimeOut) * time.Second)
		if time.Now().Unix() > timeoutAt {
			break
		}
		uuId, err := uuid.NewV4()
		if utils.CheckError(err) {
			break
		}
		step := utils.Step{
			Function:    "job.checkalive",
			IsFinished:  false,
			BlockName:   "CheckAlive",
			Creator:     "agent",
			Type:        1,
			ScriptParam: jid,
			Name:        "CheckAlive",
			IsPause:     false,
			TimeOut:     opts.TimeOut,
			Minions:     minions,
			InstanceID:  fmt.Sprintf("%s_1_1", uuId.String()),
		}
		data, err := json.Marshal(step)
		if utils.CheckError(err) {
			break
		}
		server.Publish(step.Minions, data)
		log.Info("sent msg")
		eventChan := make(chan utils.Event)
		subscribeEvent(opts, "/job", eventChan, timeoutAt)
		for event := range eventChan {
			if event.Function == "job.checkalive" && event.Params == jid {
				log.Info(event.Result)
				isBreak = true
				break
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
					Function: step.Function,
					JID:      jid,
					MinionId: mid,
					Retcode:  utils.Wait,
					JobType:  step.Type,
				})
			}
			returners.UpdateMinion(opts, events, true)
			switch step.Type {
			case utils.CmdType:
				cmdJob(&step, server)
			case utils.FileType:
				fileJob(&step, server)
			case utils.SqlType:
				SqlJob(&step, server)
			}
			time.Sleep(time.Second)
			isSuccess := checkJobStatus(opts, jid, server, step.Minions)
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
