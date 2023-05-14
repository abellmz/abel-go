package web

import (
	"fmt"
	"net"
	"net/http"
)

type HandleFunc func(ctx *Context)

var _ Server = &HTTPServer{}

type Server interface {
	http.Handler
	Start(add string) error

	// AddRoute 路由注册功能
	// method 是 HTTP 方法
	// path 是路由
	// handleFunc 是你的业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc, ms ...Middleware)

	// 这种允许注册多个，没有必要提供
	// 让用户自己去管
	// AddRoute1(method string, path string, handles ...HandleFunc)
}

type HTTPServer struct {
	// addr string 创建的时候传递，而不是 Start 接收。这个都是可以的
	router

	mdls      []Middleware
	log       func(msg string, args ...any)
	tplEngine TemplateEngine
}

type HTTPServerOption func(server *HTTPServer)

func NewHTTPServerV1(mdls ...Middleware) *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
		mdls:   mdls,
	}
}

// 第一个问题：相对路径还是绝对路径？
// 你的配置文件格式，json, yaml, xml
// func NewHTTPServerV2(cfgFilePath string) *HTTPServer {
// 你在这里加载配置，解析，然后初始化 HTTPServer
// }

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: newRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithTemplateEngine(tplEngine TemplateEngine) HTTPServerOption {
	return func(server *HTTPServer) {
		server.tplEngine = tplEngine
	}
}

func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.mdls = mdls
	}
}

// ServeHTTP 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 你的框架代码就在这里
	ctx := &Context{
		Req:       request,
		Resp:      writer,
		tplEngine: h.tplEngine,
	}
	// 最后一个是这个
	root := h.serve

	// 然后这里就是利用最后一个不断往前回溯组装链条
	// 从后往前
	// 把后一个作为前一个的 next 构造好链条
	for i := len(h.mdls) - 1; i >= 0; i-- {
		root = h.mdls[i](root)
	}

	// 这里执行的时候，就是从前往后了

	// 这里，最后一个步骤，就是把 RespData 和 RespStatusCode 刷新到响应里面
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

// flashResp 刷新响应的状态码和数据
func (h *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		h.log("写入响应失败 %v", err)
	}
}

func (h *HTTPServer) serve(ctx *Context) {
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.n.handler == nil {
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("NOT FOUND")
		return
	}
	ctx.PathParams = info.pathParams
	ctx.MatchedRoute = info.n.route
	// before execute
	info.n.handler(ctx)
	// after execute
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}

// Start 启动服务器，用户指定端口
// 这种就是编程接口
func (h *HTTPServer) Start(addr string) error {
	// 也可以自己创建 Server
	// http.Server{}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 在这里，可以让用户注册所谓的 after start 回调
	// 比如说往你的 admin 注册一下自己这个实例
	// 在这里执行一些你业务所需的前置条件

	return http.Serve(l, h)
}
