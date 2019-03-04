package transport

import (
	"../auth"
	"../config"
	"../utils"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type MinionAuth struct {
	Opts            *config.MinionOptions
	IsAuthenticated bool
	AESKey          []byte
	MasterPubKey    string
	MasterIp        string
	PublishPort     int
	Token           string
}

func NewMinionAuth(opts *config.MinionOptions) *MinionAuth {
	return &MinionAuth{
		Opts:            opts,
		IsAuthenticated: false,
		AESKey:          []byte(``),
		MasterPubKey:    "master",
		MasterIp:        "",
		PublishPort:     33411,
		Token:           string(utils.GenToken(64)),
	}
}

func (mauth *MinionAuth) SignIn(reAuth bool) error {
	var (
		pubKey = utils.GetPublicKey(mauth.Opts.PkiDir, "minion")
		load   = utils.Load{
			ID:         mauth.Opts.ID,
			PublibcKey: utils.Strings(&pubKey),
		}
		aesKey  []byte
		retLoad auth.AuthLoad
		err     error
	)
	if !reAuth {
		log.Info("Try to authenticate\n")
		load.Function = "auth"
	} else {
		log.Info("Try to reauthenticate\n")
		load.Function = "reAuth"
	}
	mpubKey := utils.GetPublicKey(mauth.Opts.PkiDir, "master")
	log.Debugf("加密token %s ", mauth.Token)
	token, err := utils.RSAEncrypt(mpubKey, mauth.Token)
	utils.CheckError(err)
	load.Token = token

	for {
		reqClient := NewReqClientChannel(mauth.Opts, "clear")
		msg, _ := json.Marshal(load)
		ret := utils.RunReflectArgsFunc(reqClient, "Send", msg)
		if ret[0].Type().Name() == "Payload" {
			payload := ret[0].Interface().(utils.Payload)
			err = json.Unmarshal(payload.Data, &retLoad)

			if !utils.CheckError(err) {
				privKey := utils.GetPrivateKey(mauth.Opts.PkiDir, "minion")
				token, err := utils.RSADecrypt(privKey, retLoad.Token)
				if !utils.CheckError(err) {
					if mauth.Token == token {

						if retLoad.MasterIP != mauth.Opts.MasterIP {
							log.Infof("change master ip to %s to sign in", retLoad.MasterIP)
							mauth.Opts.MasterIP = retLoad.MasterIP
							continue
						} else {

							aesString, err := utils.RSADecrypt(privKey, retLoad.AESKey)
							if !utils.CheckError(err) {
								aesKey = []byte(aesString)
								utils.SetAESKey(aesKey, retLoad.Version)
								mauth.MasterIp = retLoad.MasterIP
								mauth.PublishPort = retLoad.PublishPort
								mauth.IsAuthenticated = true
								break
							}
						}
					} else {
						log.Error("token mismatch")
					}
				}
			}
		}
		mauth.MasterIp = ""
		err = utils.AuthFailure
		break
	}
	return err
}

func (mauth *MinionAuth) Authenticate(reAuth bool) error {
	err := mauth.SignIn(reAuth)
	return err

}
