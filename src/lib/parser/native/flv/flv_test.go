package flv

import (
	"fmt"
	"os"
	"testing"
)

func TestParseFlv(t *testing.T) {
	i, err := os.Open("/Users/chigusa/test.flv")
	o, err := os.OpenFile("/Users/chigusa/test2.flv", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	parser, _ := NewParser(i, o)
	fmt.Println(parser.ParseFlv())
}
