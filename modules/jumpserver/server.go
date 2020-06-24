package jumpserver

import (
	"context"
	"github.com/gliderlabs/ssh"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"os"
	"os/user"
	"path"
	"zeus/common"
)

const (
	ServerStatusUnkown   = "Unknown"
	ServerStatusRunning  = "Running"
	ServerStatusStopping = "Stopping"
	ServerStatusStopped  = "Stopped"
	ServerSignalRunning  = 1
	ServerSignalStopping = -1
)

type JumpConfig struct {
	Listen  string
	User    string
	Mysql   *gorm.DB
	Redis   *redis.Client
	WorkDir string
	DataDir string
	PlayDir string
}
type JumpServer struct {
	JumpConfig
	Context   context.Context
	Status    string
	Signal    chan int
	Namespace uuid.UUID
	Server    *ssh.Server
}

// 根据配置生成jumpserver实例
func NewJumpServer(jc JumpConfig) (js *JumpServer) {
	js = &JumpServer{}
	js.JumpConfig = jc
	js.Context = context.TODO()
	js.Status = ServerStatusStopped
	js.Signal = make(chan int, 0)
	js.Namespace = uuid.New()
	return js
}

var JPS *JumpServer
var SessionNormalEventRecordDir string
var SessionKBEventRecordDir string

func InitJumpServer() {
	var currentUser, _ = user.Current()
	JPS = NewJumpServer(JumpConfig{
		Listen:  common.Config.JumpServerAddr,
		User:    currentUser.Username,
		Mysql:   common.Mysql,
		Redis:   common.Redis,
		WorkDir: common.Config.WorkDir,
		DataDir: common.Config.DataDir,
		PlayDir: common.Config.PlayDir,
	})
	// 创建event存储目录（文件存储器）
	SessionNormalEventRecordDir = path.Join(common.Config.DataDir, "sessions", "events", "normal")
	SessionKBEventRecordDir = path.Join(common.Config.DataDir, "sessions", "events", "kb")
	if err := os.MkdirAll(SessionNormalEventRecordDir, common.DefaultDirMode); err != nil {
		common.Log.Fatalf("couldn't create session normal events store dir: %s", SessionNormalEventRecordDir)
	}
	if err := os.MkdirAll(SessionKBEventRecordDir, common.DefaultDirMode); err != nil {
		common.Log.Fatalf("couldn't create session kb events store dir: %s", SessionKBEventRecordDir)
	}
}

// 运行jumpserver
func (js *JumpServer) Run() {

	// 获取hostkey， 无则生成
	hostkeyPath := path.Join(js.WorkDir, ".js", "hostkey")
	hostkey := HostKey{
		Path:  hostkeyPath,
		Value: "",
	}
	sshSigner, err := hostkey.Load()
	if err != nil {
		common.Log.Fatalf("Couldn't load host key from file: %s", hostkeyPath)
	}
	sshServer := &ssh.Server{
		Addr:                       js.Listen,
		KeyboardInteractiveHandler: checkKBI,
		PasswordHandler:            checkUserPassword,
		PublicKeyHandler:           checkUserPublicKey,
		HostSigners:                []ssh.Signer{sshSigner},
		Handler:                    sessionHandlerWrapper,
	}
	js.Server = sshServer
	// change server status to running
	js.Status = ServerStatusRunning
	common.Log.Fatal(sshServer.ListenAndServe())
}

func (js *JumpServer) Stop() {
	// graceful stop server, waiting for closing existed connections
	if err := js.Server.Shutdown(context.TODO()); err != nil {
		common.Log.Errorf("Failed to shutdown jump server")
	}
}
