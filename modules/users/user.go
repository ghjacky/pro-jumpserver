package users

import (
	"github.com/jinzhu/gorm"
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
	// create relationship between user and assets
	user.Assets = *assets
	if err := user.ReplaceAssets(); err != nil {
		return err
	}
	return nil
}
