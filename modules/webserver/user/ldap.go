package user

import (
	"github.com/go-ldap/ldap/v3"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"zeus/common"
	"zeus/models"
)

var Users models.Users
var ldapItemsLength = 0

// 从ldap获取用户列表
func FetchUserFromLDAP() (err error) {
	searchRequest := ldap.NewSearchRequest(viper.GetString("ldap.search_scope"), ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false, "(&(objectClass=User))",
		[]string{"mail", "cn", "sAMAccountName", "description"}, nil)
	sr, err := common.LdapConn.Search(searchRequest)
	if err != nil {
		common.Log.Errorf("_ldap 搜索错误：%s", err.Error())
		return
	}
	// 先循环去掉非员工账户
	if len(sr.Entries) != ldapItemsLength {
		ldapItemsLength = len(sr.Entries)
		Users = models.Users{}
		//Us_ = []user{}
		for _, entry := range sr.Entries {
			if len(entry.GetAttributeValue("mail")) == 0 {
				continue
			}
			user := models.User{
				Username: entry.GetAttributeValue("sAMAccountName"),
				Nickname: entry.GetAttributeValue("cn"),
				Email:    entry.GetAttributeValue("mail"),
			}
			Users = append(Users, &user)
			if err := user.GetInfo(nil); gorm.IsRecordNotFoundError(err) {
				_ = user.Add()
			}
		}
	}
	return
}
