# req

<p align="center">
    <p align="center"><img src="https://req.cool/images/req.png"></p>
    <p align="center"><strong>Simple Go HTTP client with Black Magic</strong></p>
    <p align="center">
        <a href="https://github.com/imroc/req/actions/workflows/ci.yml?query=branch%3Amaster"><img src="https://github.com/imroc/req/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
        <a href="https://goreportcard.com/report/github.com/imroc/req/v3"><img src="https://goreportcard.com/badge/github.com/imroc/req/v3" alt="Go Report Card"></a>
        <a href="https://pkg.go.dev/github.com/imroc/req/v3"><img src="https://pkg.go.dev/badge/github.com/imroc/req/v3.svg"></a>
        <a href="LICENSE"><img src="https://img.shields.io/github/license/imroc/req.svg" alt="License"></a>
        <a href="https://github.com/imroc/req/releases"><img src="https://img.shields.io/github/v/release/imroc/req?display_name=tag&sort=semver" alt="GitHub Releases"></a>
        <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go"></a>
    </p> 
</p>

## Documentation

Full documentation is available on the official website: https://req.cool.

## <a name="Features">Features</a>

* **Simple and Powerful**: Simple and easy to use, providing rich client-level and request-level settings, all of which are intuitive and chainable methods.
* **Easy Debugging**: Powerful and convenient debug utilities, including debug logs, performance traces, and even dump the complete request and response content (see [Debugging](https://req.cool/docs/tutorial/debugging/)).
* **Easy API Testing**: API testing can be done with minimal code, no need to explicitly create any Request or Client, or even to handle errors (See [Quick HTTP Test](https://req.cool/docs/tutorial/quick-test/))
* **Smart by Default**: Detect and decode to utf-8 automatically if possible to avoid garbled characters (See [Auto Decode](https://req.cool/docs/tutorial/auto-decode/)), marshal request body and unmarshal response body automatically according to the Content-Type.
* **Support Multiple HTTP Versions**: Support `HTTP/1.1`, `HTTP/2`, and `HTTP/3`, and can automatically detect the server side and select the optimal HTTP version for requests, you can also force the protocol if you want (See [Force HTTP version](https://req.cool/docs/tutorial/force-http-version/)).
* **Support Retry**: Support automatic request retry and is fully customizable (See [Retry](https://req.cool/docs/tutorial/retry/)).
* **TLS Fingerprinting**: Support tls fingerprinting resistance, so that we can access websites that prohibit crawler programs by identifying TLS handshake fingerprints (See [TLS Fingerprinting](https://req.cool/docs/tutorial/tls-fingerprinting/)).
* **Multiple Authentication Methods**: You can use HTTP Basic Auth, Bearer Auth Token and Digest Auth out of box (see [Authentication](https://req.cool/docs/tutorial/authentication/)).
* **Easy Download and Upload**: You can download and upload files with simple request settings, and even set a callback to show real-time progress (See [Download](https://req.cool/docs/tutorial/download/) and [Upload](https://req.cool/docs/tutorial/upload/)).
* **Exportable**: `req.Transport` is exportable. Compared with `http.Transport`, it also supports HTTP3, dump content, middleware, etc. It can directly replace the Transport of `http.Client` in existing projects, and obtain more powerful functions with minimal code change.
* **Extensible**: Support Middleware for Request, Response, Client and Transport (See [Request and Response Middleware](https://req.cool/docs/tutorial/middleware-for-request-and-response/)) and [Client and Transport Middleware](https://req.cool/docs/tutorial/middleware-for-client-and-transport/)).

## <a name="Get-Started">Get Started</a>

**Install**

You first need [Go](https://go.dev/) installed (version 1.19+ is required), then you can use the below Go command to install req:

``` sh
go get github.com/imroc/req/v3
```

**Import**

Import req to your code:

```go
import "github.com/imroc/req/v3"
```

**Basic Usage**

```bash
# assume the following codes in main.go file
$ cat main.go
```

```go
package main

import (
    "github.com/imroc/req/v3"
)

func main() {
    req.DevMode() // Treat the package name as a Client, enable development mode
    req.MustGet("https://httpbin.org/uuid") // Treat the package name as a Request, send GET request.

    req.EnableForceHTTP1() // Force using HTTP/1.1
    req.MustGet("https://httpbin.org/uuid")
}
```

```bash
$ go run main.go
2022/05/19 10:05:07.920113 DEBUG [req] HTTP/2 GET https://httpbin.org/uuid
:authority: httpbin.org
:method: GET
:path: /uuid
:scheme: https
user-agent: req/v3 (https://github.com/imroc/req/v3)
accept-encoding: gzip

:status: 200
date: Thu, 19 May 2022 02:05:08 GMT
content-type: application/json
content-length: 53
server: gunicorn/19.9.0
access-control-allow-origin: *
access-control-allow-credentials: true

{
  "uuid": "bd519208-35d1-4483-ad9f-e1555ae108ba"
}

2022/05/19 10:05:09.340974 DEBUG [req] HTTP/1.1 GET https://httpbin.org/uuid
GET /uuid HTTP/1.1
Host: httpbin.org
User-Agent: req/v3 (https://github.com/imroc/req/v3)
Accept-Encoding: gzip

HTTP/1.1 200 OK
Date: Thu, 19 May 2022 02:05:09 GMT
Content-Type: application/json
Content-Length: 53
Connection: keep-alive
Server: gunicorn/19.9.0
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true

{
  "uuid": "49b7f916-c6f3-49d4-a6d4-22ae93b71969"
}
```

The sample code above is good for quick testing purposes, which use `DevMode()` to see request details, and send requests using global wrapper methods that use the default client behind the scenes to initiate the request.

In production, it is recommended to explicitly create a client, and then use the same client to send all requests, please see other examples below.

**Videos**

The following is a series of video tutorials for req:

* [Youtube Play List](https://www.youtube.com/watch?v=Dy8iph8JWw0&list=PLnW6i9cc0XqlhUgOJJp5Yf1FHXlANYMhF&index=2)
* [BiliBili 播放列表](https://www.bilibili.com/video/BV14t4y1J7cm) (Chinese)

**More**

Check more introduction, tutorials, examples, best practices and API references on the [official website](https://req.cool/).

## <a name="Simple-Get">Simple GET</a>

```go
package main

import (
	"fmt"
	"github.com/imroc/req/v3"
	"log"
)

func main() {
	client := req.C() // Use C() to create a client.
	resp, err := client.R(). // Use R() to create a request.
		Get("https://httpbin.org/uuid")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}
```

```txt
{
  "uuid": "a4d4430d-0e5f-412f-88f5-722d84bc2a62"
}
```

## <a name="Advanced-Get">Advanced GET</a>

```go
package main

import (
  "fmt"
  "github.com/imroc/req/v3"
  "log"
  "time"
)

type ErrorMessage struct {
  Message string `json:"message"`
}

type UserInfo struct {
  Name string `json:"name"`
  Blog string `json:"blog"`
}

func main() {
  client := req.C().
    SetUserAgent("my-custom-client"). // Chainable client settings.
    SetTimeout(5 * time.Second)

  var userInfo UserInfo
  var errMsg ErrorMessage
  resp, err := client.R().
    SetHeader("Accept", "application/vnd.github.v3+json"). // Chainable request settings.
    SetPathParam("username", "imroc"). // Replace path variable in url.
    SetSuccessResult(&userInfo). // Unmarshal response body into userInfo automatically if status code is between 200 and 299.
    SetErrorResult(&errMsg). // Unmarshal response body into errMsg automatically if status code >= 400.
    EnableDump(). // Enable dump at request level, only print dump content if there is an error or some unknown situation occurs to help troubleshoot.
    Get("https://api.github.com/users/{username}")

  if err != nil { // Error handling.
    log.Println("error:", err)
    log.Println("raw content:")
    log.Println(resp.Dump()) // Record raw content when error occurs.
    return
  }

  if resp.IsErrorState() { // Status code >= 400.
    fmt.Println(errMsg.Message) // Record error message returned.
    return
  }

  if resp.IsSuccessState() { // Status code is between 200 and 299.
    fmt.Printf("%s (%s)\n", userInfo.Name, userInfo.Blog)
    return
  }

  // Unknown status code.
  log.Println("unknown status", resp.Status)
  log.Println("raw content:")
  log.Println(resp.Dump()) // Record raw content when server returned unknown status code.
}
```

Normally it will output (SuccessState):

```txt
roc (https://imroc.cc)
```

## <a name="More-Advanced-Get">More Advanced GET</a>

You can set up a unified logic for error handling on the client, so that each time you send a request you only need to focus on the success situation, reducing duplicate code.

```go
package main

import (
	"fmt"
	"github.com/imroc/req/v3"
	"log"
	"time"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func (msg *ErrorMessage) Error() string {
	return fmt.Sprintf("API Error: %s", msg.Message)
}

type UserInfo struct {
	Name string `json:"name"`
	Blog string `json:"blog"`
}

var client = req.C().
	SetUserAgent("my-custom-client"). // Chainable client settings.
	SetTimeout(5 * time.Second).
	EnableDumpEachRequest().
	SetCommonErrorResult(&ErrorMessage{}).
	OnAfterResponse(func(client *req.Client, resp *req.Response) error {
		if resp.Err != nil { // There is an underlying error, e.g. network error or unmarshal error.
			return nil
		}
		if errMsg, ok := resp.ErrorResult().(*ErrorMessage); ok {
			resp.Err = errMsg // Convert api error into go error
			return nil
		}
		if !resp.IsSuccessState() {
			// Neither a success response nor a error response, record details to help troubleshooting
			resp.Err = fmt.Errorf("bad status: %s\nraw content:\n%s", resp.Status, resp.Dump())
		}
		return nil
	})

func main() {
	var userInfo UserInfo
	resp, err := client.R().
		SetHeader("Accept", "application/vnd.github.v3+json"). // Chainable request settings
		SetPathParam("username", "imroc").
		SetSuccessResult(&userInfo). // Unmarshal response body into userInfo automatically if status code is between 200 and 299.
		Get("https://api.github.com/users/{username}")

	if err != nil { // Error handling.
		log.Println("error:", err)
		return
	}

	if resp.IsSuccessState() { // Status code is between 200 and 299.
		fmt.Printf("%s (%s)\n", userInfo.Name, userInfo.Blog)
	}
}
```

## <a name="Simple-Post">Simple POST</a>

```go
package main

import (
  "fmt"
  "github.com/imroc/req/v3"
  "log"
)

type Repo struct {
  Name string `json:"name"`
  Url  string `json:"url"`
}

type Result struct {
  Data string `json:"data"`
}

func main() {
  client := req.C().DevMode()
  var result Result

  resp, err := client.R().
    SetBody(&Repo{Name: "req", Url: "https://github.com/imroc/req"}).
    SetSuccessResult(&result).
    Post("https://httpbin.org/post")
  if err != nil {
    log.Fatal(err)
  }

  if !resp.IsSuccessState() {
    fmt.Println("bad response status:", resp.Status)
    return
  }
  fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++")
  fmt.Println("data:", result.Data)
  fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++")
}
```

```txt
2022/05/19 20:11:00.151171 DEBUG [req] HTTP/2 POST https://httpbin.org/post
:authority: httpbin.org
:method: POST
:path: /post
:scheme: https
user-agent: req/v3 (https://github.com/imroc/req/v3)
content-type: application/json; charset=utf-8
content-length: 55
accept-encoding: gzip

{"name":"req","website":"https://github.com/imroc/req"}

:status: 200
date: Thu, 19 May 2022 12:11:00 GMT
content-type: application/json
content-length: 651
server: gunicorn/19.9.0
access-control-allow-origin: *
access-control-allow-credentials: true

{
  "args": {},
  "data": "{\"name\":\"req\",\"website\":\"https://github.com/imroc/req\"}",
  "files": {},
  "form": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Content-Length": "55",
    "Content-Type": "application/json; charset=utf-8",
    "Host": "httpbin.org",
    "User-Agent": "req/v3 (https://github.com/imroc/req/v3)",
    "X-Amzn-Trace-Id": "Root=1-628633d4-7559d633152b4307288ead2e"
  },
  "json": {
    "name": "req",
    "website": "https://github.com/imroc/req"
  },
  "origin": "103.7.29.30",
  "url": "https://httpbin.org/post"
}

++++++++++++++++++++++++++++++++++++++++++++++++
data: {"name":"req","url":"https://github.com/imroc/req"}
++++++++++++++++++++++++++++++++++++++++++++++++
```

## <a name="Do-API-Style">Do API Style</a>

If you like, you can also use a Do API style like the following to make requests:

```go
package main

import (
	"fmt"
	"github.com/imroc/req/v3"
)

type APIResponse struct {
	Origin string `json:"origin"`
	Url    string `json:"url"`
}

func main() {
	var resp APIResponse
	c := req.C().SetBaseURL("https://httpbin.org/post")
	err := c.Post().
		SetBody("hello").
		Do().
		Into(&resp)
	if err != nil {
		panic(err)
	}
	fmt.Println("My IP is", resp.Origin)
}
```

```txt
My IP is 182.138.155.113
```

* The order of chain calls is more intuitive: first call Client to create a request with a specified Method, then use chain calls to set the request, then use `Do()` to fire the request, return Response, and finally call `Response.Into` to unmarshal response body into specified object.
* `Response.Into` will return an error if an error occurs during sending the request or during unmarshalling.
* The url of some APIs is fixed, and different types of requests are implemented by passing different bodies. In this scenario, `Client.SetBaseURL` can be used to set a unified url, and there is no need to set the url for each request when initiating a request. Of course, you can also call `Request.SetURL` to set it if you need it.

## <a name="Build-SDK-With-Req">Build SDK With Req</a>

Here is an example of building GitHub's SDK with req, using two styles (`GetUserProfile_Style1`, `GetUserProfile_Style2`).

```go
import (
	"context"
	"fmt"
	"github.com/imroc/req/v3"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

// Error implements go error interface.
func (msg *ErrorMessage) Error() string {
	return fmt.Sprintf("API Error: %s", msg.Message)
}

type GithubClient struct {
	*req.Client
}

func NewGithubClient() *GithubClient {
	return &GithubClient{
		Client: req.C().
			SetBaseURL("https://api.github.com").
			SetCommonErrorResult(&ErrorMessage{}).
			EnableDumpEachRequest().
			OnAfterResponse(func(client *req.Client, resp *req.Response) error {
				if resp.Err != nil { // There is an underlying error, e.g. network error or unmarshal error.
					return nil
				}
				if errMsg, ok := resp.ErrorResult().(*ErrorMessage); ok {
					resp.Err = errMsg // Convert api error into go error
					return nil
				}
				if !resp.IsSuccessState() {
					// Neither a success response nor a error response, record details to help troubleshooting
					resp.Err = fmt.Errorf("bad status: %s\nraw content:\n%s", resp.Status, resp.Dump())
				}
				return nil
			}),
	}
}

type UserProfile struct {
	Name string `json:"name"`
	Blog string `json:"blog"`
}

// GetUserProfile_Style1 returns the user profile for the specified user.
// Github API doc: https://docs.github.com/en/rest/users/users#get-a-user
func (c *GithubClient) GetUserProfile_Style1(ctx context.Context, username string) (user *UserProfile, err error) {
	_, err = c.R().
		SetContext(ctx).
		SetPathParam("username", username).
		SetSuccessResult(&user).
		Get("/users/{username}")
	return
}

// GetUserProfile_Style2 returns the user profile for the specified user.
// Github API doc: https://docs.github.com/en/rest/users/users#get-a-user
func (c *GithubClient) GetUserProfile_Style2(ctx context.Context, username string) (user *UserProfile, err error) {
	err = c.Get("/users/{username}").
		SetPathParam("username", username).
		Do(ctx).
		Into(&user)
	return
}
```

## Contributing

If you have a bug report or feature request, you can [open an issue](https://github.com/imroc/req/issues/new), and [pull requests](https://github.com/imroc/req/pulls) are also welcome.

## Contact

If you have questions, feel free to reach out to us in the following ways:

* [Github Discussion](https://github.com/imroc/req/discussions)
* [Slack](https://imroc-req.slack.com/archives/C03UFPGSNC8) | [Join](https://slack.req.cool/)
* QQ Group (Chinese): 621411351 - <a href="https://qm.qq.com/cgi-bin/qm/qr?k=P8vOMuNytG-hhtPlgijwW6orJV765OAO&jump_from=webapi"><img src="https://pub.idqqimg.com/wpa/images/group.png"></a>

## Sponsors

If you like req and it really helps you, feel free to reward me with a cup of coffee, and don't forget to mention your github id.

<table>
    <tr>
        <td align="center">
            <img src="https://req.cool/images/wechat.jpg" width="200px"   alt=""/>
            <br />
            <sub><b>Wechat</b></sub>
        </td>
        <td align="center">
            <img src="https://req.cool/images/alipay.jpg" width="200px"   alt=""/>
            <br />
            <sub><b>Alipay</b></sub>
        </td>
    </tr>
</table>

Many thanks to the following sponsors:

<table>
    <tr>
        <td align="center">
            <a href="https://github.com/M-Cosmosss">
                <img src="https://avatars.githubusercontent.com/u/46757262?v=4?s=100" width="160px"   alt=""/>
                <br />
                <sub><b>M-Cosmosss 🥇</b></sub>
            </a>
        </td>
        <td align="center">
            <a href="https://github.com/aadog">
                <img src="https://avatars.githubusercontent.com/u/18098725?v=4?s=100" width="160px"   alt=""/>
                <br />
                <sub><b>aadog 🥈</b></sub>
            </a>
        </td>
    </tr>
</table>

## <a name="License">License</a>

`Req` released under MIT license, refer [LICENSE](LICENSE) file.
