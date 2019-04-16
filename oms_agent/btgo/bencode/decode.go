package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"reflect"
	"strconv"
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
	//d.r.re
	return
}

func (d *Decoder) readNBytes(n int) (r []byte, err error) {
	var b byte
	for i := 0; i < n; i++ {
		b, err = d.r.ReadByte()
		if err != nil {
			break
		}
		r = append(r, b)
	}
	return
}

func (d *Decoder) decodeList(v reflect.Value) (err error) {
	return
}

func (d *Decoder) decodeInt(v reflect.Value) (err error) {
	return
}

func (d *Decoder) decodeString(v reflect.Value) (err error) {
	data, err := d.r.ReadBytes(':')
	if err != nil {
		return err
	}
	data = data[0 : len(data)-1]
	length, err := strconv.ParseInt(string(data), 10, 0)
	d.readNBytes(int(length))
	data = data[1:]
	return
}

func (d *Decoder) decodeDict(v reflect.Value) (err error) {
	//switch v.Kind(){
	//case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	//	err = d.decodeInt(v)
	//case reflect.String:
	//	err = d.decodeString(v)
	//case reflect.Slice,reflect.Array:
	//	err = d.decodeList(v)
	//case reflect.Map:
	//	err = d.decodeDict(v)
	//case reflect.Struct:
	//	for i := 0; i < v.NumField(); i++ {
	//		field := v.Field(i)
	//		err = d.decode(reflect.ValueOf(field))
	//	}
	//}
	d.r.ReadByte()
	for {
		v.F
		d.decodeString(v)
	}
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
