package external

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"zeus/common"
	"zeus/models"
	"zeus/utils"
)

const (
	ApiFetchTree = "/trees"
)

func FetchTags() (tags []string) {
	api := strings.Join([]string{common.Config.HeraConfig.Addr, common.Config.HeraConfig.ApiPrefix, strings.Trim(ApiFetchTree, "/")}, "/")
	hc := NewHttpClient()
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		common.Log.Errorf("Create http request error: %s", err.Error())
		return
	}
	at := utils.AccessToken{}
	at.ServiceName = "eim-ec"
	at.RequestTime = time.Now().Unix()
	req.Header.Set(utils.TokenNameInHeader, at.GenerateToken())
	resp, err := hc.Do(req)
	if err != nil {
		common.Log.Errorf("http request error: %s", err.Error())
		return
	}
	respData, err := ParseResponse(resp)
	if err != nil {
		common.Log.Errorf("response body parse error: %s", err.Error())
		return
	}
	fmt.Printf("%#v", respData)
	tags = append(tags, []string{"com.group.dev", "com.group02.dev"}...)
	return
}

func FetchServersByTag(tag string) (servers models.Servers) {
	s := models.Server{}
	s.IP = "172.16.244.28"
	s.Hostname = "dev_server"
	s.IDC = "天津"
	s.Type = "ssh"
	s.Port = 22
	servers = append(servers, &s)
	return
}
