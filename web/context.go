package web

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Context struct {
	Req *http.Request
	// Resp 如果用户直接使用这个
	// 那么他们就绕开了 RespData 和 RespStatusCode 两个
	// 那么部分 middleware 无法运作
	Resp           http.ResponseWriter
	RespData       []byte //数据，这里主要是给其他middleware读写用的
	RespStatusCode int    //状态码记录

	// Ctx context.Context

	PathParams map[string]string

	queryValues url.Values

	MatchedRoute string

	tplEngine TemplateEngine

	UserValues map[string]any

	// cookieSameSite http.SameSite
}

func (c *Context) Redirect(url string) {
	http.Redirect(c.Resp, c.Req, url, http.StatusFound)
}

func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.RespData = data
	c.RespStatusCode = status
	return nil
}
