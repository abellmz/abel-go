package v2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func Test_router_AddRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}
	// 结构体中是map,初始化 + 赋值一体化
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {path: "/", children: map[string]*node{
				"user": {path: "user", children: map[string]*node{
					"home": {path: "home", handler: mockHandler},
				}, handler: mockHandler},
				"order": {path: "order", children: map[string]*node{
					"detail": {path: "detail", handler: mockHandler},
				}},
			}, handler: mockHandler},
			http.MethodPost: {path: "/", children: map[string]*node{
				"order": {path: "order", children: map[string]*node{
					"create": {path: "create", handler: mockHandler},
				}},
				"login": {path: "login", handler: mockHandler},
			}},
		},
	}
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	//非法用例
	r = newRouter()

	//空字符串
	assert.PanicsWithValue(t, "web: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	//前导没有 /
	assert.PanicsWithValue(t, "web: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})

	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})
}

/*
*
equal router.tree的比较
r wantRouter,外来路由
y 写好的，系统路由
*/
func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		//比较是否含有对应路由树
		//y中没有该路由树，则报警
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标 router 里面没有方法 %s 的路由树\", k"), false
		}
		// 比较路由树是否相等
		str, ok := v.equal(yv)
		if !ok {
			return k + "-" + str, ok
		}
	}
	return "", true
}

/*
*
equal 对比树的节点
*/
func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标结点为 nil", false
	}
	if n.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}
	// 返回一个reflect.Value 动态值
	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}
	if len(n.children) == 0 {
		return "", true
	}

	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
	}

	mockHandler := func(ctx *Context) {}
	testCases := []struct {
		name     string
		method   string
		path     string
		found    bool
		wantNode *node
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			wantNode: &node{
				path:    "/",
				handler: mockHandler,
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			wantNode: &node{
				path:    "user",
				handler: mockHandler,
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			wantNode: &node{
				path: "order",
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			wantNode: &node{
				path:    "create",
				handler: mockHandler,
			},
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			wantVal := reflect.ValueOf(tc.wantNode.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}
