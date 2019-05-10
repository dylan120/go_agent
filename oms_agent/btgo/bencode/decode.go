package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type Decoder struct {
	r     *bufio.Reader
	cache []byte
	off   int
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

func (d *Decoder) readCache(off int) (r []byte) {
	r = d.cache[d.off:]
	d.off = len(d.cache)
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

func (d *Decoder) decodeInt(v reflect.Value) (err error) {
	data, err := d.r.ReadBytes('e')
	if err != nil {
		return err
	}
	s, err := strconv.ParseInt(string(data[1:len(data)-1]), 10, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}
	v.SetInt(s)

	return
}

func (d *Decoder) decodeString(v reflect.Value) (err error) {
	data, err := d.r.ReadBytes(':')
	if err != nil {
		fmt.Println(err)
		return err
	}
	length, err := strconv.ParseInt(string(data[0:len(data)-1]), 10, 0)
	s, err := d.readNBytes(int(length))
	if err != nil {
		fmt.Println(">>>>", err)
		return err
	}

	//d.cache = append(d.cache, s...)
	switch v.Kind() {
	case reflect.String:
		//v.SetString(string(d.readCache(d.off)))
		v.SetString(string(s))
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes(s)
		}
	}

	return
}

func (d *Decoder) decodeDict(v reflect.Value) (err error) {
	return
}

func (d *Decoder) decodeList(v reflect.Value) (err error) {
	b, err := d.r.ReadByte()
	if err != nil {
		if b == 'e' {
			_, err := d.r.ReadByte()
			return err
		}
	}

	for i := 0; ; i++ {
		ch, err := d.r.Peek(1)
		if err != nil {
			return err
		}
		if ch[0] == 'e' {
			d.r.ReadByte()
			break
		}
		if v.Kind() == reflect.Slice && i >= v.Len() {
			v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
		}

		err = d.decode(v.Index(i))
		if err != nil {
			return err
		}
	}
	return
}

func (d *Decoder) getStrcuctMap(m map[string]reflect.Value, v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		tfield := t.Field(i)
		tag := getTag(tfield.Tag)
		f := v.FieldByIndex(tfield.Index)
		key := tag.Key()
		if key == "" {
			key = tfield.Name
		}
		m[key] = f
	}
}

func (d *Decoder) decodeStruct(v reflect.Value) (err error) {
	b, err := d.r.ReadByte()
	if err != nil {
		if b == 'e' {
			_, err := d.r.ReadByte()
			return err
		}
	}
	m := make(map[string]reflect.Value)
	d.getStrcuctMap(m, v)
	for {
		ch, err := d.r.Peek(1)
		if err != nil {
			return err
		}

		if ch[0] == 'e' {
			d.r.ReadByte()
			break
		}

		data, err := d.r.ReadBytes(':')
		if err != nil {
			fmt.Println(err)
			return err
		}
		length, err := strconv.ParseInt(string(data[0:len(data)-1]), 10, 0)
		s, err := d.readNBytes(int(length))
		if err != nil {
			fmt.Println(">>>>", err)
			return err
		}

		mkey := string(s)

		val, ok := m[mkey] //TODO
		if ok {
			err = d.decode(val)
		} else {
			ch, err := d.r.Peek(1)
			if err != nil {
				return err
			}
			switch ch[0] {
			case 'i':
				var i int
				d.decodeInt(reflect.ValueOf(&i).Elem())
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				var s string
				d.decodeString(reflect.ValueOf(&s).Elem())
				//case 'l':
				//	var s []byte
				//	d.decodeList(reflect.ValueOf(&s))
				//case 'd':
				//	d.decodeStruct(reflect.New(reflect.TypeOf(reflect.Struct)))

			default:
				err = errors.New("Unsupport type")
			}
		}

	}
	return err
}

func (d *Decoder) indirect(v reflect.Value) reflect.Value {
	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}
		if v.Elem().Kind() != reflect.Ptr && v.CanSet() {
			v = v.Elem()
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func (d *Decoder) decode(v reflect.Value) (err error) {
	val := d.indirect(v)
	ch, err := d.r.Peek(1)
	if err != nil {
		return
	}
	switch ch[0] {
	case 'i':
		err = d.decodeInt(val)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		err = d.decodeString(val)
	case 'l':
		err = d.decodeList(val)
	case 'd':
		err = d.decodeStruct(val)
	default:
		err = errors.New("invalid input")
	}
	return
}
