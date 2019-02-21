package utils

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"runtime"
)

var ErrInterrupt = errors.New("received an interrupt signal")

//var wg sync.WaitGroup

//type base func(i interface{}) string

type RunerManager struct {
	//name      string
	interrupt chan os.Signal
	//stopped    chan interface{}
	//tasks     []func(int)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func NewRunerManager() *RunerManager {
	return &RunerManager{
		interrupt: make(chan os.Signal, 1),
	}
}

func (fm *RunerManager) run() error {
	return nil
}

func gotInterrupt() error {
	var sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case <-sig:
		signal.Stop(sig)
		//fmt.Printf("receive a Interrupt signal\n")
		return ErrInterrupt
		//return true
	}
}

func (fm *RunerManager) Add(fc interface{}, args ...interface{}) {
	go func(fc interface{}, args ...interface{}) {
		funcName := GetFunctionName(fc)
		log.Infof("add a %s function\n", funcName)
		funcValue := reflect.ValueOf(fc)
		if funcValue.Kind() != reflect.Func {
			panic(funcName + " is not a func type")
		}
		input := make([]reflect.Value, len(args))
		for i, arg := range args {
			input[i] = reflect.ValueOf(arg)
		}
		funcValue.Call(input)

	}(fc, args...)
}

func (fm *RunerManager) Start() {
	gotInterrupt()
}

var ForkEnvKey = "go_agent"

func Fork() int {
	args := os.Args[1:]
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = os.Environ()
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.ExtraFiles = nil
	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	Setsid: true,
	//}

	if err := cmd.Start(); err != nil {
		CheckError(err)
		return 0
	}

	//os.Setenv(ForkEnvKey, "1")
	return cmd.Process.Pid
}

func Daemon() {
	pid := Fork()
	if pid != 0 {
		log.Debug(os.Getpid())
	}
}
