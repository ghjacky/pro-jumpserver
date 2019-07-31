package jumpserver

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
	"zeus/common"
	"zeus/utils"
)

const (
	IDCTJ  = "天津"
	IDCBJ  = "北京"
	IDCHK  = "香港"
	IDCAM  = "美国"
	IDCALL = "全部"
)

var IDCS = []string{IDCBJ, IDCHK, IDCTJ, IDCAM, IDCALL}

type MenuItem struct {
	id       int
	instruct string
	helpText string
	showText string
}

type Banner struct {
	Menu
	Title string
}

func newDefaultBanner() Banner {
	defaultTitle := utils.WrapperTitle("欢迎使用跳板机") + utils.WrapperTitle("您已进入被监控状态") // +
	//utils.WrapperTitle("您有权保持沉默") + utils.WrapperTitle("您的所有操作都将被记录作为呈堂证供") +
	//utils.WrapperTitle("知悉！")
	defaultMenu := Menu{
		{showText: "请输入序号切换相应IDC:\n"},
	}
	for index, idc := range IDCS {
		defaultMenu = append(defaultMenu, MenuItem{showText: fmt.Sprintf("	%d) %s\n", index, idc)})
	}
	defaultMenu = append(defaultMenu, MenuItem{showText: "输入q退出！\n\n"})
	return Banner{defaultMenu, defaultTitle}
}

func (mi *MenuItem) Text() string {
	if mi.showText != "" {
		return mi.showText
	}
	cm := ColorMeta{GreenBoldColor: "\033[1;32m", ColorEnd: "\033[0m"}
	line := fmt.Sprintf("\t%d) Enter {{.GreenBoldColor}}%s{{.ColorEnd}} to %s.%s", mi.id, mi.instruct, mi.helpText, "\r\n")
	tmpl := template.Must(template.New("item").Parse(line))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, cm)
	if err != nil {
		common.Log.Error(err)
	}
	mi.showText = string(buf.Bytes())
	return mi.showText
}

type Menu []MenuItem

type ColorMeta struct {
	GreenBoldColor string
	ColorEnd       string
}

func displayBanner(sess io.ReadWriter, user string, banner Banner) {
	title := banner.Title

	prefix := utils.CharClear
	suffix := utils.CharNewLine + utils.CharNewLine
	welcomeMsg := prefix + utils.WrapperTitle(user+", 你好") + title + suffix
	_, err := io.WriteString(sess, welcomeMsg)
	if err != nil {
		common.Log.Errorf("Send to client error, %s", err)
		return
	}
	for _, v := range banner.Menu {
		utils.IgnoreErrWriteString(sess, v.Text())
	}
}

func (b *Banner) setMainMenu(idc string) {
	mainMenu := Menu{
		{showText: fmt.Sprintf("IDC: %s\n", idc)},
		{id: 1, instruct: "ID", helpText: "directly login"},
		{id: 2, instruct: "part IP, Hostname, Comment", helpText: "to search login if unique"},
		{id: 3, instruct: "/ + IP, Hostname, Comment", helpText: "to search, such as: /192.168"},
		{id: 4, instruct: "p", helpText: "display the host you have permission"},
		{id: 5, instruct: "g", helpText: "display the node that you have permission"},
		{id: 6, instruct: "r", helpText: "refresh your assets and nodes"},
		{id: 7, instruct: "h", helpText: "print help"},
		{id: 8, instruct: "q", helpText: "exit"},
	}
	b.Menu = mainMenu
}
