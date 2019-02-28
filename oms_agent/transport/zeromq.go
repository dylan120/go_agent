package transport

import (
	"../config"
	"../utils"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type ZMQReqClientChannel struct {
	Opts      *config.MinionOptions
	Crypt     string
	MasterUri string
}

type ZMQPubClientChannel struct {
	opts      *config.MinionOptions
	crypt     string
	MasterUri string
	Auth      *MinionAuth
}

type ZMQReqServerChannel struct {
	ReqServerChannel
	eventSock *zmq.Socket
	closing   bool
}

type ZMQPubServerChannel struct {
	PubServerChannel
	pubSock *zmq.Socket
	//redis   *redis.Client
	closing bool
}

//type ZMQEventServerChannel struct {
//	EventServerChannel
//	eventSock *zmq.Socket
//}

func SetBit(b bool) int {
	var bitSetVar int
	if b {
		bitSetVar = 1
	}
	return bitSetVar
}

func NewZMQReqClientChannel(opts *config.MinionOptions, crypt string) *ZMQReqClientChannel {
	//masterUri :=
	return &ZMQReqClientChannel{
		opts,
		crypt,
		"tcp://" + net.JoinHostPort(opts.MasterIP, strconv.Itoa(opts.RetPort)),
	}
}

func NewZMQPubClientChannel(opts *config.MinionOptions, crypt string) *ZMQPubClientChannel {
	proxyUri := "tcp://" + net.JoinHostPort(opts.MasterIP, strconv.Itoa(opts.RetPort))
	return &ZMQPubClientChannel{
		opts,
		crypt,
		proxyUri,
		NewMinionAuth(opts),
	}
}

func (reqClient *ZMQReqClientChannel) Send(data []byte) utils.Payload {
	var (
		msg []byte
		ret = utils.Payload{}
	)

	context, _ := zmq.NewContext()
	defer context.Term()

	// Socket to talk to clients
	reqSock, _ := context.NewSocket(zmq.REQ)
	reqSock.Connect(reqClient.MasterUri)
	defer reqSock.Close()
	msg = utils.PackPayload(data, reqClient.Crypt)
	reqSock.SendBytes(msg, 0)
	resp, err := reqSock.RecvBytes(0)
	if err == nil {
		err = utils.UnPackPayload(resp, &ret)
		if !utils.CheckError(err) {
			log.Debugf("收到 %s", ret)
		}
	}

	return ret

}

func (pubClient *ZMQPubClientChannel) Connect() (*zmq.Socket, error) {
	var (
		subSock *zmq.Socket = nil
		err     error       = nil
	)
	context, _ := zmq.NewContext()
	//defer context.Term()
	// Socket to talk to clients
	subSock, err = context.NewSocket(zmq.SUB)
	if !utils.CheckError(err) {
		//defer subSock.Close()
		if !pubClient.Auth.IsAuthenticated {
			log.Info("try to connect master to get auth")
			err = pubClient.Auth.Authenticate(false)
		}
		pubUri := "tcp://" + net.JoinHostPort(pubClient.Auth.MasterIp, strconv.Itoa(pubClient.Auth.PublishPort))
		log.Info(pubUri)
		err = subSock.Connect(pubUri)
		subSock.SetSubscribe("")
	}
	return subSock, err
}

func setTcpKeepalive(zmqSock *zmq.Socket, opts *config.MasterOptions) {
	zmqSock.SetTcpKeepalive(SetBit(opts.TCPKeepAlive))
	zmqSock.SetTcpKeepaliveIdle(opts.TcpKeepAliveIdle)
	zmqSock.SetTcpKeepaliveCnt(opts.TcpKeepAliveCnt)
	zmqSock.SetTcpKeepaliveIntvl(opts.TcpKeepAliveIntvl)
}

func NewZMQReqServerChannel(opts *config.MasterOptions) *ZMQReqServerChannel {
	return &ZMQReqServerChannel{
		ReqServerChannel: ReqServerChannel{opts},
		closing:          false,
	}
}

func NewZMQPubServerChannel(opts *config.MasterOptions) *ZMQPubServerChannel {
	return &ZMQPubServerChannel{
		PubServerChannel: PubServerChannel{opts},
		//redis:            client,
		closing: false,
	}
}

//func NewZMQEventServerChannel(opts *config.MasterOptions) *ZMQEventServerChannel {
//	return &ZMQEventServerChannel{
//		EventServerChannel: EventServerChannel{opts},
//	}
//}

func (reqServer *ZMQReqServerChannel) PreFork() {
	context, _ := zmq.NewContext()
	defer context.Term()
	router, _ := context.NewSocket(zmq.ROUTER)
	defer router.Close()
	routerUri := "tcp://" + net.JoinHostPort(reqServer.Opts.PublicIp, strconv.Itoa(reqServer.Opts.RetPort))
	router.Bind(routerUri)

	context1, _ := zmq.NewContext()
	defer context1.Term()
	dealer, _ := context1.NewSocket(zmq.DEALER)
	defer dealer.Close()
	if _, err := os.Stat(reqServer.Opts.SockDir); os.IsNotExist(err) {
		os.Mkdir(reqServer.Opts.SockDir, os.ModePerm) //TODO
	}
	dealer.Bind("ipc://" + filepath.Join(reqServer.Opts.SockDir, "dealer.ipc"))

	context2, _ := zmq.NewContext()
	defer context2.Term()
	reqServer.eventSock, _ = context.NewSocket(zmq.PUB)
	defer reqServer.eventSock.Close()
	setTcpKeepalive(reqServer.eventSock, reqServer.Opts)
	reqServer.eventSock.SetRcvhwm(1000) //TODO
	reqServer.eventSock.SetSndhwm(1000)
	pubUri := "ipc://" + filepath.Join(reqServer.Opts.SockDir, "event_publish.ipc")
	reqServer.eventSock.Bind(pubUri)

	log.Info("start request server...")
	zmq.Proxy(router, dealer, nil)
}

func (reqServer *ZMQReqServerChannel) PostFork(i int, handlePayLoad func(*config.MasterOptions, []byte) ([]byte, error)) {
	context, _ := zmq.NewContext()
	defer context.Term()

	// Socket to talk to clients
	repSock, _ := context.NewSocket(zmq.REP)
	defer repSock.Close()
	repSock.Connect("ipc://" + filepath.Join(reqServer.Opts.SockDir, "dealer.ipc"))
	for {
		recvMsg, _ := repSock.RecvBytes(0)
		reqServer.eventSock.SendBytes(recvMsg, 0)
		//log.Debugf("[worker %d]Received request: [%s]\n", i, recvMsg)
		out, _ := handlePayLoad(reqServer.Opts, recvMsg) //TODO can make this a async task ?
		repSock.SendBytes(out, 0)
	}
}

func (reqServer *ZMQReqServerChannel) Publish(target []string, data []byte) {}

func (pubServer *ZMQPubServerChannel) PreFork() {
	context, _ := zmq.NewContext()
	defer context.Term()
	pubServer.pubSock, _ = context.NewSocket(zmq.PUB)
	defer pubServer.pubSock.Close()
	setTcpKeepalive(pubServer.pubSock, pubServer.Opts)
	pubServer.pubSock.SetRcvhwm(1000) //TODO
	pubServer.pubSock.SetSndhwm(1000)
	pubUri := "tcp://" + net.JoinHostPort(pubServer.Opts.PublicIp, strconv.Itoa(pubServer.Opts.PublishPort))

	//err := pubSock.Monitor("inproc://monitor.rep", zmq.EVENT_ACCEPTED)
	//if !utils.CheckError(err) {
	//	go socketMonitor("inproc://monitor.rep")
	//}
	pubServer.pubSock.Bind(pubUri)

	//pullSock, _ := context.NewSocket(zmq.PULL)
	//defer pullSock.Close()
	//pullUri := "ipc://" + filepath.Join(pubServer.Opts.SockDir, "publish_pull.ipc")
	//pullSock.Bind(pullUri)
	for {
		time.Sleep(100 * time.Millisecond)
	}
}

func (pubServer *ZMQPubServerChannel) PostFork(
	i int, handlePayLoad func(*config.MasterOptions, []byte) ([]byte, error)) {
}

func (pubServer *ZMQPubServerChannel) Publish(target []string, data []byte) {
	load := utils.Load{Target: target}
	load.Data = data
	ret, _ := json.Marshal(load)
	_, err := pubServer.pubSock.SendBytes(utils.PackPayload(ret, "crypt"), 0)
	utils.CheckError(err)
	log.Info("sent msg")
}

//func (eventServer *ZMQEventServerChannel) PreFork() {
//	context, _ := zmq.NewContext()
//	defer context.Term()
//	eventServer.eventSock, _ = context.NewSocket(zmq.PUB)
//	defer eventServer.eventSock.Close()
//	setTcpKeepalive(eventServer.eventSock, eventServer.Opts)
//	eventServer.eventSock.SetRcvhwm(1000) //TODO
//	eventServer.eventSock.SetSndhwm(1000)
//	pubUri := "ipc://" + filepath.Join(eventServer.Opts.SockDir, "event_publish.ipc")
//	eventServer.eventSock.Bind(pubUri)
//	for {
//		time.Sleep(100 * time.Millisecond)
//	}
//}

func socketMonitor(addr string) {
	log.Info("create socket monitor...")
	s, err := zmq.NewSocket(zmq.PAIR)
	if !utils.CheckError(err) {
		err = s.Connect(addr)
		if !utils.CheckError(err) {
			log.Info("socket monitor start running!")
			for {
				a, b, c, err := s.RecvEvent(0)
				if err != nil {
					utils.CheckError(err)
					break
				}
				log.Info(a, b, c)
			}
			s.Close()
		}
	}
}
