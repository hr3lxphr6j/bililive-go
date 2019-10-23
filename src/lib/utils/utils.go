package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"net/url"
	"os/exec"
	"regexp"
)

func IsFFmpegExist() bool {
	_, ok := exec.LookPath("ffmpeg")
	return ok == nil
}

func GetMd5String(b []byte) string {
	md5Obj := md5.New()
	md5Obj.Write(b)
	return hex.EncodeToString(md5Obj.Sum(nil))
}

var (
	lowercaseRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	uppercaseRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	lettersRunes   = append(lowercaseRunes, uppercaseRunes...)
	digitsRunes    = []rune("0123456789")
	allRunes       = append(lettersRunes, digitsRunes...)
)

func GenRandomName(n int) string {
	b := make([]rune, n)
	b[0] = lowercaseRunes[rand.Intn(len(lowercaseRunes))]
	for i := 1; i < n; i++ {
		b[i] = allRunes[rand.Intn(len(allRunes))]
	}
	return string(b)
}

func Match1(re, str string) string {
	reg, err := regexp.Compile(re)
	if err != nil {
		return ""
	}
	match := reg.FindStringSubmatch(str)
	if match == nil || len(match) < 2 {
		return ""
	}
	return match[1]
}

func GenUrls(strs ...string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(strs))
	for _, str := range strs {
		u, err := url.Parse(str)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
