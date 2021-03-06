package btgo

import (
	"../btgo/bencode"
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	AnnounceList [][]string `bencode:"announce-list"`
	Info         Info       `bencode:"info"`
	CreationDate int64      `bencode:"creation date,omitempty"`
	Comment      string     `bencode:"comment,omitempty"`
	CreatedBy    string     `bencode:"created by,omitempty"`
}

type Torrent struct {
	MetaInfo MetaInfo
}

func (info *Info) GenPieces(files []File) (err error) {
	var (
		fi     *os.File
		pieces []byte
	)
	for _, f := range files {
		path := append([]string{"/"}, f.Path...)
		fi, err = os.Open(
			filepath.Join(path...))
		defer fi.Close()
		if err != nil {
			return
		}
		for buf, reader := make([]byte, info.PieceLength), bufio.NewReader(fi); ; {
			h := sha1.New()
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
			}
			h.Write(buf)
			pieces = h.Sum(pieces)
			if int64(n) < info.PieceLength {
				break
			}
		}
	}
	info.Pieces = pieces
	return
}

func NewTorrent(jid string, files []string, announceList [][]string) (t *Torrent, err error) {
	info := Info{Name: jid, PieceLength: PieceLength}
	metaInfo := MetaInfo{
		Info:         info,
		AnnounceList: announceList,
		Comment:      "",
		CreatedBy:    "go agent",
		CreationDate: time.Now().Unix()}
	for _, f := range files {
		fi, err := os.Stat(f)
		if err != nil {
			return nil, err
		}
		switch mode := fi.Mode(); {
		case mode.IsRegular():
			relPath, err := filepath.Rel("/", f)
			if err != nil {
				return nil, err
			}
			metaInfo.Info.Files = append(
				metaInfo.Info.Files,
				File{
					Length: fi.Size(),
					Path:   strings.Split(relPath, string(filepath.Separator))})
			metaInfo.Info.GenPieces(metaInfo.Info.Files)

		default:
			fmt.Println("directory")
		}
	}
	s, _ := json.Marshal(metaInfo)
	log.Infof("%s", s)
	x, _ := bencode.Marshal(metaInfo)
	tfile, err := os.Create(jid + ".torrent")
	if err != nil {
		return
	}
	defer tfile.Close()
	tfile.Write(x)
	log.Errorf("%s", string(x))
	return

}

func NewTorrentFromFile(torrentFile string) (t *Torrent, err error) {
	var (
		c []byte
		f *os.File
	)
	f, err = os.Open(torrentFile)
	if err != nil {
		return
	}
	defer f.Close()

	for buf, reader := make([]byte, 256*1024), bufio.NewReader(f); ; {
		_, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		c = append(c, buf...)
	}
	mi := MetaInfo{}
	err = bencode.Unmarshal(c, &mi)
	if err != nil {
		return
	}
	t.MetaInfo = mi
	return
}
