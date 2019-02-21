package jobs

import (
	"../utils"
	"encoding/json"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
	"strconv"
	"sync"
)

type Broker struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Passwd string `json:"passwd"`
	DB     string `json:"db"`
}

var (
	client interface{}
	once   sync.Once
)

func GetRedis(broker *Broker) interface{} {
	db, _ := strconv.Atoi(broker.DB)
	cli := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(broker.Host, strconv.Itoa(broker.Port)),
		Password: broker.Passwd, // no password set
		DB:       db,            // use default DB
	})
	if _, err := cli.Ping().Result(); err != nil {
		log.Panic(err)
	}
	return cli
}

func (broker *Broker) Init() {
	switch broker.Name {
	case "redis":
		client = GetRedis(broker)
	}
}

func (broker *Broker) Get() (utils.Task, error) {
	var (
		jobload = utils.Jobload{}
		//job     = utils.Job{}
		data string

		err error
	)
	if client == nil {
		panic("broker client did not init first")
	}
	switch broker.Name {
	case "redis":
		cli := reflect.ValueOf(client).Interface().(*redis.Client)
		cmd := cli.LPop("oms_jobs")
		data, err = cmd.Result()
		if err == nil {
			//err = utils.Loads([]byte(data), &s)
			err = json.Unmarshal([]byte(data), &jobload)
			log.Debug(jobload)
			//err = json.Unmarshal(jobload.Data, &job)
			//log.Debug(err)
			//log.Debug(job)
		}
	}
	return jobload.Data, err
}
