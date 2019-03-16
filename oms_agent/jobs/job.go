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
	"io/ioutil"
	"os"
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

func fileJob(step *utils.Step, opts *config.MasterOptions, funcMap map[string]interface{}, server transport.ServerChannel) {
	log.Info("run file job")
	fileSource := step.FileSource
	if len(fileSource) == 0 {
		log.Error(defaults.InValidFileArg, errors.New("file source is null"))
	}

	base := filepath.Join("/tmp", step.InstanceID, strings.Join([]string{step.InstanceID, "torrent"}, "."))
	if _, err := os.Stat(base); os.IsNotExist(err) {
		os.Mkdir(base, 0555)
	}
	for _, fs := range fileSource {
		//step.FileSource[i].MD5sum
		srcFile := strings.TrimSpace(fs.File)
		if fs.Mode == "web_agent" {
			log.Debugf("init local file transfer.")
			// mtgt = get_masters(web_minions, step['minions'], oms_client.opts)
			torrentDir := filepath.Join("/tmp", step.InstanceID)
			if _, err := os.Stat(torrentDir); os.IsNotExist(err) {
				err := os.MkdirAll(torrentDir, 0755)
				utils.CheckError(err)
			}
			torrentPath := filepath.Join(torrentDir,
				strings.Join([]string{step.InstanceID, "torrent"}, "."))
			f, err := os.OpenFile(torrentPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0400)
			defer f.Close()
			err = funcMap["bt.maketorrent"].(func(*os.File, []string, string) error)(
				f,
				opts.BtAnnouce,
				srcFile,
			)
			if !utils.CheckError(err) {
				md5, err := utils.MD5sum(srcFile)
				if !utils.CheckError(err) {
					var (
					//torrentStream []byte
					//r             = bufio.NewReader(f)
					//streamChan = make([]byte, 1024)
					)
					//for {
					//	n, err := r.Read(streamChan)
					//	if err != nil && err != io.EOF {
					//		log.Error(err)
					//		break
					//	}
					//	if 0 == n {
					//		break
					//	}
					//
					//	//fmt.Println(string(buf[:n]))
					//}
					torrentStream, err := ioutil.ReadFile(torrentPath)
					step.Function = "bt.download"
					step.FileParam = []string{string(torrentStream), md5}
					data, err := json.Marshal(step)
					if !utils.CheckError(err) {
						server.Publish(step.Minions, data)
						funcMap["bt.mdownload"].(func([]string, []string,
							string, string, string))(
							[]string{opts.ID}, []string{opts.ID},
							torrentPath, md5, filepath.Dir(srcFile))
					}
				}
			}

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

func run(opts *config.MasterOptions, task *utils.Task, funcMap map[string]interface{}, server transport.ServerChannel) {
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
				fileJob(&step, opts, funcMap, server)
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
	funcMap := utils.LoadPlugins(opts)
	broker := Broker{}
	err := json.Unmarshal([]byte(opts.JobBroker), &broker)
	utils.RaiseError(err)
	broker.Init()
	for {
		task, err := broker.Get()
		if err == nil {
			log.Debugf("receive job with id %s", task.InstanceID)
			go run(opts, &task, funcMap, server)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
