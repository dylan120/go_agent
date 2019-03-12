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
	"strings"
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
	if !CheckError(err) {
		mi.InfoBytes, err = bencode.Marshal(info)
		if !CheckError(err) {
			err = mi.Write(f)
			CheckError(err)
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

func addTorrents(client *torrent.Client) {
	for _, arg := range flags.Torrent {
		t := func() *torrent.Torrent {
			if strings.HasPrefix(arg, "magnet:") {
				t, err := client.AddMagnet(arg)
				if err != nil {
					log.Fatalf("error adding magnet: %s", err)
				}
				return t
			} else if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
				response, err := http.Get(arg)
				if err != nil {
					log.Fatalf("Error downloading torrent file: %s", err)
				}

				metaInfo, err := metainfo.Load(response.Body)
				defer response.Body.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "error loading torrent file %q: %s\n", arg, err)
					os.Exit(1)
				}
				t, err := client.AddTorrent(metaInfo)
				if err != nil {
					log.Fatal(err)
				}
				return t
			} else if strings.HasPrefix(arg, "infohash:") {
				t, _ := client.AddTorrentInfoHash(metainfo.NewHashFromHex(strings.TrimPrefix(arg, "infohash:")))
				return t
			} else {
				metaInfo, err := metainfo.LoadFromFile(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error loading torrent file %q: %s\n", arg, err)
					os.Exit(1)
				}
				t, err := client.AddTorrent(metaInfo)
				if err != nil {
					log.Fatal(err)
				}
				return t
			}
		}()
		torrentBar(t)
		//t.AddPeers(func() (ret []torrent.Peer) {
		//	for _, ta := range flags.TestPeer {
		//		ret = append(ret, torrent.Peer{
		//			IP:   ta.IP,
		//			Port: ta.Port,
		//		})
		//	}
		//	return
		//}())
		go func() {
			<-t.GotInfo()
			t.DownloadAll()
		}()
	}
}

func Download(srcMaster []string, mtgt []string,
	srcFile string, torrentFile *os.File, md5 string, fileTargetPath string) {
	clientConfig := torrent.NewDefaultClientConfig()
	//clientConfig.Debug = flags.Debug
	clientConfig.Seed = true
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
	clientConfig.SetListenAddr("0.0.0.0")
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
	addTorrents(client)
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
