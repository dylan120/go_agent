package btgo

type Tracker struct {
	InfoHash   string
	PeerID     string
	IP         string
	Port       int64
	Uploaded   string
	Downloaded string
	left       string
	Event      string
}
