package btgo

import (
	"../btgo/bencode"
	"../utils"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

const PieceLength = 256 * 1024 //256KB

type File struct {
	Length int64
	Path   string
}

type Info struct {
	Name        string
	PieceLength int64
	Pieces      []byte
	Length      int64 `bencode:"length,omitempty"`
	//Path        string `bencode:"path,omitempty"`
	Files []File `bencode:"files,omitempty"`
}

type MetaInfo struct {
	Announce string
	Info     Info
}

type Torrent struct {
	MetaInfo []byte
}

func (info *Info) GenPieces(f string) {
	file, err := os.Open(f)

	if !utils.CheckError(err) {
		buf := make([]byte, info.PieceLength)
		h := sha1.New()
		for {
			file.Seek(info.PieceLength, 0)
			n, _ := file.Read(buf)
			h.Write(buf)
			info.Pieces = h.Sum(info.Pieces)
			if int64(n) < info.PieceLength {
				break
			}
		}

	}

}

func NewTorrent(jid string, files []string) (t *Torrent) {
	info := Info{Name: jid, PieceLength: PieceLength}
	metaInfo := MetaInfo{Info: info, Announce: "http://test.com/"}
	for _, f := range files {
		fi, err := os.Stat(f)
		if !utils.CheckError(err) {
			switch mode := fi.Mode(); {
			case mode.IsRegular():
				info.Files = append(info.Files, File{Length: fi.Size(), Path: f})
				log.Info(f)
				info.GenPieces(f)
				tfile, err := os.Create(jid + ".torrent")
				if !utils.CheckError(err) {
					defer tfile.Close()

				}
			default:
				fmt.Println("directory")
			}
		}
	}
	var err error
	s, _ := json.Marshal(metaInfo)
	log.Infof("%s", s)
	t.MetaInfo, err = bencode.Marshal(metaInfo)
	log.Error(err)
	return t

}
