package common

import "github.com/spf13/viper"

type SConfig struct {
	Listen string
	Log    string
}

var ConfigPath *string
var Config = &SConfig{}

func ParseConfig() {
	viper.SetConfigFile(*ConfigPath)
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		Log.Fatalf("配置文件读取错误")
	}
	Config.Listen = viper.GetString("main.listen")
	Config.Log = viper.GetString("main.log")
}
