package errhdl

import (
	"abel-go/web"
	"net/http"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := NewMiddlewareBuilder()
	builder.AddCode(http.StatusNotFound, []byte(`
<html>
    <body>
        <h1>嘿嘿嘿~~~</h1>
    </body>
</html>
`)).AddCode(http.StatusBadRequest, []byte(`
<html>
    <body>
        <h1>请求不对</h1>
    </body>
</html>
`))
	server := web.NewHTTPServer(web.ServerWithMiddleware(builder.Build()))
	server.Start(":8081")
}
