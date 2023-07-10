package login

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
)

var appkey = "4409e2ce8ffd12b8"
var appsec = "59b43e04ad6965f34319062b478f83dd"

func signature(params *map[string]string) {
	var keys []string
	(*params)["appkey"] = appkey
	for k := range *params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var query string
	for _, k := range keys {
		query += k + "=" + url.QueryEscape((*params)[k]) + "&"
	}
	query = query[:len(query)-1] + appsec
	hash := md5.New()
	hash.Write([]byte(query))
	(*params)["sign"] = hex.EncodeToString(hash.Sum(nil))
}

func mapToString(params map[string]string) string {
	var query string
	for k, v := range params {
		query += k + "=" + v + "&"
	}
	query = query[:len(query)-1]
	return query
}
