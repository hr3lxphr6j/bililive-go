package login

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/tidwall/gjson"
)

var (
// authapi = "http://passport.bilibili.com/x/passport-tv-login/qrcode/auth_code"
// api     = "http://passport.bilibili.com/x/passport-tv-login/qrcode/poll"
)

type Bilibili struct {
	LoginClient
	apiAuthUrl string
	apiPollUrl string
}

func NewBilibili(apiAuthUrl, apiPollUrl string, client *http.Client) *Bilibili {
	return &Bilibili{
		apiAuthUrl: apiAuthUrl,
		apiPollUrl: apiPollUrl,
		LoginClient: LoginClient{
			client: client,
		},
	}
}

func (b *Bilibili) getQRcodeUrlAuthCode() (string, string) {
	data := make(map[string]string)
	data["local_id"] = "0"
	data["ts"] = fmt.Sprintf("%d", time.Now().Unix())

	signature(&data)

	signData := strings.NewReader(mapToString(data))
	req, _ := http.NewRequest("POST", b.apiAuthUrl, signData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := b.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	code := gjson.Parse(string(body)).Get("code").Int()
	if code == 0 {
		qrcodeUrl := gjson.Parse(string(body)).Get("data.url").String()
		authCode := gjson.Parse(string(body)).Get("data.auth_code").String()
		return qrcodeUrl, authCode
	} else {
		panic("get_tv_qrcode_url_and_auth_code error")
	}
}

func (b *Bilibili) verifyLogin(authCode string) {
	data := make(map[string]string)
	data["auth_code"] = authCode
	data["local_id"] = "0"
	data["ts"] = fmt.Sprintf("%d", time.Now().Unix())
	signature(&data)

	dataString := strings.NewReader(mapToString(data))
	req, _ := http.NewRequest("POST", b.apiPollUrl, dataString)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for {
		resp, err := b.client.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ := io.ReadAll(resp.Body)
		code := gjson.Parse(string(body)).Get("code").Int()
		if code == 0 {
			fmt.Println("登录成功")
			filename := "cookie.json"
			err := os.WriteFile(filename, []byte(string(body)), 0644)
			if err != nil {
				panic(err)
			}
			fmt.Println("cookie 已保存在", filename)
			break
		} else {
			time.Sleep(time.Second * 3)
		}
		resp.Body.Close()
	}
}

func (b *Bilibili) Login() {
	fmt.Println("请最大化窗口，以确保二维码完整显示，回车继续")
	fmt.Scanf("%s", "")
	loginUrl, authCode := b.getQRcodeUrlAuthCode()
	qrcode := qrcodeTerminal.New()
	qrcode.Get([]byte(loginUrl)).Print()
	fmt.Println("或将此链接复制到手机B站打开:", loginUrl)
	b.verifyLogin(authCode)
}
