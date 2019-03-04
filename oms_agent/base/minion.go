package base

import (
	"../config"
	"../defaults"
	"../returners"
	"../transport"
	"../utils"
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

var JobTagPrefix = "/job"

type Minion struct {
	Opts    *config.MinionOptions
	funcMap map[string]interface{}
}

func NewMinion(opts *config.MinionOptions) *Minion {
	return &Minion{
		Opts:    opts,
		funcMap: utils.LoadPlugins(opts),
	}
}

func SelectAliveMaster(masters []config.Master, retPort int) (string, error) {
	var (
		masterIP = ""
		//masterID string = ""
		err error = nil
	)
	for _, master := range masters {
		fmt.Printf("try to connect to %s:%d\n", master.MasterIP, retPort)
		isAlive := utils.CheckAlive(master.MasterIP, strconv.Itoa(retPort))
		if isAlive {
			masterIP = master.MasterIP
			//masterID = master.MasterID
			break
		} else {
			err = utils.ConnectFailed
		}
	}
	return masterIP, err
}

func (minion *Minion) ConnectMaster(opts *config.MinionOptions) {
	//isConnected := false
	//var subSock *zmq.Socket = nil
	masterIP, err := SelectAliveMaster(opts.Masters, opts.RetPort)
	opts.MasterIP = masterIP
	if utils.CheckError(err) {
		fmt.Errorf("failed to connect all masters")
	} else {
		pubClient := transport.NewPubClientChannel(opts, "crypt")
		ret := utils.RunReflectArgsFunc(pubClient, "Connect")
		subSock := ret[0].Interface().(*zmq.Socket)
		for {
			log.Println("minion ready to receive!")
			recvPayLoad, err := subSock.RecvBytes(0)
			if !utils.CheckError(err) {
				err := minion.HandlePayload(recvPayLoad)
				if utils.CheckError(err) {
					if err == utils.DecryptDataFailure {
						subSock.Close()
						log.Warnf("retry to reauth")
						ret = utils.RunReflectArgsFunc(pubClient, "Connect")
						subSock = ret[0].Interface().(*zmq.Socket)
					}
				}
			}
		}
	}
}

func (minion *Minion) CheckPayload(load *utils.Load) bool {
	return utils.SliceExists(load.Target, minion.Opts.ID)
}

func (minion *Minion) HandlePayload(recvPayLoad []byte) error {

	var (
		err       error
		payload   utils.Payload
		load      utils.Load
		step      utils.Step
		clearLoad []byte
	)
	//recvPayLoad, err := subSock.RecvBytes(0)
	err = utils.Loads(recvPayLoad, &payload)
	if !utils.CheckError(err) {
		if payload.Crypt == "crypt" {
			clearLoad, err = utils.AESDecrypt(payload.Data, payload.Version)
			if !utils.CheckError(err) {
				err = json.Unmarshal(clearLoad, &load)
				if !utils.CheckError(err) {
					err = json.Unmarshal(load.Data, &step)
					if !utils.CheckError(err) {
						if err == nil {
							if minion.CheckPayload(&load) {
								log.Infof("receive job with id %s : %s", step.InstanceID, step.Function)
								go minion.doTask(step.Function, step)
							}
						} else {
							log.Errorf("receive unexpected data structure")
						}
					}
				}
			}
		}
	}
	return err
}

func EventTag(prefix string, jid string, minionId string, seq int) string {
	return filepath.Join(prefix, jid, minionId, strconv.Itoa(seq))
}

func (minion *Minion) fireEvent(tag string, event *utils.Event) bool {
	var (
		load = utils.Load{
			ID: minion.Opts.ID,
			//Data: event,
		}
		isTrue = false
	)
	log.Debug(event)
	data, err := json.Marshal(event)
	if !utils.CheckError(err) {
		load.Data = data
		msg, _ := json.Marshal(load)
		reqClient := transport.NewReqClientChannel(minion.Opts, "crypt")
		utils.RunReflectArgsFunc(reqClient, "Send", msg)
		isTrue = true
	}
	return isTrue
}

func (minion *Minion) doTask(funcName string, step utils.Step) {
	if fun, ok := minion.funcMap[funcName]; ok {
		resultChannel := make(chan string)
		nowTimestamp := time.Now().Unix()
		timeOutAt := nowTimestamp + int64(step.TimeOut)*2 //set max timeout
		status := defaults.NewStatus()
		go fun.(func(utils.Step, string, chan string, *defaults.Status))(
			step, minion.Opts.ProcDir, resultChannel, status)

		seq := 0
		isBreak := false
		for {
			if isBreak {
				break
			}
			select {
			case result := <-resultChannel:
				log.Debug(result)
				tag := EventTag(JobTagPrefix, step.InstanceID, minion.Opts.ID, seq)
				event := utils.Event{
					Function:  funcName,
					Params:    step.ScriptParam,
					Tag:       tag,
					StartTime: nowTimestamp,
					MinionId:  minion.Opts.ID,
					JID:       step.InstanceID,
					Result:    result,
					Retcode:   defaults.Run,
				}
				minion.fireEvent(tag, &event)
				seq += 1
			default:
				if status.IsFinished == true || time.Now().Unix() > timeOutAt {
					close(resultChannel)
					isBreak = true
				}
			}
		}

		tag := EventTag(JobTagPrefix, step.InstanceID, minion.Opts.ID, -1)
		event := utils.Event{
			Function:  funcName,
			Params:    step.ScriptParam,
			Tag:       tag,
			MinionId:  minion.Opts.ID,
			JID:       step.InstanceID,
			Retcode:   status.Code,
			StartTime: nowTimestamp,
			EndTime:   time.Now().Unix(),
		}
		if status.Code != defaults.Success {
			event.Result = status.Desc
		} else {
			log.Debugf("job %s exit with code %d", step.Function, status.Code)
		}
		minion.fireEvent(tag, &event)
	}
}

func test(opts *config.MinionOptions) {
	tgt := filepath.Join(opts.PkiDir, "minion.pub")
	content, _ := ioutil.ReadFile(tgt)
	returners.UpsertRSAPublicKey(&config.Opts, content, opts.ID, config.Opts.ID)
}

func (minion *Minion) Start() {
	utils.GenRSAKeyPairs(minion.Opts.PkiDir, minion.Opts.Mode, 2048)
	test(minion.Opts)
	minion.ConnectMaster(minion.Opts)
}
