package configure

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/kr/pretty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/*
{
  "server": [
    {
      "appname": "live",
      "live": true,
	  "hls": true,
	  "static_push": []
    }
  ]
}
*/

type Application struct {
	Appname    string   `mapstructure:"appname"`
	Live       bool     `mapstructure:"live"`
	Hls        bool     `mapstructure:"hls"`
	Flv        bool     `mapstructure:"flv"`
	Api        bool     `mapstructure:"api"`
	StaticPush []string `mapstructure:"static_push"`
}

type Applications []Application

type JWT struct {
	Secret    string `mapstructure:"secret"`
	Algorithm string `mapstructure:"algorithm"`
}
type ServerCfg struct {
	Level           string       `mapstructure:"level" json:"level"`
	ConfigFile      string       `mapstructure:"config_file" json:"config_file"`
	FLVArchive      bool         `mapstructure:"flv_archive" json:"flv_archive"`
	FLVDir          string       `mapstructure:"flv_dir" json:"flv_dir"`
	RTMPNoAuth      bool         `mapstructure:"rtmp_noauth" json:"rtmp_no_auth"`
	RTMPAddr        string       `mapstructure:"rtmp_addr" json:"rtmp_addr"`
	HTTPFLVAddr     string       `mapstructure:"httpflv_addr" json:"httpflv_addr"`
	HLSAddr         string       `mapstructure:"hls_addr" json:"hls_addr"`
	HLSKeepAfterEnd bool         `mapstructure:"hls_keep_after_end" json:"hls_keep_after_end"`
	APIAddr         string       `mapstructure:"api_addr" json:"api_addr"`
	RedisAddr       string       `mapstructure:"redis_addr" json:"redis_addr"`
	RedisPwd        string       `mapstructure:"redis_pwd" json:"redis_pwd"`
	ReadTimeout     int          `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout    int          `mapstructure:"write_timeout" json:"write_timeout"`
	GopNum          int          `mapstructure:"gop_num" json:"gop_num"`
	JWT             JWT          `mapstructure:"jwt" json:"jwt"`
	Server          Applications `mapstructure:"server" json:"server"`
}

// default config
var defaultConf = ServerCfg{
	ConfigFile:      "livego.yaml",
	FLVArchive:      false,
	RTMPNoAuth:      false,
	RTMPAddr:        ":1935",
	HTTPFLVAddr:     ":7001",
	HLSAddr:         ":7002",
	HLSKeepAfterEnd: false,
	APIAddr:         ":8090",
	WriteTimeout:    10,
	ReadTimeout:     10,
	GopNum:          1,
	Server: Applications{{
		Appname:    "live",
		Live:       true,
		Hls:        true,
		Flv:        true,
		Api:        true,
		StaticPush: nil,
	}},
}

var Config = viper.New()

func initLog() {
	if l, err := log.ParseLevel(Config.GetString("level")); err == nil {
		log.SetLevel(l)
		log.SetReportCaller(l == log.DebugLevel)
	}
}

func init() {
	defer Init()

	// Default config
	b, _ := json.Marshal(defaultConf)
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("json")
	viper.ReadConfig(defaultConfig)
	Config.MergeConfigMap(viper.AllSettings())

	// Environment
	replacer := strings.NewReplacer(".", "_")
	Config.SetEnvKeyReplacer(replacer)
	Config.AllowEmptyEnv(true)
	Config.AutomaticEnv()

	// Log
	initLog()

	// Print final config
	c := ServerCfg{}
	Config.Unmarshal(&c)
	log.Debugf("Current configurations: \n%# v", pretty.Formatter(c))
}

func SetConfig(config ServerCfg) (err error) {
	b, _ := json.Marshal(config)
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("json")
	err = viper.ReadConfig(defaultConfig)
	if err != nil {
		return err
	}
	err = Config.MergeConfigMap(viper.AllSettings())
	if err != nil {
		return err
	}

	return nil
}

func CheckAppName(appname string) bool {
	apps := Applications{}
	Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
		if app.Appname == appname {
			return app.Live
		}
	}
	return false
}

func GetStaticPushUrlList(appname string) ([]string, bool) {
	apps := Applications{}
	Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
		if (app.Appname == appname) && app.Live {
			if len(app.StaticPush) > 0 {
				return app.StaticPush, true
			} else {
				return nil, false
			}
		}
	}
	return nil, false
}
