package bencode

type Decoder struct {
}

func Unmarshal(d []byte) (err error) {
	e := Decoder{}
	err = e.decode(d)
	return err
}

func (e *Decoder) decode(d []byte) (err error) {

	return err
}
