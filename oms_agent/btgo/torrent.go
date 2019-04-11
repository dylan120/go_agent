package btgo

import (
	"../btgo/bencode"
	"../utils"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

const PieceLength = 256 * 1024 //256KB

type File struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
}

type Info struct {
	Name        string `bencode:"name"`
	PieceLength int64  `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`
	Length      int64  `bencode:"length,omitempty"`
	Files       []File `bencode:"files,omitempty"`
}

type MetaInfo struct {
	Announce string `bencode:"announce"`
	Info     Info   `bencode:"info"`
}

type Torrent struct {
	MetaInfo []byte
}

func (info *Info) GenPieces(files []File) {
	var pieces []byte
	for _, f := range files {
		fi, err := os.Open(strings.Join(f.Path, "/"))
		//pr, pw := io.Pipe()
		//wn, err := io.CopyN(pw, fi, f.Length)
		fi.Close()

		buf := make([]byte, info.PieceLength)
		if !utils.CheckError(err) {
			for {
				h := sha1.New()
				//file.Seek(info.PieceLength-1, 0)
				n, err := fi.Read(buf)

				if err != nil {
					if err == io.EOF {
						break
					}
				}
				h.Write(buf)
				//info.Pieces = h.Sum(info.Pieces)
				pieces = h.Sum(pieces)
				//log.Info(pieces)
				if int64(n) < info.PieceLength {
					break
				}
			}

		}
	}
	log.Println(string(pieces))
	//info.Pieces = pieces

}

func NewTorrent(jid string, files []string) (t *Torrent) {
	info := Info{Name: jid, PieceLength: PieceLength}
	metaInfo := MetaInfo{Info: info, Announce: "http://test.com/"}
	//var pieces []byte
	for _, f := range files {
		fi, err := os.Stat(f)
		if !utils.CheckError(err) {
			switch mode := fi.Mode(); {
			case mode.IsRegular():
				metaInfo.Info.Files = append(metaInfo.Info.Files, File{Length: fi.Size(), Path: []string{f}})
				//pieces = append(pieces, metaInfo.Info.GenPieces(f)...)
				metaInfo.Info.GenPieces(metaInfo.Info.Files)

			default:
				fmt.Println("directory")
			}
		}
	}
	//metaInfo.Info.Pieces = pieces
	s, _ := json.Marshal(metaInfo)
	log.Infof("%s", s)
	x, _ := bencode.Marshal(metaInfo)
	tfile, err := os.Create(jid + ".torrent")
	if !utils.CheckError(err) {
		defer tfile.Close()
		tfile.Write(x)
	}
	log.Errorf("%s", string(x))
	return t

}
