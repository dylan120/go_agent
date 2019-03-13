package utils

import (
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

func MakeTorrent(f *os.File, btAnnouce []string, srcFile string) (string, error) {
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
			//err = mi.Write(f)
			//CheckError(err)
			info, err := mi.UnmarshalInfo()
			if !CheckError(err) {
				return mi.Magnet(info.Name, mi.HashInfoBytes()).String(), nil
			}
		}
	}
	return "", err
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

func addTorrents(client *torrent.Client, magnetStream string) {
	t := func() *torrent.Torrent {
		t, err := client.AddMagnet(magnetStream)
		if err != nil {
			log.Fatalf("error adding magnet: %s", err)
		}
		return t
	}()
	torrentBar(t)
	go func() {
		<-t.GotInfo()
		t.DownloadAll()
	}()
}

func Download(srcMaster []string, mtgt []string,
	srcFile string, magnetStream string, md5 string, fileTargetPath string) {
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.Debug = true
	clientConfig.Seed = true
	clientConfig.DataDir = "/tmp"
	//clientConfig.PublicIp4 = flags.PublicIP
	//clientConfig.PublicIp6 = flags.PublicIP
	//if flags.PackedBlocklist != "" {
	//	blocklist, err := iplist.MMapPackedFile(flags.PackedBlocklist)
	//	if err != nil {
	//		log.Fatalf("error loading blocklist: %s", err)
	//	}
	//	defer blocklist.Close()
	//	clientConfig.IPBlocklist = blocklist
	//}
	//if flags.Mmap {
	//	clientConfig.DefaultStorage = storage.NewMMap("")
	//}
	//clientConfig.SetListenAddr(net.TCPAddr{IP:"0.0.0.0",Port:6881})
	//if flags.UploadRate != -1 {
	//	clientConfig.UploadRateLimiter = rate.NewLimiter(rate.Limit(flags.UploadRate), 256<<10)
	//}
	//if flags.DownloadRate != -1 {
	//	clientConfig.DownloadRateLimiter = rate.NewLimiter(rate.Limit(flags.DownloadRate), 1<<20)
	//}

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
	addTorrents(client, magnetStream)
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
