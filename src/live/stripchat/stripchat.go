package main

import (
	"fmt"
	"regexp"

	"github.com/parnurzeal/gorequest"
	"github.com/tidwall/gjson"
)

func get_modelId(modleName string) string {

	// modleName := "S-wan"
	// 创建一个新的 Request 对象
	request := gorequest.New()

	// 添加头部信息
	request.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	request.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	request.Set("Accept-Encoding", "gzip, deflate")
	request.Set("Upgrade-Insecure-Requests", "1")
	request.Set("Sec-Fetch-Dest", "document")
	request.Set("Sec-Fetch-Mode", "navigate")
	request.Set("Sec-Fetch-Site", "none")
	request.Set("Sec-Fetch-User", "?1")
	request.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	// request.Set("If-Modified-Since", "Mon, 29 Jul 2024 08:41:12 GMT")
	request.Set("Te", "trailers")
	request.Set("Connection", "close")

	// 发起 GET 请求
	_, body, errs := request.Get("https://zh.stripchat.com/api/front/v2/models/username/" + modleName + "/chat").End()

	// 处理响应
	if len(errs) > 0 {
		fmt.Println("请求出错:", errs)
		return ""
	} else {
		// fmt.Println("成功获取modelId", resp.StatusCode)
		// 解析 JSON 响应
		modelId := gjson.Get(body, "messages.0.modelId").String()
		return modelId
	}
}

func get_M3u8(modelId string) string {
	url := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + "_auto.m3u8?playlistType=lowLatency"
	request := gorequest.New()
	_, body, errs := request.Get(url).End()

	if len(errs) > 0 {
		// fmt.Println("请求出错:", errs)
		return ""
	} else {
		// fmt.Println((body))
		return body
	}
}

func main() {
	m3u8_old := get_M3u8(get_modelId("Lucky-is-lucky"))
	// 定义正则表达式
	re := regexp.MustCompile(`(https:\/\/[\w\-\.]+\/hls\/[\d]+\/[\d]+\.m3u8\?playlistType=lowLatency)`)

	// 查找匹配的字符串
	matches := re.FindString(m3u8_old)

	// 输出匹配的结果

	fmt.Println(matches)

}
