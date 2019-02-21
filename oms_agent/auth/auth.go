package auth

import (
	"../config"
	"../returners"
	"../utils"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type AuthLoad struct {
	Crypt       string `json:"crypt"`
	PubKey      []byte `json:"pub_key"`
	MasterIp    string `json:"master_ip"`
	PublishPort int    `json:"publish_port"`
	AESKey      string `json:"aes_key"`
	Token       string `json:"token"`
}

func SelectOptimalMaster() string {
	return "192.168.0.206"
}

func Auth(opts *config.MasterOptions, load *utils.Load, reAuth bool) ([]byte, error) {
	var (
		ret  []byte
		auth = AuthLoad{
			"crypt",
			[]byte(""),
			"",
			0,
			"",
			""}
		err       error
		master_ip string
	)
	log.Infof("receive auth request from %s\n", load.ID)
	validPubKey, err := returners.GetMinionPubKeyByID(opts, load.ID)
	if utils.CheckError(err) {
		err = utils.AuthFailure
	} else {
		//pubKey := strings.TrimSpace(string(load.PublibcKey))
		if load.PublibcKey == utils.Strings(&validPubKey) {
			if !reAuth && opts.Mode == "master" {
				master_ip = SelectOptimalMaster()
			} else {
				master_ip = opts.PublicIp
			}
			if len(load.Token) != 0 {
				privateKey := utils.GetPrivateKey(opts.PkiDir, "master")
				token, err := utils.RSADecrypt(privateKey, load.Token)
				utils.CheckError(err)
				log.Debugf("加密token %s for %s", token, load.ID)
				encyptToken, err := utils.RSAEncrypt(validPubKey, token)
				utils.CheckError(err)
				auth.Token = encyptToken
			}

			auth.AESKey, err = utils.RSAEncrypt(validPubKey, string(utils.GetAESKey()))
			if err == nil {
				auth.Crypt = "clear"
				auth.PubKey = utils.GetPublicKey(opts.PkiDir, "master")
				auth.MasterIp = master_ip
				auth.PublishPort = opts.PublishPort
			}
		} else {
			err = utils.MissMatchPubPkey
		}
	}
	if !utils.CheckError(err) {
		ret, err = json.Marshal(auth)
		utils.CheckError(err)
	}
	return ret, err

}

func ReAuth(opts *config.MasterOptions, load *utils.Load) ([]byte, error) {
	return Auth(opts, load, true)
}
