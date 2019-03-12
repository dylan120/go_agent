package utils

import (
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"os"
	"path/filepath"
	"strings"
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
			base := filepath.Join("/tmp", strings.Join([]string{instanceID, ".torrent"}, "."))
			f, err := os.OpenFile(base, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			defer f.Close()
			err = mi.Write(f)
			CheckError(err)
		}
	}
}
