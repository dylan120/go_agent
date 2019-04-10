package bencode

import (
	"../../utils"
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
		if v.Type().Key().Kind() != reflect.String { //nil value
			//continue
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
				mval := v.MapIndex(k)
				if isNilValue(mval) {
					continue
				}
				e.encode(k)
				e.encode(v.MapIndex(k))
			}
			e.write("e")
		}

	case reflect.Struct:
		e.write("d")
		s := make(structSlice, 0, v.NumField())
		s, err = encodeStruct(s, v)
		if !utils.CheckError(err) {
			sort.Sort(s)
			for _, val := range s {
				log.Info(isNilValue(val.value))
				if val.omit_empty && isNilValue(val.value) {
					continue
				}
				e.write(fmt.Sprintf("%d:%s", len(val.key), val.key))
				e.encode(val.value)
			}
		}

		e.write("e")

	case reflect.Interface:
		e.encode(v.Elem())
	}

	return err
}

type structDict struct {
	key        string
	value      reflect.Value
	omit_empty bool
}
type structSlice []structDict

func (s structSlice) Len() int           { return len(s) }
func (s structSlice) Less(i, j int) bool { return s[i].key < s[j].key }
func (s structSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func isNilValue(v reflect.Value) bool {
	return (v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr) &&
		v.IsNil()
}

func encodeStruct(s structSlice, v reflect.Value) (structSlice, error) {
	t := v.Type()
	var err error
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldValue := v.FieldByIndex(f.Index)
		if !fieldValue.CanInterface() { //unexported value
			continue
		}
		if isNilValue(fieldValue) { //nil value
			continue
		}
		//tagValue := f.Tag.Get("bencode")

		tag := getTag(f.Tag)
		if tag.Ignore() {
			continue
		}

		//if f.Type.Kind() == reflect.Struct {
		//	s, err = encodeStruct(s, fieldValue)
		//} else {
		//	s = append(s, structDict{tag.Key(), fieldValue})
		//	log.Info(s)
		//}
		s = append(s, structDict{key: tag.Key(), value: fieldValue, omit_empty: tag.OmitEmpty()})
	}
	return s, err
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
