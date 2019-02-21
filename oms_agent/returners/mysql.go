package returners

import (
	"../config"
	"../utils"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Uint8JNableScie []uint8

func (u Uint8JNableScie) MarshalJSON() ([]byte, error) {
	var result string
	if u == nil {
		result = "null"
	} else {
		result = string(strings.Join(strings.Fields(fmt.Sprintf("%v", u)), ","))
	}
	return []byte(result), nil
}

type Proxy struct {
	MasterId       string `json:"master_id"`
	Ip             string `json:"master_ip"`
	LastUpdateTime string `json:"last_update_time"`
}

func getConnectArgs(opts *config.MasterOptions) string {
	connUri := opts.Returner.Mysql.User + ":" +
		opts.Returner.Mysql.Passwd + "@tcp(" +
		net.JoinHostPort(opts.Returner.Mysql.Ip, strconv.Itoa(opts.Returner.Mysql.Port)) +
		")/" + opts.Returner.Mysql.DB
	return connUri
}

var instance *sql.DB
var once sync.Once

func Connect(opts *config.MasterOptions) *sql.DB {
	created := false
	once.Do(func() {
		connUri := getConnectArgs(opts)
		db, err := sql.Open("mysql", connUri)
		db.SetMaxOpenConns(2000)
		db.SetMaxIdleConns(1000)
		utils.CheckError(err)

		instance = db
		created = true
	})
	if err := instance.Ping(); err != nil {
		log.Info("db instance lost, try to reconnect")
		connUri := getConnectArgs(opts)
		db, err := sql.Open("mysql", connUri)
		utils.CheckError(err)
		instance = db
		created = true
	}

	if created {
		log.Info("create a db instance")

	} else {
		log.Info("reuse a db instance")
	}
	return instance
}

func IterRows(rows *sql.Rows, columnObj ...interface{}) error {
	var (
		noResult bool
		err      error = nil
	)
	for rows.Next() {
		if err = rows.Scan(columnObj...); err != nil {
			if noResult {
				err = utils.RowNoFound
			}
			break
		}
		noResult = true
	}
	return err
}

//func SafelyClose(rows *sql.Rows, err error) {
//	if err == nil {
//		defer rows.Close()
//	}
//}

func GetAllProxies(opts *config.MasterOptions, region string) []Proxy {
	//TODO

	var (
		querySql string
		proxy    Proxy
		result   []Proxy
	)
	db := Connect(opts)
	if region != "" {
		querySql = `SELECT master_id,ip,last_update_time FROM master where 
              agent_mode = "oms_agent" and MINUTE (timediff( now(), last_update_time) )  < ? and region = ?`
	} else {
		querySql = `SELECT master_id,ip,last_update_time FROM master where 
                      agent_mode = "oms_agent" and MINUTE (timediff( now(), last_update_time) )  < ? and region != ?`
	}

	rows, err := db.Query(querySql, 1.5*float32(opts.PingInterval), region)
	utils.CheckError(err)
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&proxy.MasterId, &proxy.Ip, &proxy.LastUpdateTime); err != nil {
			utils.CheckError(err)
		}
		result = append(result, proxy)
	}
	return result
}

func GetMinionPubKeyByID(opts *config.MasterOptions, id string) ([]byte, error) {
	var (
		minionId = ``
		querySql string
	)
	querySql = `SELECT pubkey from  minion WHERE minion_id = ?`
	db := Connect(opts)
	rows, err := db.Query(querySql, id)
	if err == nil {
		defer rows.Close()
		IterRows(rows, &minionId)
	}
	return []byte(minionId), err
}

func UpsertRSAPublicKey(opts *config.MasterOptions, pubKey []byte, minionID string, masterID string) {
	sql := `REPLACE INTO minion 
          (minion_id,master_id,pubkey) VALUES (?,?,?)`
	db := Connect(opts)
	_, err := db.Exec(sql, minionID, masterID, pubKey)
	utils.CheckError(err)
}
