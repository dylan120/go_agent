package bencode

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"reflect"
	"sort"
)

type sortValues []reflect.Value

func (p sortValues) Len() int           { return len(p) }
func (p sortValues) Less(i, j int) bool { return p[i].String() < p[j].String() }
func (p sortValues) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

type Encoder struct {
	w    io.Writer
	data [64]byte
}

func A() {

}
func (e *Encoder) write(s string) {
	for {
		if s == "" {
			break
		}
		n := copy(e.data[:], s)
		_, err := e.w.Write(e.data[:n])
		if err != nil {
			log.Error(err)
			break
		}
		s = s[n:]
	}

}

func (e *Encoder) encode(v reflect.Value) (err error) {
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		e.write(fmt.Sprintf("%d:%s", len(s), s))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.write(fmt.Sprintf("i%de", v.Int()))
	case reflect.Bool:
		if v.Bool() {
			e.write("i1e")
		} else {
			e.write("i0e")
		}
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			s := v.Bytes()
			e.write(fmt.Sprintf("%d:%s", len(s), s))
			break
		}
		if v.IsNil() {
			e.write("le")
			break
		}
		e.write("l")
		for i := 0; i < v.Len(); i++ {
			e.encode(v.Index(i))
		}
		e.write("e")
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			log.Error("dict keys must be string")
		} else {
			if v.IsNil() {
				e.write("de")
				break
			}
			e.write("d")
			var (
				keys sortValues = v.MapKeys()
			)
			sort.Sort(keys)
			for _, k := range keys {
				e.encode(k)
				e.encode(v.MapIndex(k))
			}
			e.write("e")
		}

	//case reflect.Struct:

	case reflect.Interface:
		e.encode(v.Elem())
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
