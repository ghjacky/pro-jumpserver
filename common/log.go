package common

import (
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var Log = logrus.New()

func initLog() {
	Log.Debugln("starting init log configuration")
	logFormatter := &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat: " %time%  [%lvl%]	- %msg% \n",
	}
	Log.SetFormatter(logFormatter)
	Log.SetLevel(logrus.DebugLevel)
	Log.SetOutput(Config.mainConfig.LogFile)
	Log.Debugln("init log configuration successfully")
}
