package jumpserver

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
	"zeus/common"
	"zeus/utils"
)

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

var IDCs []string

func init() {
	go func() {
		for {
			if len(common.Config.IDCs) != 0 {
				IDCs = append(common.Config.IDCs, "全部")
				break
			}
		}
	}()
}

func newDefaultBanner() Banner {
	defaultTitle := utils.WrapperTitle("欢迎使用跳板机, 您已进入被监控状态") // +
	//utils.WrapperTitle("您有权保持沉默") + utils.WrapperTitle("您的所有操作都将被记录作为呈堂证供") +
	//utils.WrapperTitle("知悉！")
	defaultMenu := Menu{
		{showText: "请输入序号切换相应IDC:\n"},
	}
	for index, idc := range IDCs {
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
	line := fmt.Sprintf("\t%d) 输入 {{.GreenBoldColor}}%s{{.ColorEnd}} %s.%s", mi.id, mi.instruct, mi.helpText, "\r\n")
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
		{showText: fmt.Sprintf("当前IDC: %s\n", idc)},
		{id: 1, instruct: "ID", helpText: "直接登陆"},
		{id: 2, instruct: "part IP, Hostname, Comment", helpText: "搜索，如果资产唯一，则直接登陆"},
		{id: 3, instruct: "/ + IP, Hostname, Comment", helpText: "搜索, 例: /192.168"},
		{id: 4, instruct: "p", helpText: "展示您有权限的资源"},
		{id: 6, instruct: "r", helpText: "返回上一级"},
		{id: 7, instruct: "h", helpText: "帮助菜单"},
		{id: 8, instruct: "q", helpText: "退出"},
	}
	b.Menu = mainMenu
}
