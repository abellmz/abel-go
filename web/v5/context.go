package v5

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req        *http.Request
	Resp       http.ResponseWriter
	PathParams map[string]string

	// 缓存的数据
	cacheQueryValues url.Values
}

type StringValue struct {
	val string
	err error
}

func (s StringValue) String() (string, error) {
	return s.val, s.err
}

func (s StringValue) ToInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	//
	return strconv.ParseInt(s.val, 10, 64)
}

//不能用泛型
//func (s StringValue) To[T any]() (T, error)  {
//
//}

// BindJSON 将字节流 装换为json格式
func (c *Context) BindJSON(val any) error {
	if c.Req.Body == nil {
		return errors.New("web:body 为 nil")
	}
	//不可用，c.Req.Body是io.ReadCloser，而不是字节数组
	//bs ,_ := io.ReadAll(c.Req.Body)
	//json.Unmarshal(bs,val)
	// 一次性Req.Body反序列化赋值decoder
	decoder := json.NewDecoder(c.Req.Body)
	//其中有未知字段就会报错
	//decoder.DisallowUnknownFields()
	//decoder.UseNumber()//将数字类型转为接口的数字类型，否则默认为float64
	//decoder.DisallowUnknownFields()//
	return decoder.Decode(val)
}

// FormValue 获取表单数据
func (c *Context) FormValue(key string) StringValue {
	//  Form和PostForm 都需要ParseForm解析数据 Form可以获取所有表单数据，PostForm只有编码为x-www-form-urlencoded，时才能拿到数据
	//ParseForm不用当心重复解析
	if err := c.Req.ParseForm(); err != nil {
		return StringValue{err: err}
	}
	return StringValue{val: c.Req.FormValue(key)}
}

// QueryValue 获取Query参数，Query和表单对比起来，没有缓存
func (c *Context) QueryValue(key string) StringValue {
	if c.cacheQueryValues == nil {
		c.cacheQueryValues = c.Req.URL.Query()
	}
	vals, ok := c.cacheQueryValues[key]
	if !ok {
		return StringValue{err: errors.New("web:找不到对应的key")}
	}
	return StringValue{val: vals[0]}
}

// PathValue 获取路径参数
func (c *Context) PathValue(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{err: errors.New("web: 找不到这个 key")}
	}
	return StringValue{val: val}
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}

func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

// RespJSON 响应输出json
func (c *Context) RespJSON(code int, val any) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.Resp.WriteHeader(code)
	_, err = c.Resp.Write(bs)
	return err
}

// 返回int64的数据
// func (c *Context) QueryValueAsInt64(key string) (int64, error) {
// 	val, err := c.QueryValue(key)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return strconv.ParseInt(val, 10, 64)
// }
