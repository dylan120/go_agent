package utils

import (
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"os"
)

func MakeTorrent(f *os.File, btAnnouce []string, srcFile string) error {
	var (
		//f   *os.File
		mi  metainfo.MetaInfo
		err error
	)

	for _, a := range btAnnouce {
		mi.AnnounceList = append(mi.AnnounceList, []string{a})
	}
	mi.AnnounceList = make([][]string, 0)
	mi.SetDefaults()
	info := metainfo.Info{
		PieceLength: 256 * 1024,
	}
	err = info.BuildFromFilePath(srcFile)
	if !CheckError(err) {
		mi.InfoBytes, err = bencode.Marshal(info)
		if !CheckError(err) {
			err = mi.Write(f)
			CheckError(err)
		}
	}
	return err
}

func Download(srcMaster []string, mtgt []string,
	srcFile string, torrentFile *os.File, md5 string) {

}
