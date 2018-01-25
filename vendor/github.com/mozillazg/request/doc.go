// Package request is a developer-friendly HTTP request library for Gopher.
//
// GET Request:
//
// 	c := &http.Client{}
// 	req := request.NewRequest(c)
// 	resp, err := req.Get("http://httpbin.org/get")
// 	defer resp.Body.Close()  // **Don't forget close the response body**
// 	j, err := resp.Json()
//
// POST Request:
//
// 	req = request.NewRequest(c)
//	req.Data = map[string]string{
//		"key": "value",
//		"a":   "123",
//	}
//	resp, err := req.Post("http://httpbin.org/post")
//
// Custom Cookies:
//
// 	req = request.NewRequest(c)
//	req.Cookies = map[string]string{
//		"key": "value",
//		"a":   "123",
//	}
//	resp, err := req.Get("http://httpbin.org/cookies")
//
//
// Custom Headers:
//
// 	req = request.NewRequest(c)
//	req.Headers = map[string]string{
//		"Accept-Encoding": "gzip,deflate,sdch",
//		"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
//	}
//	resp, err := req.Get("http://httpbin.org/get")
//
// Upload Files:
//
// 	req = request.NewRequest(c)
//	f, err := os.Open("test.txt")
//	req.Files = []request.FileField{
//		request.FileField{"file", "test.txt", f},
//	}
//	resp, err := req.Post("http://httpbin.org/post")
//
// Json Body:
//
// 	req = request.NewRequest(c)
//	req.Json = map[string]string{
//		"a": "A",
//		"b": "B",
//	}
//	resp, err := req.Post("http://httpbin.org/post")
//	req.Json = []int{1, 2, 3}
//	resp, err = req.Post("http://httpbin.org/post")
//
// others body:
//
// 	req = request.NewRequest(c)
//	// not set Content-Type
//	req.Body = strings.NewReader("<xml><a>abc</a></xml")
//	resp, err := req.Post("http://httpbin.org/post")
//
//	// form
// 	req = request.NewRequest(c)
//	req.Body = strings.NewReader("a=1&b=2")
//	req.Headers = map[string]string{
//		"Content-Type": request.DefaultContentType,
//	}
//	resp, err = req.Post("http://httpbin.org/post")
//
// Proxy:
//
// 	req = request.NewRequest(c)
//	req.Proxy = "http://127.0.0.1:8080"
//	// req.Proxy = "https://127.0.0.1:8080"
//	// req.Proxy = "socks5://127.0.0.1:57341"
//	resp, err := req.Get("http://httpbin.org/get")
//
// HTTP Basic Authentication:
//
// 	req = request.NewRequest(c)
//	req.BasicAuth = request.BasicAuth{"user", "passwd"}
//	resp, err := req.Get("http://httpbin.org/basic-auth/user/passwd")
//
// Need more control?
//
// You can setup req.Client(you know, it's an &http.Client),
// for example: set timeout
//
//	timeout := time.Duration(1 * time.Second)
//	req.Client.Timeout = timeout
//	req.Get("http://httpbin.org/get")
package request
