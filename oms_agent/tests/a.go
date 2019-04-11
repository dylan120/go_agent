package main

import (
	"../btgo"
	"../btgo/bencode"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

func BencodeTest() {
	s, _ := bencode.Marshal(map[string]interface{}{
		"a": 1, "d": "xxx", "c": true, "b": []interface{}{"a", 2}})
	fmt.Printf("%s", s)
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
	initLog(true)
	BencodeTest()
	//fmt.Println(btgo.NewTorrent("1", []string{"/tmp/44e67cf3-4c48-4e41-a8ee-a781adfd97cd_1_1_0.sh"}))
	fmt.Println(btgo.NewTorrent("1", []string{"/tmp/test_1_1_1"}, []string{"http://192.168.0.206:8760/announce"}))
}
