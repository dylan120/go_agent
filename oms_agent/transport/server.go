package transport

import (
	"../config"
)

type ServerChannel interface {
	//Run()
	//NewServer()
	//Close()
	PreFork()
	PostFork(int, func(*config.MasterOptions, []byte) ([]byte, error))
	Publish([]string, []byte)
	//Decode()
	//Handle()
}

type ReqServerChannel struct {
	Opts *config.MasterOptions
}

type PubServerChannel struct {
	Opts *config.MasterOptions
}

func NewReqServerChannel(opts *config.MasterOptions) ServerChannel {
	if opts.Transport == "zeromq" {
		return NewZMQReqServerChannel(opts)
	}
	return nil
}

func NewPubServerChannel(opts *config.MasterOptions) ServerChannel {
	if opts.Transport == "zeromq" {
		return NewZMQPubServerChannel(opts)
	}
	return nil
}

//func (reqServer *ReqServerChannel) Run() {
//	output, err := exec.Command("F:\\Go\\oms_agent\\bin\\nsqd.exe --lookupd-tcp-address=127.0.0.1:4160").Output()
//	if err != nil {
//		fmt.Println(err.Error())
//	}
//	fmt.Println(string(output))
//}
