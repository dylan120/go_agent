package main

import (
	"../btgo"
	"../btgo/bencode"
	"fmt"
)

func BencodeTest() {
	s, _ := bencode.Marshal(map[string]interface{}{
		"a": 1, "d": "xxx", "c": true, "b": []interface{}{"a", 2}})
	fmt.Printf("%s", s)
}

func main() {
	BencodeTest()
	fmt.Println(btgo.NewTorrent("1", []string{"/tmp/test_1_1_1"}))
}
