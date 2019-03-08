package utils

import (
	"github.com/vmihailenco/msgpack"
)

type (
	ip string
	//minionID string
)
type FileSource struct {
	Account string `json:"project_id"`
	IPList  []ip   `json:"ip_list"`
	File    string `json:"file"`
	MD5sum  string `json:"md5sum"`
	Size    int    `json:"size"`
	Mode    string `json:"mode"`
}

type Step struct {
	ProjectID      int          `json:"project_id"`
	ScriptID       int          `json:"script_id"`
	IsFinished     bool         `json:"is_finished"`
	BlockName      string       `json:"block_name"`
	ScriptName     string       `json:"script_name"`
	Creator        string       `json:"creater"`
	ScriptContent  string       `json:"script_content"`
	Text           string       `json:"text"`
	InstanceID     string       `json:"step_instance_id"`
	Function       string       `json:"function"`
	Type           int          `json:"type"` // 1 cmd 2 file transport 3 sql
	ScriptParam    string       `json:"script_param"`
	Account        string       `json:"account"`
	Name           string       `json:"name"`
	IsPause        bool         `json:"is_pause"`
	TimeOut        int          `json:"timeout"`
	ScriptType     string       `json:"script_type"`
	FileTargetPath string       `json:"file_target_path"`
	FileSource     []FileSource `json:"file_source"`
	ID             string       `json:"step_id"`
	Minions        []string     `json:"minions"`
}

type Block struct {
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

type TaskData struct {
	Blocks []Block `json:"blocks"`
}

type Message struct {
	Type int `json:"msgType"` // 0 page 1 schedule 2 api
}

type Task struct {
	Msg        Message  `json:"msg"`
	ProjectID  int      `json:"project_id"`
	ID         int      `json:"task_id"`
	Name       string   `json:"task_name"`
	InstanceID string   `json:"task_instance_id"`
	Type       string   `json:"type"`
	Operator   string   `json:"operator"`
	IsSchedule bool     `json:"is_schedule"`
	Data       TaskData `json:"data"`
}

type Event struct {
	Function      string `json:"function"`
	Tag           string `json:"tag"`
	JID           string `json:"jid"`
	MinionId      string `json:"minion_id"`
	Params        string `json:"params"`
	Retcode       int    `json:"retcode"`
	Result        string `json:"result"`
	JobType       int    `json:"job_type"`
	StartTime     int64  `json:"start_time"`
	EndTime       int64  `json:"end_time"`
	TimeConsuming int64  `json:"time_consuming"`
}

type Load struct {
	Function   string   `json:"function"`
	ID         string   `json:"id"`
	Target     []string `json:"tgt"`
	PublibcKey string   `json:"pub_key"`
	Token      string   `json:"token"`
	Data       []byte   `json:"data"`
}

type Payload struct {
	Crypt   string `json:"crypt"`
	Data    []byte `json:"data"`
	Version int64  `json:"version"` //aes version
}

type Jobload struct {
	Crypt string `json:"crypt"`
	Data  Task   `json:"data"`
}

type MasterPayload struct {
	Crypt string
}

type ProcessInfo struct {
	JID       string   `json:"jid"`
	ProcessID int      `json:"process_id"`
	Cmd       []string `json:"cmd"`
}

func Dumps(data interface{}) ([]byte, error) {
	ret, err := msgpack.Marshal(&data)
	return ret, err
}

func Loads(data []byte, out interface{}) error {
	err := msgpack.Unmarshal(data, out)
	return err
}

func PackPayload(msg []byte, crypt string) []byte {
	var (
		payload Payload
		out     []byte
	)
	if crypt == "crypt" {
		payload.Crypt = "crypt"
		cryptedData, version, err := AESEncrypt(msg)
		if !CheckError(err) {
			payload.Data = cryptedData
		} else {
			payload.Data = []byte("")
		}
		payload.Version = version
	} else {
		payload.Crypt = "clear"
		payload.Data = msg
	}
	out, _ = Dumps(payload)
	//out = append(prefix,out...)
	return out
}

func UnPackPayload(msg []byte, payload *Payload) error {
	err := Loads(msg, payload)
	if payload.Crypt == "crypt" {
		clearData, err := AESDecrypt(payload.Data, payload.Version)
		if !CheckError(err) {
			payload.Data = clearData
		} else {
			payload.Data = []byte("")
		}
	}
	return err
}
