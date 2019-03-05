package transport

import (
	"../config"
)

type ServerChannel interface {
	PreFork()
	PostFork(int, func(*config.MasterOptions, []byte) ([]byte, error))
	Publish([]string, []byte)
}

type ReqServerChannel struct {
	Opts *config.MasterOptions
}

type PubServerChannel struct {
	Opts *config.MasterOptions
}

type EventServerChannel struct {
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
