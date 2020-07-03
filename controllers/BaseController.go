package controllers

import (
	"bytes"

	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/changming1987117/mindoc/conf"
	"github.com/changming1987117/mindoc/models"
	"github.com/changming1987117/mindoc/utils"
	"path/filepath"
	"io/ioutil"
	"html/template"
	"net/url"
	"net/http"
	"crypto/tls"
	"fmt"
)

type BaseController struct {
	beego.Controller
	Member                *models.Member
	Option                map[string]string
	EnableAnonymous       bool
	EnableDocumentHistory bool
}

type CookieRemember struct {
	MemberId int
	Account  string
	Time     time.Time
}

/**
 * get_user_info
 */
func (c *BaseController) getUserInfo(ticket string) []byte {
	appid := beego.AppConfig.String("appid")
	appkey := beego.AppConfig.String("sec_key")
	getUserUrl := beego.AppConfig.String("getUserUrl")
	proxyUrl := beego.AppConfig.String("proxy")
	realurl := getUserUrl + "?appid=" + appid + "&appsecret=" + appkey + "&" + ticket
	/*
		1. 代理请求
		2. 跳过https不安全验证
	*/
	proxy, _ := url.Parse(proxyUrl)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5, //超时时间
	}
	resp, err := client.Get(realurl)
	if err != nil {
		fmt.Println("出错了", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	return body

}

// Prepare 预处理.
func (c *BaseController) Prepare() {
	c.Data["SiteName"] = "MinDoc"
	c.Data["Member"] = models.NewMember()
	controller, action := c.GetControllerAndAction()

	c.Data["ActionName"] = action
	c.Data["ControllerName"] = controller

	c.EnableAnonymous = false
	c.EnableDocumentHistory = false
	conf.BaseUrl = c.BaseUrl()
	c.Data["BaseUrl"] = c.BaseUrl()

	if options, err := models.NewOption().All(); err == nil {
		c.Option = make(map[string]string, len(options))
		for _, item := range options {
			c.Data[item.OptionName] = item.OptionValue
			c.Option[item.OptionName] = item.OptionValue
		}
		c.EnableAnonymous = strings.EqualFold(c.Option["ENABLE_ANONYMOUS"], "true")
		c.EnableDocumentHistory = strings.EqualFold(c.Option["ENABLE_DOCUMENT_HISTORY"], "true")
	}
	c.Data["HighlightStyle"] = beego.AppConfig.DefaultString("highlight_style", "github")

	if b, err := ioutil.ReadFile(filepath.Join(beego.BConfig.WebConfig.ViewsPath, "widgets", "scripts.tpl")); err == nil {
		c.Data["Scripts"] = template.HTML(string(b))
	}
	if member, ok := c.GetSession(conf.LoginSessionName).(models.Member); ok && member.MemberId > 0 {
		c.Member = &member
		c.Data["Member"] = c.Member
	}
	var remember CookieRemember
	// 如果 Cookie 中存在登录信息
	if cookie, ok := c.GetSecureCookie(conf.GetAppKey(), "login"); ok {
		if err := utils.Decode(cookie, &remember); err == nil {
			if member, err := models.NewMember().Find(remember.MemberId); err == nil {
				c.SetMember(*member)
				c.Member = member
				c.Data["Member"] = member
			}
		}
	}

}

//登录
func (c * BaseController) Logged(){
	u := c.GetString("url")
	if member, ok := c.GetSession(conf.LoginSessionName).(models.Member); ok && member.MemberId > 0 {
		if u == "" {

			u = conf.URLFor("DocumentController.Index", ":key", "bumenzichanku")
		}
		c.Redirect(u, 302)
	}
	var remember CookieRemember
	var account AccountController
	// 如果 Cookie 中存在登录信息
	if cookie, ok := c.GetSecureCookie(conf.GetAppKey(), "login"); ok {
		if err := utils.Decode(cookie, &remember); err == nil {
			if member, err := models.NewMember().Find(remember.MemberId); err == nil {
				c.SetMember(*member)
				c.Member = member
				c.Data["Member"] = member
				account.LoggedIn(false)
				u = conf.URLFor("DocumentController.Index", ":key", "bumenzichanku")
				c.Redirect(u, 302)
			}
		}
	}
	//如果没有开启匿名访问，则跳转到登录页面
	if c.Member == nil {

		loginUrl := beego.AppConfig.String("loginUrl")
		sysUrl := beego.AppConfig.String("sysUrl")
		appid := beego.AppConfig.String("appid")
		ticket := beego.AppConfig.String("ticket")

		u := c.Ctx.Request.URL.RequestURI()
		beego.Info(u)
		if strings.Contains(u, ticket) {
			ticketLists := strings.Split(u, "?")
			realticket := strings.Split(ticketLists[1], "&")[0]
			returnUrl := ticketLists[0]
			resp := c.getUserInfo(realticket)
			var res Result
			json.Unmarshal(resp, &res)
			chineseName := res.Data["chineseName"]
			userName := res.Data["userName"]
			email := res.Data["email"]
			member := models.NewMember()
			member, err := member.FindByAccount(userName)
			if err == nil && member.MemberId > 0 {
			} else {
				member.Account = userName
				member.Password = email
				member.Role = 2
				member.Avatar = conf.GetDefaultAvatar()
				member.CreateAt = 1
				member.Email = email
				member.RealName = chineseName
				err :=member.Add()
				if err != nil {
					beego.Info(err)
				}
				m := models.NewRelationship()
				m.BookId = 2
				m.MemberId = member.MemberId
				m.RoleId = 3
				m.Insert()
			}
			loginMem, err := member.Login(userName, email)
			if err == nil {
				loginMem.LastLoginTime = time.Now()
				loginMem.Update()
				c.SetMember(*loginMem)
				returnUrl = sysUrl + returnUrl
				beego.Info(returnUrl)
				//returnUrl = conf.URLFor("DocumentController.Index", ":key", "bumenzichanku")
				c.Redirect(returnUrl, 302)
				return
			}
		}
		redirecturl := loginUrl + "?appId=" + appid + "&url=" + url.PathEscape(sysUrl+u)
		c.Redirect(redirecturl, 302)
	}
}
//判断用户是否登录.
func (c *BaseController)isUserLoggedIn() bool {
	return c.Member != nil && c.Member.MemberId > 0
}

// SetMember 获取或设置当前登录用户信息,如果 MemberId 小于 0 则标识删除 Session
func (c *BaseController) SetMember(member models.Member) {

	if member.MemberId <= 0 {
		c.DelSession(conf.LoginSessionName)
		c.DelSession("uid")
		c.DestroySession()
	} else {
		c.SetSession(conf.LoginSessionName, member)
		c.SetSession("uid", member.MemberId)
	}
}

// JsonResult 响应 json 结果
func (c *BaseController) JsonResult(errCode int, errMsg string, data ...interface{}) {
	jsonData := make(map[string]interface{}, 3)

	jsonData["errcode"] = errCode
	jsonData["message"] = errMsg

	if len(data) > 0 && data[0] != nil {
		jsonData["data"] = data[0]
	}

	returnJSON, err := json.Marshal(jsonData)

	if err != nil {
		beego.Error(err)
	}

	c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache, no-store")
	io.WriteString(c.Ctx.ResponseWriter, string(returnJSON))

	c.StopRun()
}

//如果错误不为空，则响应错误信息到浏览器.
func (c *BaseController) CheckJsonError(code int,err error) {

	if err == nil {
		return
	}
	jsonData := make(map[string]interface{}, 3)

	jsonData["errcode"] = code
	jsonData["message"] = err.Error()

	returnJSON, err := json.Marshal(jsonData)

	if err != nil {
		beego.Error(err)
	}

	c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache, no-store")
	io.WriteString(c.Ctx.ResponseWriter, string(returnJSON))

	c.StopRun()
}

// ExecuteViewPathTemplate 执行指定的模板并返回执行结果.
func (c *BaseController) ExecuteViewPathTemplate(tplName string, data interface{}) (string, error) {
	var buf bytes.Buffer

	viewPath := c.ViewPath

	if c.ViewPath == "" {
		viewPath = beego.BConfig.WebConfig.ViewsPath

	}

	if err := beego.ExecuteViewPathTemplate(&buf, tplName, viewPath, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c *BaseController) BaseUrl() string {
	baseUrl := beego.AppConfig.DefaultString("baseurl", "")
	if baseUrl != "" {
		if strings.HasSuffix(baseUrl, "/") {
			baseUrl = strings.TrimSuffix(baseUrl, "/")
		}
	} else {
		baseUrl = c.Ctx.Input.Scheme() + "://" + c.Ctx.Request.Host
	}
	return baseUrl
}

//显示错误信息页面.
func (c *BaseController) ShowErrorPage(errCode int, errMsg string) {
	c.TplName = "errors/error.tpl"

	c.Data["ErrorMessage"] = errMsg
	c.Data["ErrorCode"] = errCode

	var buf bytes.Buffer

	if err := beego.ExecuteViewPathTemplate(&buf, "errors/error.tpl", beego.BConfig.WebConfig.ViewsPath, map[string]interface{}{"ErrorMessage": errMsg, "ErrorCode": errCode, "BaseUrl": conf.BaseUrl}); err != nil {
		c.Abort("500")
	}
	if errCode >= 200 && errCode <= 510 {
		c.CustomAbort(errCode, buf.String())
	} else {
		c.CustomAbort(500, buf.String())
	}
}


func (c *BaseController) CheckErrorResult(code int,err error) {
	if err != nil {
		c.ShowErrorPage(code, err.Error())
	}
}
