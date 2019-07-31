package audit

import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"io"
	"os"
	"zeus/common"
)

const (
	StoreRedis = "redisStore"
	StoreMysql = "mysqlStore"
	StoreFile  = "fileStore"
)

// 事件存储器接口
type IStore interface {
	io.WriteCloser
}

// mysql存储器
type MysqlStore struct {
	client *gorm.DB
}

func (ms MysqlStore) Write(data []byte) (n int, err error) {
	return
}

func (ms MysqlStore) Close() (err error) {
	return
}

// redis存储器
type RedisStore struct {
	client *redis.Client
}

func (rs RedisStore) Write(data []byte) (n int, err error) {

	return
}
func (rs RedisStore) Close() (err error) {
	return
}

// 文件存储器
type FileStore struct {
	File *os.File
}

func (fs FileStore) Write(data []byte) (n int, err error) {
	return fs.File.Write(data)
}

func (fs FileStore) Close() (err error) {
	if err = fs.File.Close(); err != nil {
		common.Log.Warnf("Couldn't close filestore: %s", fs.File.Name())
	}
	return
}

// 创建存储器
func NewStore(t string, args ...interface{}) (store IStore) {
	switch t {
	case StoreFile:
		for _, arg := range args {
			switch v := arg.(type) {
			case string:
				f, err := os.OpenFile(v, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, common.DefaultFileMode)
				if err != nil {
					return
				}
				store = FileStore{f}
			default:
				break
			}
			break
		}
	}
	return
}
