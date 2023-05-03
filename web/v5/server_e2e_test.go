package v5

import (
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, world"))
	})
	s.Get("/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, user"))
	})

	s.Post("/form", func(ctx *Context) {
		err := ctx.Req.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
	})

	s.Start(":8081")
}
