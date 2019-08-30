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
	RpcAddr         = App.Flag("rpc-addr", "RPC server listen port").Default(":8080").String()
	RpcToken        = App.Flag("rpc-token", "RPC server token.").String()
	RpcTLS          = App.Flag("enable-rpc-tls", "Enable TLS for RPC server").Bool()
	CertFile        = App.Flag("rpc-tls-cert-file", "Cert file for TLS on RPC").String()
	KeyFile         = App.Flag("rpc-tls-key-file", "Key file for TLS on RPC").String()
	NativeFlvParser = App.Flag("native-flv-parser", "use native flv parser").Default("false").Bool()
)

func init() {
	kingpin.MustParse(App.Parse(os.Args[1:]))
}

func GenConfigFromFlags() *configs.Config {
	return &configs.Config{
		RPC: configs.RPC{
			Enable: *Rpc,
			Bind:   *RpcAddr,
			Token:  *RpcToken,
			TLS: configs.TLS{
				Enable:   *RpcTLS,
				CertFile: *CertFile,
				KeyFile:  *KeyFile,
			},
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
