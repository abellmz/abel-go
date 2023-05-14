package web

import "context"

type TemplateEngine interface {
	// Render 渲染页面
	// tplName 模板的名字，按名索引
	// data 渲染页面用的数据
	Render(ctx context.Context, tmlName string, data any) ([]byte, error)
}
