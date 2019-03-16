package base

import (
	"../auth"
	"../config"
	"../jobs"
	"../returners"
	"../transport"
	"../utils"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

type Master struct {
	Opts *config.MasterOptions
}

func NewMaster(opts *config.MasterOptions) *Master {
	return &Master{
		opts,
	}
}

func HandlePayLoad(opts *config.MasterOptions, msg []byte) (ret []byte, err error) {
	var (
		data []byte
		load utils.Load
	)
	payLoad := utils.Payload{}
	err = utils.UnPackPayload(msg, &payLoad)

	if err == nil {
		err = json.Unmarshal(payLoad.Data, &load)
		if !utils.CheckError(err) {
			switch load.Function {
			case "auth":
				data, err = auth.Auth(opts, &load, false)
			case "reAuth":
				data, err = auth.ReAuth(opts, &load)
			default: //event data
				event := utils.Event{}
				err = json.Unmarshal(load.Data, &event)
				if !utils.CheckError(err) {
					log.Debugf("receive event data: %s", load.Data)
					if strings.HasPrefix(event.Tag, utils.JobTagPrefix) {
						returners.UpdateMinion(opts, []*utils.Event{&event}, true)
					} else if event.Tag == utils.PingTag {
						returners.UpdateMinionStatus(opts, &event, false)
					}
				}
			}
		}
	}
	ret = utils.PackPayload(data, payLoad.Crypt)
	return
}

func (master *Master) Start() {
	utils.GenRSAKeyPairs(master.Opts.PkiDir, master.Opts.Mode, 2048)

	runtime.GOMAXPROCS(runtime.NumCPU())
	manager := utils.NewRunerManager()
	reqChan := transport.NewReqServerChannel(master.Opts)

	manager.Add(reqChan.PreFork)
	j := 0
	for i := master.Opts.WorkerThread; i >= 1; i-- {
		manager.Add(reqChan.PostFork, j, HandlePayLoad)
		j++
	}
	pubChanl := transport.NewPubServerChannel(master.Opts)
	manager.Add(pubChanl.PreFork)
	manager.Add(pubChanl.PostFork, 0, HandlePayLoad)
	manager.Add(transport.NodeRegister, master.Opts)
	manager.Add(jobs.Start, master.Opts, pubChanl)
	manager.Start()
}
