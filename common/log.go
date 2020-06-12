package common

import (
	runtimeFormatter "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"os"
)

var Log = logrus.New()

func initLog() {
	Log.Infoln("初始化日志配置......")
	Log.SetLevel(logrus.DebugLevel)
	Log.SetReportCaller(true)
	logFormatter := &runtimeFormatter.Formatter{
		File: true,
		Line: true,
		ChildFormatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat: "%time%  [%lvl%]	%file%:%line%	- %msg%	\n",
		},
	}
	Log.SetFormatter(logFormatter)
	Log.SetOutput(os.Stdout)
	//Log.SetOutput(Config.Logfile)
	Log.Infoln("日志初始化完成")
}
