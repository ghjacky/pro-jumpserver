package wsserver

import socketio "github.com/googollee/go-socket.io"

func registerMustListener(server *socketio.Server) {
	server.OnConnect("/", handleConnect)
	server.OnDisconnect("/", handleDisconnect)
	server.OnError("/", handleError)
	server.OnEvent("/", "join", handleJoinRoom)
}

func registerOtherListener(server *socketio.Server) {
	server.OnEvent("/", "playing", handlePlay)
	server.OnEvent("/", "play_term", handleStopPlay)
}
