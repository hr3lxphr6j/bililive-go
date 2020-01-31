package flag

import (
	"os"

	"github.com/alecthomas/kingpin"

	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/consts"
)

var (
	App             = kingpin.New(consts.AppName, "A command-line live stream save tools.").Version(consts.AppVersion)
	Debug           = App.Flag("debug", "Enable debug mode.").Default("false").Bool()
	Interval        = App.Flag("interval", "Interval of query live status").Default("20").Short('t').Int()
	Output          = App.Flag("output", "Output file path.").Short('o').Default("./").String()
	Input           = App.Flag("input", "Live room urls").Short('i').Strings()
	Conf            = App.Flag("config", "Config file.").Short('c').String()
	Rpc             = App.Flag("enable-rpc", "Enable RPC server.").Default("false").Bool()
	RpcBind         = App.Flag("rpc-bind", "RPC server bind address").Default(":8080").String()
	NativeFlvParser = App.Flag("native-flv-parser", "use native flv parser").Default("false").Bool()
)

func init() {
	kingpin.MustParse(App.Parse(os.Args[1:]))
}

func GenConfigFromFlags() *configs.Config {
	return &configs.Config{
		RPC: configs.RPC{
			Enable: *Rpc,
			Bind:   *RpcBind,
		},
		Debug:      *Debug,
		Interval:   *Interval,
		OutPutPath: *Output,
		LiveRooms:  *Input,
		Feature: configs.Feature{
			UseNativeFlvParser: *NativeFlvParser,
		},
	}
}
