package main

//var wg sync.WaitGroup
import (
	"../base"
	"../config"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func init() {
	//log.EnableDebug = config.Opts.Debug
	//log.LogPath = config.Opts.LogDir

}

func initLog(debugLevel bool) {
	if debugLevel {
		log.SetLevel(log.DebugLevel)
	}

	formatter := &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
		ForceColors:     true,
		FieldMap: log.FieldMap{
			log.FieldKeyFile: "@file",
		},
	}

	log.SetReportCaller(true)
	log.SetFormatter(formatter)
}

func main() {
	//fmt.Println(utils.CheckAlive("121.201.6.10", "223"))
	initLog(config.Opts.Debug)
	mode := os.Args[1]
	if mode == "master" {
		base.NewMaster(&config.Opts).Start()
	} else if mode == "minion" {
		base.NewMinion(&config.MOpts).Start()
	}
}
