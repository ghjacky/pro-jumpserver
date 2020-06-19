package external

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func NewHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	return httpClient
}

func ParseResponse(resp *http.Response) (bodyMap map[string]interface{}, err error) {
	bodyBytes, e := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if e != nil {
		return bodyMap, e
	}
	if err = json.Unmarshal(bodyBytes, &bodyMap); err != nil {
		return
	}
	return
}
