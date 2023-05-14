package prometheus

import (
	"abel-go/web"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

// Summary 是在客户端计算后，推送到prometheus服务器，因此会占用客户端资源
// histogram则是将数据放在服务端计算
// vector向量

func (m MiddlewareBuilder) Build() web.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      m.Name,
		Subsystem: m.Subsystem, //横向：http或卡夫卡 竖向分：业务user/order等
		Namespace: m.Name,      //可以是app的名称
		Help:      m.Help,
		Objectives: map[float64]float64{
			0.5:   0.01, //key百分比，value误差，误差越小需要越多的cpu资源， 此为中位数 0.5+_0.01
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001, //99线，99%的响应时间
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"}) //动态label
	prometheus.MustRegister(vector) //注册失败则panic
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			startTime := time.Now()
			defer func() {
				duration := time.Now().Sub(startTime).Milliseconds()
				pattern := ctx.MatchedRoute
				if pattern == "" {
					pattern = "unknown"
				}
				vector.WithLabelValues(pattern, ctx.Req.Method,
					strconv.Itoa(ctx.RespStatusCode)).Observe(float64(duration))
			}()
			next(ctx)
		}
	}
}
