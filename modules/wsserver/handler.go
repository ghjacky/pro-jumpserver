package wsserver

import (
	"encoding/json"
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	"path"
	"strings"
	"zeus/common"
	"zeus/modules/audit"
)

const (
	NSPPlay = "play"
)

// 生命周期相关必要handler
func handleConnect(conn socketio.Conn) error {
	common.Log.Infof("%s connected", conn.ID())
	conn.SetContext("")
	initialize(conn)
	return nil
}

func handleJoinRoom(conn socketio.Conn, room string) {
	common.Log.Infof("Joining room: %s", room)
	conn.Join(room)
}

func handleDisconnect(conn socketio.Conn, reason string) {
	common.Log.Infof("Websocket disconnected: %s", reason)
	clear(conn)
}

func handleError(conn socketio.Conn, err error) {
	common.Log.Errorf("Error on connected to websocket: %s", err.Error())
	clear(conn)
}

func clear(conn socketio.Conn) {
	conn.LeaveAll()
	conn.Close()
}

func initialize(conn socketio.Conn) {
}

// 其他handler
func handlePlay(conn socketio.Conn, msg string) {
	msgm := map[string]string{}
	if err := json.Unmarshal([]byte(msg), &msgm); err != nil {
		conn.Emit("error", fmt.Sprintf("Got wrone arguments: %s", err.Error()))
		return
	}
	SessionKBEventRecordDir := path.Join(common.Config.DataDir, "sessions", "events", "kb")
	filePrefix := audit.EventTypeKeyBoardPress
	serverIP := msgm["server_ip"]
	sessionID := msgm["session_id"]
	ID := msgm["id"]
	filename := strings.Join([]string{filePrefix, serverIP, sessionID, ID}, "_")
	filepath := SessionKBEventRecordDir + "/" + filename
	go play(filepath, conn)
}

func handleStopPlay(conn socketio.Conn, msg int) {
	go playStop(msg)
}
