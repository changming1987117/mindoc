package routers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/changming1987117/mindoc/conf"
	"github.com/changming1987117/mindoc/models"
	"net/url"
	"regexp"
)

func init() {
	var FilterUser = func(ctx *context.Context) {
		_, ok := ctx.Input.Session(conf.LoginSessionName).(models.Member)

		if !ok {
			if ctx.Input.IsAjax() {
				jsonData := make(map[string]interface{}, 3)

				jsonData["errcode"] = 403
				jsonData["message"] = "请登录后再操作"

				returnJSON, _ := json.Marshal(jsonData)

				ctx.ResponseWriter.Write(returnJSON)
			} else {
				beego.Info("初始访问")
				login_url := beego.AppConfig.String("login_url")
				sys_url := beego.AppConfig.String("sys_url")
				appid := beego.AppConfig.String("appid")
				url := login_url + "?appId=" + appid + "&url=" + url.PathEscape(sys_url+ctx.Request.URL.RequestURI())
				//ctx.Redirect(302, conf.URLFor("AccountController.Login")+"?url="+url.PathEscape(conf.BaseUrl+ctx.Request.URL.RequestURI()))
				beego.Info(url)
				ctx.Redirect(302, url)
			}
		}
	}
	beego.InsertFilter("/manager", beego.BeforeRouter, FilterUser)
	beego.InsertFilter("/manager/*", beego.BeforeRouter, FilterUser)
	beego.InsertFilter("/setting", beego.BeforeRouter, FilterUser)
	beego.InsertFilter("/setting/*", beego.BeforeRouter, FilterUser)
	beego.InsertFilter("/book", beego.BeforeRouter, FilterUser)
	beego.InsertFilter("/book/*", beego.BeforeRouter, FilterUser)
	beego.InsertFilter("/api/*", beego.BeforeRouter, FilterUser)
	beego.InsertFilter("/manage/*", beego.BeforeRouter, FilterUser)

	var FinishRouter = func(ctx *context.Context) {
		ctx.ResponseWriter.Header().Add("MinDoc-Version", conf.VERSION)
		ctx.ResponseWriter.Header().Add("MinDoc-Site", "https://www.iminho.me")
	}

	var StartRouter = func(ctx *context.Context) {
		sessionId := ctx.Input.Cookie(beego.AppConfig.String("sessionname"))
		if sessionId != "" {
			//sessionId必须是数字字母组成，且最小32个字符，最大1024字符
			if ok, err := regexp.MatchString(`^[a-zA-Z0-9]{32,512}$`, sessionId); !ok || err != nil {
				panic("401")
			}
		}
	}
	beego.InsertFilter("/*", beego.BeforeStatic, StartRouter, false)
	beego.InsertFilter("/*", beego.BeforeRouter, FinishRouter, false)
}
