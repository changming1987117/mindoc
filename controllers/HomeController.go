package controllers

import (
	"math"
	"github.com/astaxie/beego"
	"github.com/changming1987117/mindoc/models"
	"github.com/changming1987117/mindoc/utils/pagination"
	"github.com/changming1987117/mindoc/conf"
	"net/url"
	"net/http"
	"crypto/tls"
	"time"
	"fmt"
	"io/ioutil"
	"strings"
	"encoding/json"
)

type HomeController struct {
	BaseController
}

type Result struct {
	Code    int
	Message string
	Data    map[string]string
}

/**
 * get_user_info
 */
func (c *HomeController) getUserInfo(ticket string) []byte {
	appid := beego.AppConfig.String("appid")
	appkey := beego.AppConfig.String("sec_key")
	getUserUrl := beego.AppConfig.String("getUserUrl")
	proxyUrl := beego.AppConfig.String("proxy")
	realurl := getUserUrl + "?appid=" + appid + "&appsecret=" + appkey + "&" + ticket
	beego.Info(realurl)
	/*
		1. 代理请求
		2. 跳过https不安全验证
	*/
	proxy, _ := url.Parse(proxyUrl)
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
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

func (c *HomeController) Prepare() {
	c.BaseController.Prepare()
	//如果没有开启匿名访问，则跳转到登录页面
	if !c.EnableAnonymous && c.Member == nil {
		beego.Info("初始访问")
		loginUrl := beego.AppConfig.String("loginUrl")
		sysUrl := beego.AppConfig.String("sysUrl")
		appid := beego.AppConfig.String("appid")
		ticket := beego.AppConfig.String("ticket")
		u := c.Ctx.Request.URL.RequestURI()
		beego.Info(u)
		if strings.Contains(u, ticket) {
			beego.Info(ticket)
			ticketLists := strings.Split(u, "?")
			realticket := strings.Split(ticketLists[1], "&")[0]
			returnUrl := ticketLists[0]
			beego.Info(returnUrl)
			beego.Info(realticket)
			resp := c.getUserInfo(realticket)
			var res Result
			json.Unmarshal(resp, &res)
			beego.Info(res.Data)
			chineseName := res.Data["chineseName"]
			userName := res.Data["userName"]
			email := res.Data["email"]
			member := models.NewMember()
			m, err := member.FindByAccount(userName)
			if err == nil && member.MemberId > 0 {
				m.Password = email
				m.Email = email
				m.RealName = chineseName
			} else{
			member.Account = userName
			member.Password = email
			member.Role = 2
			member.Avatar = conf.GetDefaultAvatar()
			member.CreateAt = 0
			member.Email = email
			member.RealName = chineseName
			}
			loginMem, err := member.Login(userName, email)
			if err == nil {
				loginMem.LastLoginTime = time.Now()
				loginMem.Update()
				c.SetMember(*loginMem)
				redirecturl := url.PathEscape(sysUrl + u)
				beego.Info(redirecturl)
				c.Redirect(redirecturl, 302)
			}
		}
		redirecturl := loginUrl + "?appId=" + appid + "&url=" + url.PathEscape(sysUrl+u)
		beego.Info(redirecturl)
		c.Redirect(redirecturl, 302)
	}

	//c.Redirect(conf.URLFor("AccountController.Login")+"?url="+url.PathEscape(conf.BaseUrl+c.Ctx.Request.URL.RequestURI()), 302)
}


func (c *HomeController) Index() {
	c.Prepare()
	c.TplName = "home/index.tpl"

	pageIndex, _ := c.GetInt("page", 1)
	pageSize := 18

	memberId := 0

	if c.Member != nil {
		memberId = c.Member.MemberId
	}
	books, totalCount, err := models.NewBook().FindForHomeToPager(pageIndex, pageSize, memberId)

	if err != nil {
		beego.Error(err)
		c.Abort("500")
	}
	if totalCount > 0 {
		pager := pagination.NewPagination(c.Ctx.Request, totalCount, pageSize, c.BaseUrl())
		c.Data["PageHtml"] = pager.HtmlPages()
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))

	c.Data["Lists"] = books
}
