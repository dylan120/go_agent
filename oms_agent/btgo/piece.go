package btgo

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"hash/crc32"
	"io"
	"net"
)

type PieceBlock struct {
	Index    int
	SHA1Hash []byte
}

func ParsePieces(pieces []byte, pieceLength int64) (p []PieceBlock, err error) {
	buf := bytes.NewBuffer(pieces)
	i := 0
	for buf, reader := make([]byte, pieceLength), bufio.NewReader(buf); ; {
		_, err = reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		p = append(p, PieceBlock{Index: i, SHA1Hash: buf})
		i += 1
	}
	return
}

type peerPriority = uint32

//BEP0040

func ipv4Mask(cIP, pIP net.IP) (m net.IPMask, err error) {
	m = net.IPv4Mask(0xff, 0xff, 0x55, 0x55)
	if !cIP.Mask(m).Equal(pIP.Mask(m)) {
		return
	}
	m = net.IPv4Mask(0xff, 0xff, 0xff, 0x55)
	if !cIP.Mask(m).Equal(pIP.Mask(m)) {
		return
	}
	m = net.IPv4Mask(0xff, 0xff, 0xff, 0xff)
	return
}

func GetPriorityBytes(clientIP string, clientPort int, PeerIP string, peerPort int) (d []byte, err error) {
	cIP := net.ParseIP(clientIP)
	pIP := net.ParseIP(PeerIP)

	if cIP.Equal(pIP) {
		d = make([]byte, 4)
		binary.LittleEndian.PutUint16(d[:2], uint16(clientPort))
		binary.LittleEndian.PutUint16(d[2:], uint16(peerPort))
		return d, nil
	}

	if cip, pip := cIP.To4(), pIP.To4(); cip != nil && pip != nil {
		m, err := ipv4Mask(cIP, pIP)
		if err != nil {
			return d, err
		}
		a := cip.Mask(m)
		b := pip.Mask(m)

		if bytes.Compare(a, b) > 0 {
			d = append(b, a...)
		} else {
			d = append(a, b...)
		}
		return d, nil
	}
	return nil, errors.New("incomparable IPs")
}

func GetPriority(clientIP string, clientPort int, PeerIP string, PeerPort int) (prio peerPriority, err error) {
	crc32q := crc32.MakeTable(crc32.Castagnoli)
	d, err := GetPriorityBytes(clientIP, clientPort, PeerIP, PeerPort)
	if err != nil {
		return prio, errors.New("failed to get peer priority")
	}
	return crc32.Checksum(d, crc32q), nil

}
