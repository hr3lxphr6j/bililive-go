package douyu

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"

	"github.com/robertkrimen/otto"
	"github.com/satori/go.uuid"
	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
)

/*
	From https://github.com/zhangn1985/ykdl

	Thanks
*/
const (
	domain = "www.douyu.com"
	cnName = "斗鱼"

	liveInfoUrl = "https://open.douyucdn.cn/api/RoomApi/room"
	liveEncUrl  = "https://www.douyu.com/swf_api/homeH5Enc"
	liveAPIUrl  = "https://www.douyu.com/lapi/live/getH5Play"
)

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url),
	}, nil
}

var (
	cryptoJS []byte
	header   = map[string]string{
		"Referer":      "https://www.douyu.com",
		"content-type": "application/x-www-form-urlencoded",
	}
	douyuRoomIDRegs = []*regexp.Regexp{
		regexp.MustCompile(`\$ROOM\.room_id\s*=\s*(\d+)`),
		regexp.MustCompile(`room_id\s*=\s*(\d+)`),
		regexp.MustCompile(`"room_id.?":(\d+)`),
		regexp.MustCompile(`data-onlineid=(\d+)`),
	}
	workflowReg = regexp.MustCompile(`function ub98484234\(.+?\Weval\((\w+)\);`)
	jsDomTmpl   = template.Must(template.New("jsDom").Parse(`
		{{.DebugMessages}} = { {{.DecryptedCodes}}: []};
		if (!this.window) {window = {};}
		if (!this.document) {document = {};}
	`))
	jsPatchTmpl = template.Must(template.New("jsPatch").Parse(`
		{{.DebugMessages}}.{{.DecryptedCodes}}.push({{.Workflow}});
		var patchCode = function(workflow) {
			var testVari = /(\w+)=(\w+)\([\w\+]+\);.*?(\w+)="\w+";/.exec(workflow);
			if (testVari && testVari[1] == testVari[2]) {
				{{.Workflow}} += testVari[1] + "[" + testVari[3] + "] = function() {return true;};";
			}
		};
		patchCode({{.Workflow}});
		var subWorkflow = /(?:\w+=)?eval\((\w+)\)/.exec({{.Workflow}});
		if (subWorkflow) {
			var subPatch = (
				"{{.DebugMessages}}.{{.DecryptedCodes}}.push('sub workflow: ' + subWorkflow);" +
				"patchCode(subWorkflow);"
			).replace(/subWorkflow/g, subWorkflow[1]) + subWorkflow[0];
			{{.Workflow}} = {{.Workflow}}.replace(subWorkflow[0], subPatch);
		}
		eval({{.Workflow}});
	`))

	jsDebugTmpl = template.Must(template.New("jsDebug").Parse(`
		var {{.Ub98484234}} = ub98484234;
		ub98484234 = function(p1, p2, p3) {
			try {
				var resoult = {{.Ub98484234}}(p1, p2, p3);
				{{.DebugMessages}}.{{.Resoult}} = resoult;
			} catch(e) {
				{{.DebugMessages}}.{{.Resoult}} = e.message;
			}
			return {{.DebugMessages}};
		};
	`))
)

func render(tmpl *template.Template, data interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func loadCryptoJS() {
	body, err := http.Get("https://cdnjs.cloudflare.com/ajax/libs/crypto-js/3.1.9-1/crypto-js.min.js", nil, nil)
	if err != nil {
		// TODO: not panic
		panic(err)
	}
	cryptoJS = body
}

func getEngineWithCryptoJS() (*otto.Otto, error) {
	if cryptoJS == nil {
		loadCryptoJS()
	}
	engine := otto.New()
	if _, err := engine.Eval(cryptoJS); err != nil {
		return nil, err
	}
	return engine, nil
}

type Live struct {
	internal.BaseLive
	roomID string
}

func (l *Live) fetchRoomID() {
	if l.roomID != "" {
		return
	}
	l.roomID = strings.Split(l.Url.Path, "/")[1]
	body, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return
	}
	for _, reg := range douyuRoomIDRegs {
		strs := reg.FindStringSubmatch(string(body))
		if len(strs) == 2 {
			l.roomID = strs[1]
			return
		}
	}
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	l.fetchRoomID()
	body, err := http.Get(fmt.Sprintf("%s/%s", liveInfoUrl, l.roomID), nil, nil)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "error").Int() != 0 {
		return nil, live.ErrRoomNotExist
	}
	info = &live.Info{
		Live:     l,
		HostName: gjson.GetBytes(body, "data.owner_name").String(),
		RoomName: gjson.GetBytes(body, "data.room_name").String(),
		Status:   gjson.GetBytes(body, "data.room_status").String() == "1",
	}
	return info, nil

}

func (l *Live) getSignParams() (url.Values, error) {
	body, err := http.Get(liveEncUrl, nil, map[string]string{
		"rids": l.roomID,
	})
	if err != nil {
		return nil, err
	}

	jsEnc := gjson.GetBytes(body, fmt.Sprintf("data.room%s", l.roomID)).String()

	workflow := ""
	if workflowMatch := workflowReg.FindStringSubmatch(jsEnc); len(workflowMatch) == 2 {
		workflow = workflowMatch[1]
	}

	context := struct {
		DebugMessages  string
		DecryptedCodes string
		Resoult        string
		Ub98484234     string
		Workflow       string
	}{
		DebugMessages:  utils.GenRandomName(8),
		DecryptedCodes: utils.GenRandomName(8),
		Resoult:        utils.GenRandomName(8),
		Ub98484234:     utils.GenRandomName(8),
		Workflow:       workflow,
	}
	jsDom, err := render(jsDomTmpl, context)
	if err != nil {
		return nil, err
	}
	jsPatch, err := render(jsPatchTmpl, context)
	if err != nil {
		return nil, err
	}
	jsDebug, err := render(jsDebugTmpl, context)
	if err != nil {
		return nil, err
	}

	jsEnc = strings.ReplaceAll(jsEnc, fmt.Sprintf("eval(%s);", context.Workflow), jsPatch)
	engine, err := getEngineWithCryptoJS()
	if err != nil {
		return nil, err
	}
	if _, err := engine.Eval(jsDom); err != nil {
		return nil, err
	}
	if _, err := engine.Eval(jsEnc); err != nil {
		return nil, err
	}
	if _, err := engine.Eval(jsDebug); err != nil {
		return nil, err
	}
	did := strings.ReplaceAll(uuid.Must(uuid.NewV4()).String(), "-", "")
	ts := time.Now()
	res, err := engine.Call("ub98484234", nil, l.roomID, did, ts.Unix())
	if err != nil {
		return nil, err
	}
	values := url.Values{
		"cdn":  {""},
		"iar":  {"0"},
		"ive":  {"0"},
		"rate": {"0"},
	}
	resoult, err := res.Object().Get(context.Resoult)
	if err != nil {
		return nil, err
	}
	for _, entry := range strings.Split(resoult.String(), "&") {
		if entry == "" {
			continue
		}
		strs := strings.SplitN(entry, "=", 2)
		values.Set(strs[0], strs[1])
	}
	return values, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	l.fetchRoomID()
	params, err := l.getSignParams()
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(fmt.Sprintf("%s/%s", liveAPIUrl, l.roomID), header, nil, []byte(params.Encode()))
	if gjson.GetBytes(resp, "error").Int() != 0 {
		return nil, fmt.Errorf("get stream error")
	}
	return utils.GenUrls(
		fmt.Sprintf("%s/%s",
			gjson.GetBytes(resp, "data.rtmp_url").String(),
			gjson.GetBytes(resp, "data.rtmp_live").String(),
		),
	)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
