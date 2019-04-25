package btgo

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	AnnounceList [][]string
	MetaInfo     MetaInfo
}

func NewClient(torrentFile string, savePath string) (c *Client, err error) {
	t, err := NewTorrentFromFile(torrentFile)
	if err != nil {
		return
	}
	c.MetaInfo = t.MetaInfo
	c.AnnounceList = t.MetaInfo.AnnounceList
	return
}

func (cli *Client) RequestTracker() (err error) {
	for _, announces := range cli.AnnounceList {
		for _, announce := range announces {
			resp, err := http.Get(announce)
			if err != nil {
				return
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}
			fmt.Println(body)
			return
		}
	}

}

func (cli *Client) Run() (err error) {
	cli.RequestTracker()
	return
}
