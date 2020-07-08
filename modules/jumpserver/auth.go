package jumpserver

import (
	"encoding/base32"
	"fmt"
	"github.com/dgryski/dgoogauth"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"io/ioutil"
	"net/url"
	"rsc.io/qr"
	"zeus/common"
	"zeus/models"
	"zeus/modules/webserver/users"
)

const (
	kbiQuestionPassword = "Password: "
	kbiQuestionCode     = "Google Authentication Code: "
	kbiInstruction      = `
######################################################################################
#### Using user password and google authenticator for Two-Factor-Authentication ! ####
######################################################################################
`
)

var gacSecret = []byte{'&', 'b', 'A', '!', ';', 'O', '\'', 'd', 'z', '1'}

func checkKBI(ctx ssh.Context, challenge ssh2.KeyboardInteractiveChallenge) (res bool) {
	// 登陆认证
	username := ctx.User()
	answers, err := challenge(username, kbiInstruction, []string{kbiQuestionPassword, kbiQuestionCode}, []bool{false, true})
	if err != nil || len(answers) != 2 {
		return
	}
	password := answers[0]
	// 首先检测账户是否可用
	user := models.User{Username: username}
	if !users.IsValid(&user) {
		return false
	}
	//code := answers[1]
	// GAC + LDAP认证
	res = authLDAP(user, password)
	//res = authGAC(code) && authLDAP(username, password)
	if res {
		// 登陆成功，将用户信息写入context
		ctx.SetValue("loginUser", username)
		ctx.SetValue("loginPass", password)
		common.Log.Infof("用户：%s 登陆成功", username)
	} else {
		common.Log.Errorf("用户：%s 登陆失败", username)
	}
	return
}

//
// 密码认证
func checkUserPassword(ctx ssh.Context, password string) (res bool) {
	if len(password) != 0 {
		username := ctx.User()
		// 走ldap认证
		res = authLDAP(models.User{Username: username}, password)
		if res {
			ctx.SetValue("loginUser", username)
			ctx.SetValue("loginPass", password)
			common.Log.Infof("用户：%s 登陆成功", username)
		} else {
			common.Log.Errorf("用户：%s 登陆失败", username)
		}
	}
	common.Log.Debugln("checkUserPassword")
	return res
}

//
// 公钥认证
func checkUserPublicKey(ctx ssh.Context, publickey ssh.PublicKey) (res bool) {

	return false
}

// ldap 认证
func authLDAP(user models.User, pass string) (res bool) {
	defer func() {
		_ = common.LdapConn.Bind(common.Config.LdapConfig.BindUser, common.Config.LdapConfig.Password)
	}()
	if err := common.LdapConn.Bind(fmt.Sprintf("%s@aibee", user.Username), pass); err != nil {
		common.Log.Errorf("Authentication failure for user (%s): %s", user.Username, err.Error())
		return false
	} else {
		common.Log.Infof("Login successfully: %s", user.Username)
		return true
	}
}

// google一次性验证码认证
func authGAC(code string) (res bool) {
	secretBase32 := base32.StdEncoding.EncodeToString(gacSecret)

	// The OTPConfig gets modified by otpc.Authenticate() to prevent passcode replay, etc.,
	// so allocate it once and reuse it for multiple calls.
	otpc := &dgoogauth.OTPConfig{
		Secret:      secretBase32,
		WindowSize:  3,
		HotpCounter: 0,
		// UTC:         true,
	}

	//for {
	//	var token string
	//	fmt.Printf("Please enter the token value (or q to quit): ")
	//	fmt.Scanln(&token)
	//
	//	if token == "q" {
	//		break
	//	}

	val, err := otpc.Authenticate(code)
	if err != nil {
		return
	}

	if !val {
		return
	}
	res = true
	//}
	return
}

// 生成google一次性验证码认证服务二维码
func GenGACqr() {
	qrFilename := "/tmp/qr.png"
	secretBase32 := base32.StdEncoding.EncodeToString(gacSecret)
	fmt.Println(secretBase32)
	account := "chaos@aibee.cn"
	issuer := "zeus"

	URL, err := url.Parse("otpauth://totp")
	if err != nil {
		panic(err)
	}

	URL.Path += "/" + url.PathEscape(issuer) + ":" + url.PathEscape(account)

	params := url.Values{}
	params.Add("secret", secretBase32)
	params.Add("issuer", issuer)

	URL.RawQuery = params.Encode()
	fmt.Printf("URL is %s\n", URL.String())

	qrdata, err := qr.Encode(URL.String(), qr.Q)
	if err != nil {
		panic(err)
	}
	b := qrdata.PNG()
	err = ioutil.WriteFile(qrFilename, b, 0600)
	if err != nil {
		panic(err)
	}
	fmt.Printf("QR code is in %s. Please scan it into Google Authenticator app.\n", qrFilename)
}
