package users

import (
	"github.com/jinzhu/gorm"
	"sync"
	"zeus/common"
	"zeus/external"
	"zeus/models"
)

func AddPermissionAssets(user *models.User, assets *models.Assets) error {
	// user不存在则创建
	if err := user.GetInfo(nil); gorm.IsRecordNotFoundError(err) {
		if err := user.Add(); err != nil {
			return err
		}
	}
	// save assets
	for _, ast := range *assets {
		if err := ast.Add(); err != nil {
			return err
		}
	}
	// replace relationship between user and assets
	user.Assets = *assets
	if err := user.ReplaceAssets(); err != nil {
		return err
	}
	return nil
}

func FetchPermissionAssets(user *models.User) error {
	if err := user.GetInfo(nil); err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, ast := range user.Assets {
		// 判断asset type
		if ast.Type == models.AssetTypeTag {
			wg.Add(1)
			go func(ast *models.Asset) {
				defer wg.Done()
				// TODO
				// 如果是"tag"，则需要从服务树拉取相应server资源进行填充
				//
				ast.Servers = external.FetchServersByTag(ast.Tag)
			}(ast)
		}
	}
	wg.Wait()
	return nil
}

func FilterPermissionServersByIDC(user *models.User, idc string) (ss models.Servers) {
	// 首先根据user获取对应权限资源
	if err := FetchPermissionAssets(user); err != nil {
		common.Log.Errorf("获取用户：%s的权限资源失败：%s", user.Username, err.Error())
		//return nil
	}
	for _, ast := range user.Assets {
		ss = append(ss, ast.Servers...)
	}
	////// 测试数据
	s := models.Server{}
	s.IP = "192.168.32.7"
	s.Hostname = "dev_server_01"
	s.IDC = "北京"
	s.Type = "ssh"
	s.Port = 22
	s1 := models.Server{}
	s1.IP = "172.16.244.28"
	s1.Hostname = "dev_server_02"
	s1.IDC = "天津"
	s1.Type = "ssh"
	s1.Port = 22
	ss = append(ss, []*models.Server{&s, &s1}...)
	//////
	// 根据IDC名称过滤权限资源
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	for _, ast := range user.Assets {
		for _, s := range ast.Servers {
			wg.Add(1)
			go func(s *models.Server) {
				defer wg.Done()
				if s.IDC == idc {
					lock.Lock()
					ss = append(ss, s)
					lock.Unlock()
				}
			}(s)
		}
	}
	wg.Wait()
	return
}
