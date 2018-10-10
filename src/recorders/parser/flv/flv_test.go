package flv

import (
	"fmt"
	"os"
	"testing"
)

func TestParseFlv(t *testing.T) {
	f, err := os.Open("/Users/chigusa/test.flv")
	if err != nil {
		panic(err)
	}
	parser := NewParser(f)
	fmt.Println(parser.ParseFlv())
}
