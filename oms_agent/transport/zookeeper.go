package transport

import (
	"../config"
	"../utils"
	"encoding/json"
	"github.com/samuel/go-zookeeper/zk"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type MasterNodeData struct {
	MasterIP string
	MasterID string
	Minions  int
}

var (
	instance     *zk.Conn
	once         sync.Once
	MasterPrefix = "/masters"
	JobPrefix    = "/jobs"
)

func ZKConnect(opts *config.MasterOptions) *zk.Conn {

	created := false
	once.Do(func() {
		for {
			client, _, err := zk.Connect(opts.ZKCluster, 10*time.Second)
			if err == nil {
				instance = client
				created = true
				break
			}
		}

	})

	if created {
		log.Debug("create a zk cient instance")
	}
	return instance

}

func NodeRegister(opts *config.MasterOptions) {
	zkClient := ZKConnect(opts)
	if zkClient != nil {
		nodePath := filepath.Join(MasterPrefix, opts.Region, opts.ID)
		regionPath := filepath.Join(MasterPrefix, opts.Region)
		if isTrue, _, _ := zkClient.Exists(MasterPrefix); !isTrue {
			zkClient.Create(MasterPrefix, []byte(MasterPrefix), 0, zk.WorldACL(zk.PermAll))
		}
		if isTrue, _, _ := zkClient.Exists(regionPath); !isTrue {
			zkClient.Create(regionPath, []byte(opts.Region), 0, zk.WorldACL(zk.PermAll))
		}
		if isTrue, _, _ := zkClient.Exists(nodePath); isTrue {
			zkClient.Delete(nodePath, -1)
		}
		masterNodeData := &MasterNodeData{
			opts.PublicIp,
			opts.ID,
			0}
		data, _ := json.Marshal(masterNodeData)
		flags := int32(zk.FlagEphemeral)
		_, err := zkClient.Create(nodePath, data, flags, zk.WorldACL(zk.PermAll))
		utils.CheckError(err)
		log.Info("registered to zookeeper successfully!")
	}
}

func JobRegister(opts *config.MasterOptions, jid string) (*zk.Conn, string, error) {
	var (
		//eventCh <-chan zk.Event
		nodePath string
		err      error
	)
	zkClient := ZKConnect(opts)
	if zkClient != nil {
		nodePath = filepath.Join(JobPrefix, jid)
		if isTrue, _, _ := zkClient.Exists(JobPrefix); !isTrue {
			zkClient.Create(JobPrefix, []byte(JobPrefix), 0, zk.WorldACL(zk.PermAll))
		}
		if isTrue, _, _ := zkClient.Exists(nodePath); !isTrue {
			_, err = zkClient.Create(nodePath, []byte("0"), 0, zk.WorldACL(zk.PermAll))
		}
	}
	return zkClient, nodePath, err
}

func JobUpdate(opts *config.MasterOptions, jid string) error {
	var (
		nodePath string
		count    int
		err      error
	)
	zkClient := ZKConnect(opts)
	log.Debug("xxxxxx")
	if zkClient != nil {
		nodePath = filepath.Join(JobPrefix, jid)
		log.Debug(nodePath)
		if isTrue, _, _ := zkClient.Exists(nodePath); !isTrue {
			log.Debug(nodePath)
			data, stat, err := zkClient.Get(nodePath)
			if !utils.CheckError(err) {
				count, err = strconv.Atoi(string(data))
				count += 1
				_, err = zkClient.Set(nodePath, []byte(strconv.Itoa(count)), stat.Cversion)
			}
		}
	}
	return err
}
