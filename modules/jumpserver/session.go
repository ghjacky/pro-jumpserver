package jumpserver

import "github.com/gliderlabs/ssh"

const (
	SessionRedisPrefix = "zeus_jump_session"
)

func sessionHandler(session ssh.Session) {
	JPS.Context

}
