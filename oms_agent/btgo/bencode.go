package btgo

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"reflect"
)

type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

type Encoder struct {
	w    io.Writer
	data []byte
}

func (e *Encoder) write(s string) {
	for s != "" {
		n := copy(e.data[:], s)
		s = s[n:]
		//e.write(e.data[:n])
		_, err := e.w.Write(s)
		if err != nil {
			log.Error(err)
			break
		}
	}

}

func (e *Encoder) encode(v reflect.Value) (err error) {
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			e.write("i1e")
		} else {
			e.write("i0e")
		}
	}
	return err
}

func (e *Encoder) decode() {

}

func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	e := Encoder{w: &buf}
	err := e.encode(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(d []byte) (err error) {
	return err
}
