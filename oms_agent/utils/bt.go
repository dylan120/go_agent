package utils

import (
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"os"
)

func MakeTorrent(btAnnouce []string, srcFile string, instanceID string) {
	mi := metainfo.MetaInfo{}

	for _, a := range btAnnouce {
		mi.AnnounceList = append(mi.AnnounceList, []string{a})
	}
	mi.AnnounceList = make([][]string, 0)
	mi.SetDefaults()
	info := metainfo.Info{
		PieceLength: 256 * 1024,
	}
	err := info.BuildFromFilePath(srcFile)
	if !CheckError(err) {
		mi.InfoBytes, err = bencode.Marshal(info)
		if !CheckError(err) {
			err = mi.Write(os.Stdout)
			CheckError(err)
		}
	}
}
