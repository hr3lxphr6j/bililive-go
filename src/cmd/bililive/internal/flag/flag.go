package flag

import (
	"os"

	"github.com/alecthomas/kingpin"

	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/consts"
)

var (
	app = kingpin.New(consts.AppName, "A command-line live stream save tools.").Version(consts.AppVersion)

	Debug           = app.Flag("debug", "Enable debug mode.").Default("false").Bool()
	Interval        = app.Flag("interval", "Interval of query live status").Default("20").Short('t').Int()
	Output          = app.Flag("output", "Output file path.").Short('o').Default("./").String()
	Input           = app.Flag("input", "Live room urls").Short('i').Strings()
	Conf            = app.Flag("config", "Config file.").Short('c').String()
	RPC             = app.Flag("enable-rpc", "Enable RPC server.").Default("false").Bool()
	RPCBind         = app.Flag("rpc-bind", "RPC server bind address").Default(":8080").String()
	NativeFlvParser = app.Flag("native-flv-parser", "use native flv parser").Default("false").Bool()
)

func init() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

// GenConfigFromFlags generates configuration by parsing command line parameters.
func GenConfigFromFlags() *configs.Config {
	return &configs.Config{
		RPC: configs.RPC{
			Enable: *RPC,
			Bind:   *RPCBind,
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
