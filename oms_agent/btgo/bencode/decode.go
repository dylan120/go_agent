package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
)

type Decoder struct {
	r *bufio.Reader
	//raw   []byte
	index int
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

func Unmarshal(data []byte, v interface{}) (err error) {
	buf := bytes.NewBuffer(data)
	d := NewDecoder(buf)
	err = d.decode(reflect.ValueOf(v))
	return
}

func (d *Decoder) readPeek(delim string) (b byte, err error) {
	d.r.re
	return
}

func (d *Decoder) decodeDict(v reflect.Value) (err error) {
	return
}

func (d *Decoder) decodeList(v reflect.Value) (err error) {
	return
}

func (d *Decoder) decodeString(v reflect.Value) (err error) {
	return
}

func (d *Decoder) decodeInt(v reflect.Value) (err error) {
	d.
	return
}

func (d *Decoder) decode(v reflect.Value) (err error) {
	ch, err := d.r.Peek(1)
	if err != nil {
		return
	}

	switch ch[0] {
	case 'i':
		err = d.decodeInt(v)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		err = d.decodeString(v)
	case 'l':
		err = d.decodeList(v)
	case 'd':
		err = d.decodeDict(v)
	default:
		err = errors.New("Invalid input")
	}

	return
}
