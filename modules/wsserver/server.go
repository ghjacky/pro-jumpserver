package wsserver

import (
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
	"zeus/common"
)

func newServer() (server *socketio.Server, err error) {
	server, err = socketio.NewServer(nil)
	if err != nil {
		common.Log.Errorf("Websocket server created error: %s", err.Error())
		return nil, err
	}
	return
}

func servHttp(server *socketio.Server) {
	router := gin.New()
	router.GET(common.Config.WSPath+"/*any", gin.WrapH(server))
	router.POST(common.Config.WSPath+"/*any", gin.WrapH(server))
	router.StaticFile("/", common.Config.WSStaticPath)
	if err := router.Run(common.Config.WSListen); err != nil {
		common.Log.Errorf(err.Error())
	}
}

func Run() {
	server, err := newServer()
	if err != nil {
		return
	}
	registerMustListener(server)
	registerOtherListener(server)
	go func() {
		if err := server.Serve(); err != nil {
			common.Log.Errorf("Failed to run websocket server: %s", err.Error())
		}
	}()
	defer server.Close()
	servHttp(server)
}
