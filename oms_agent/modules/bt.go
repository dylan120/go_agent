package main

import (
	"../defaults"
	"../utils"
	"expvar"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/dustin/go-humanize"
	"github.com/gosuri/uiprogress"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var progress = uiprogress.New()

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
	return err
}

func stdoutAndStderrAreSameFile() bool {
	fi1, _ := os.Stdout.Stat()
	fi2, _ := os.Stderr.Stat()
	return os.SameFile(fi1, fi2)
}

func exitSignalHandlers(client *torrent.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for {
		log.Printf("close signal received: %+v", <-c)
		client.Close()
	}
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
			return fmt.Sprintf("downloading (%s/%s)", humanize.Bytes(uint64(t.BytesCompleted())), humanize.Bytes(uint64(t.Info().TotalLength())))
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
			log.Fatal(err)
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
	clientConfig.DataDir = fileTargetPath
	clientConfig.NoDHT = true

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}
	defer client.Close()
	go exitSignalHandlers(client)

	// Write status on the root path on the default HTTP muxer. This will be
	// bound to localhost somewhere if GOPPROF is set, thanks to the envpprof
	// import.
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		client.WriteStatus(w)
	})
	if stdoutAndStderrAreSameFile() {
		log.SetOutput(progress.Bypass())
	}
	progress.Start()
	addTorrents(client, torrentStream)
	if client.WaitAll() {
		log.Print("downloaded ALL the torrents")
	} else {
		log.Fatal("y u no complete torrents?!")
	}
	outputStats(client)
	select {}
	outputStats(client)
}

func Download(step utils.Step, procDir string, resultChannel chan string, status *defaults.Status) {
	clientConfig := torrent.NewDefaultClientConfig()
	torrentStream := step.FileParam[0]
	//md5 := step.FileParam[1]
	fileTargetPath := step.FileTargetPath
	clientConfig.Debug = true
	clientConfig.Seed = true
	clientConfig.DataDir = fileTargetPath
	clientConfig.NoDHT = true

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("error creating client: %s", err)
	}
	defer client.Close()
	go exitSignalHandlers(client)

	// Write status on the root path on the default HTTP muxer. This will be
	// bound to localhost somewhere if GOPPROF is set, thanks to the envpprof
	// import.
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		client.WriteStatus(w)
	})
	if stdoutAndStderrAreSameFile() {
		log.SetOutput(progress.Bypass())
	}
	progress.Start()
	addTorrents(client, torrentStream)
	if client.WaitAll() {
		log.Print("downloaded ALL the torrents")
	} else {
		log.Fatal("y u no complete torrents?!")
	}
	outputStats(client)
	select {}
	outputStats(client)
}

func outputStats(cl *torrent.Client) {

	expvar.Do(func(kv expvar.KeyValue) {
		fmt.Printf("%s: %s\n", kv.Key, kv.Value)
	})
	cl.WriteStatus(os.Stdout)
}
