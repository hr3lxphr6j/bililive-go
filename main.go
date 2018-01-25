package main

import (
	"net/url"
	"fmt"
	"os"
	"strings"
)

func main() {
	u, err := url.Parse("https://www.panda.tv/10300/")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(len(strings.Split(u.Path, "/")))
}
