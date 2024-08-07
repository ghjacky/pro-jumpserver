package permission

import (
	"github.com/jinzhu/gorm"
	"sync"
	"zeus/common"
	"zeus/external"
	"zeus/models"
)

func AddPermissions(username string, permissions *models.Permissions, db *gorm.DB) error {
	// user不存在则创建
	var user = models.User{Username: username}
	if err := user.GetInfo(nil); gorm.IsRecordNotFoundError(err) {
		if err := user.Add(); err != nil {
			return err
		}
	}
	// save permissions
	for _, perm := range *permissions {
		if err := perm.Add(db); err != nil {
			return err
		}
		// if sudo needed (don't needed), add (delete) sudo permission in background jobs
		go PermSudo(perm)
	}
	return nil
}

func FetchPermissions(user *models.User) error {
	if err := user.GetInfo(nil); err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, perm := range user.Permissions {
		// 判断asset type
		if perm.Type == models.PermissionTypeTag {
			wg.Add(1)
			go func(perm *models.Permission) {
				defer wg.Done()
				// TODO
				// 如果是"tag"，则需要从服务树拉取相应server资源进行填充
				//
				perm.Servers = external.FetchServersByTag(perm.Tag)
			}(perm)
		} else if perm.Type == models.PermissionTypeServer {
			wg.Add(1)
			go func() {
				defer wg.Done()
			}()
		}
	}
	wg.Wait()
	return nil
}

func FetchPermissionServers(user *models.User) (ss models.Servers) {
	// 首先根据user获取对应权限资源
	if err := FetchPermissions(user); err != nil {
		common.Log.Errorf("获取用户：%s的权限资源失败：%s", user.Username, err.Error())
		//return nil
	}
	for _, perm := range user.Permissions {
		ss = append(ss, perm.Servers...)
	}
	return
}

func FetchAllPermissions(query models.Query) (models.Permissions, int, error) {
	var perms = &models.Permissions{}
	total, err := perms.FetchList(query)
	if err != nil {
		return nil, total, err
	}
	return *perms, total, nil
}

func DeletePermission(perm *models.Permission, db *gorm.DB) (err error) {
	if err := perm.GetInfo(); err != nil {
		return err
	}
	return perm.Delete(db)
}

func Update(perm *models.Permission, db *gorm.DB) (err error) {
	return perm.Update(db)
}
