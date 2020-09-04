package permission

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"sync"
	"zeus/common"
	"zeus/external"
	"zeus/models"
	"zeus/modules/assets"
	"zeus/utils"
)

func PermSudo(perm *models.Permission) {
	// first, get all servers with sudo permission needed
	getServersOfPermission(perm)
	// second, create goroutine for adding sudo permission by exec command through ssh connection
	wg := sync.WaitGroup{}
	for i, server := range perm.Servers {
		wg.Add(1)
		go func(server *models.Server) {
			defer wg.Done()
			execThroughSsh(perm.Username, perm.Sudo, server)
		}(server)
		if (i+1)%10 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
}

func getServersOfPermission(perm *models.Permission) {
	// do something base on perm.type
	if perm.Type == models.PermissionTypeTag {
		perm.Servers = external.FetchServersByTag(perm.Tag)
	}
}

func execThroughSsh(username string, sudo uint8, server *models.Server) {
	hostkey := utils.HostKey{
		Path:  models.UserRootPubKeyPath,
		Value: common.Config.PrivateKey,
	}
	// before everything, you should create a asset and connect it while judge it if the connection needed a proper proxy
	as := &assets.ASSH{}
	as.IDC = server.IDC
	as.IP = server.IP
	as.PORT = server.Port
	as.USER = models.UserRoot
	as.HOSTKEY = &hostkey
	ias, err := assets.NewAssetClient(as)
	if err != nil {
		common.Log.Errorf("failed to add sudo permission for user: %s (connect): %s", username, err.Error())
		return
	}
	session := ias.NewSession().(*ssh.Session)
	cmd := ""
	if sudo == 0x01 {
		cmd = fmt.Sprintf("usermod -a -G rops %s", username)
	} else {
		cmd = fmt.Sprintf("deluser %s rops", username)
	}
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		common.Log.Errorf("failed to add sudo permission for user: %s (exec): %s", username, err.Error())
	}
	common.Log.Warnf("add sudo permission for user: %s (stdout): %s", username, string(out))
}
