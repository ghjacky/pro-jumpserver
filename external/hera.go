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
	ApiFetchTags         = "/trees-not-verify/0/tags?page=1&page_per=9999"
	ApiFetchEnvs         = "/help/trees-not-verify/resource/options"
	ApiFetchServersByTag = "/trees-not-verify/${tid}/resources?exact=true"
)

func doHeraRequest(api string) *http.Response {
	hc := NewHttpClient()
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		common.Log.Errorf("Create http request error: %s", err.Error())
		return nil
	}
	at := utils.AccessToken{}
	at.ServiceName = "zeus"
	at.RequestTime = time.Now().Unix()
	req.Header.Set(utils.TokenNameInHeader, at.GenerateToken())
	resp, err := hc.Do(req)
	if err != nil {
		common.Log.Errorf("http request error: %s", err.Error())
		return nil
	}
	return resp
}

func FetchTags() (tags []map[string]interface{}) {
	api := strings.Join([]string{common.Config.HeraConfig.Addr, common.Config.HeraConfig.ApiPrefix, strings.Trim(ApiFetchTags, "/")}, "/")
	resp := doHeraRequest(api)
	respData, err := ParseResponse(resp)
	if err != nil {
		common.Log.Errorf("response body parse error: %s", err.Error())
		return
	}
	if int(respData["code"].(float64)) != 100000 {
		return nil
	}
	for _, t := range respData["data"].([]interface{}) {
		if v, ok := t.(map[string]interface{}); ok {
			tags = append(tags, v)
		}
	}
	return
}

func FetchEnvs() (envs []string) {
	api := strings.Join([]string{common.Config.HeraConfig.Addr, common.Config.HeraConfig.ApiPrefix, strings.Trim(ApiFetchEnvs, "/")}, "/")
	resp := doHeraRequest(api)
	respData, err := ParseResponse(resp)
	if err != nil {
		common.Log.Errorf("response body parse error: %s", err.Error())
		return nil
	}
	if int(respData["code"].(float64)) != 100000 {
		return nil
	}
	if v, ok := respData["data"].(map[string]interface{}); ok {
		if vv, okk := v["idc_list"].([]interface{}); okk {
			for _, vvv := range vv {
				if vvvv, okkk := vvv.(map[string]interface{}); okkk {
					envs = append(envs, vvvv["idc_name"].(string))
				}
			}
		}
	}
	return
}

func FetchServersByTag(tname string) (servers models.Servers) {
	tid := getTagIdByName(tname)
	api := strings.Join([]string{common.Config.HeraConfig.Addr, common.Config.HeraConfig.ApiPrefix, strings.Trim(strings.ReplaceAll(ApiFetchServersByTag, `${tid}`, fmt.Sprintf("%d", tid)), "/")}, "/")
	resp := doHeraRequest(api)
	respData, err := ParseResponse(resp)
	if err != nil {
		common.Log.Errorf("response body parse error: %s", err.Error())
		return nil
	}
	if int(respData["code"].(float64)) != 100000 {
		return nil
	}
	if v, ok := respData["data"].([]interface{}); ok {
		for _, vv := range v {
			if vvv, okkk := vv.(map[string]interface{}); okkk {
				servers = append(servers, &models.Server{
					Type:     "ssh",
					Hostname: vvv["hostname"].(string),
					IP:       vvv["ipv4"].(string),
					IDC:      vvv["idc_name"].(string),
					Port:     22,
				})
			}
		}
	}
	return
}

func getTagIdByName(tname string) (tid uint) {
	if len(common.Config.Tags) != 0 {
		for _, tag := range common.Config.Tags {
			if strings.EqualFold(tag["tag"].(string), tname) {
				return uint(int(tag["id"].(float64)))
			}
		}
	}
	return 0
}
