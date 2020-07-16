package main

import (
	"flag"
	"math/rand"
	"time"
	"zeus/proxy/common"
	"zeus/proxy/server"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	common.ConfigPath = flag.String("config", "./configs/config.toml", "")
	flag.Parse()
	common.ParseConfig()
	common.InitLog()
	server.ProxyServerRun()
}
