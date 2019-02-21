package transport

import (
	"../config"
)



type ReqClientChannel interface {
	Send(msg []byte)
}

type PubClientChannel interface {
	Connect()
}

func NewReqClientChannel(opts *config.MinionOptions, crypt string) interface{} {
	switch opts.Transport {
	case "zeromq":
		return NewZMQReqClientChannel(opts, crypt)
	default:
		return nil
	}

}

func NewPubClientChannel(opts *config.MinionOptions, crypt string) interface{} {
	switch opts.Transport {
	case "zeromq":
		return NewZMQPubClientChannel(opts, crypt)
	default:
		return nil
	}

}


