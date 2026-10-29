package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"backend/internal/controller/hello"
	"backend/internal/controller/tryon"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Middleware(corsMiddleware)
				group.Bind(
					hello.NewV1(),
					tryon.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)

// corsMiddleware 跨域中间件，允许前端页面跨域调用
func corsMiddleware(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}
