package main

import (
	"../defaults"
	"../utils"
	"encoding/base64"
	"expvar"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/dustin/go-humanize"
	"github.com/gosuri/uiprogress"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var progress = uiprogress.New()

func MakeTorrent(f *os.File, btAnnouce []string, srcFile string) (err error) {
	var (
		mi metainfo.MetaInfo
	)
	mi.SetDefaults()
	mi.AnnounceList = make([][]string, 0)
	for _, a := range btAnnouce {
		mi.AnnounceList = append(mi.AnnounceList, []string{a})
	}
	info := metainfo.Info{
		PieceLength: 256 * 1024,
	}
	err = info.BuildFromFilePath(srcFile)
	if !utils.CheckError(err) {
		mi.InfoBytes, err = bencode.Marshal(info)
		if !utils.CheckError(err) {
			err = mi.Write(f)
			//info, err := mi.UnmarshalInfo()
			//if !CheckError(err) {
			//	return mi.Magnet(info.Name, mi.HashInfoBytes()).String(), nil
			//}
		}
	}
	return
}

func stdoutAndStderrAreSameFile() bool {
	fi1, _ := os.Stdout.Stat()
	fi2, _ := os.Stderr.Stat()
	return os.SameFile(fi1, fi2)
}

func torrentBar(t *torrent.Torrent) {
	bar := progress.AddBar(1)
	bar.AppendCompleted()
	bar.AppendFunc(func(*uiprogress.Bar) (ret string) {
		select {
		case <-t.GotInfo():
		default:
			return "getting info"
		}
		if t.Seeding() {
			return "seeding"
		} else if t.BytesCompleted() == t.Info().TotalLength() {
			return "completed"
		} else {
			return fmt.Sprintf("downloading (%s/%s)",
				humanize.Bytes(uint64(t.BytesCompleted())),
				humanize.Bytes(uint64(t.Info().TotalLength())))
		}
	})
	bar.PrependFunc(func(*uiprogress.Bar) string {
		return t.Name()
	})
	go func() {
		<-t.GotInfo()
		tl := int(t.Info().TotalLength())
		if tl == 0 {
			bar.Set(1)
			return
		}
		bar.Total = tl
		for {
			bc := t.BytesCompleted()
			bar.Set(int(bc))
			time.Sleep(time.Second)
		}
	}()
}

func addTorrents(client *torrent.Client, torrentStream string) {
	t := func() *torrent.Torrent {
		metaInfo, err := metainfo.LoadFromFile(torrentStream)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading torrent file %q: %s\n", torrentStream, err)
			os.Exit(1)
		}
		t, err := client.AddTorrent(metaInfo)
		if err != nil {
			log.Error(err)
		}
		return t
	}()
	torrentBar(t)
	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()
}

func MDownload(srcMaster []string, mtgt []string,
	torrentStream string, md5 string, fileTargetPath string) {
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.Debug = true
	clientConfig.Seed = true
	//clientConfig.DataDir = fileTargetPath
	clientConfig.DefaultStorage = storage.NewMMap(fileTargetPath)
	clientConfig.NoDHT = true

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		log.Errorf("error creating client: %s", err)
	}
	defer client.Close()

	if stdoutAndStderrAreSameFile() {
		log.SetOutput(progress.Bypass())
	}
	progress.Start()
	addTorrents(client, torrentStream)
	if client.WaitAll() {
		log.Info("downloaded ALL the torrents")
	} else {
		log.Error("y u no complete torrents?!")
	}
	outputStats(client)
	log.Info("downloaded ALL the torrents")
	time.Sleep(60 * time.Second)
	log.Infof("seeded %d s ALL the torrents", 60)
	outputStats(client)
	//client.Close()
}

func Download(step utils.Step, _ string, _ chan string, status *defaults.Status) {
	clientConfig := torrent.NewDefaultClientConfig()
	t := step.FileParam[0]
	log.Println(t)
	torrentStream, _ := base64.StdEncoding.DecodeString(step.FileParam[0].(string))
	torrentPath := filepath.Join("/tmp", strings.Join([]string{step.InstanceID, "torrent"}, "."))
	f, err := os.OpenFile(torrentPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0400)
	defer f.Close()
	f.Write(torrentStream)
	//md5 := step.FileParam[1]
	fileTargetPath := step.FileTargetPath
	clientConfig.Debug = true
	clientConfig.Seed = true
	clientConfig.DefaultStorage = storage.NewMMap(fileTargetPath)

	clientConfig.NoDHT = true

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		log.Errorf("error creating client: %s", err)
	}
	defer client.Close()

	if stdoutAndStderrAreSameFile() {
		log.SetOutput(progress.Bypass())
	}
	progress.Start()
	addTorrents(client, torrentPath)
	if client.WaitAll() {
		outputStats(client)
		log.Info("downloaded ALL the torrents")
		time.Sleep(60 * time.Second)
		log.Infof("seeded %d s ALL the torrents", 60)
		outputStats(client)
		status.Code = defaults.Success
		status.IsFinished = true
	} else {
		log.Warn("y u no complete torrents?!")
	}

}

func outputStats(cl *torrent.Client) {
	expvar.Do(func(kv expvar.KeyValue) {
		log.Infof("%s: %s\n", kv.Key, kv.Value)
	})
	cl.WriteStatus(os.Stdout)
}
