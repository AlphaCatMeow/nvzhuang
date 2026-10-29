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
	// 允许 HTTPS 页面在用户明确配置本机后端时完成 Private Network Access 预检。
	// 后端仍应只在可信环境运行；ComfyUI 地址与工作流不能由浏览器请求覆盖。
	if r.Header.Get("Access-Control-Request-Private-Network") == "true" {
		r.Response.Header().Set("Access-Control-Allow-Private-Network", "true")
	}
	// CORSDefault 会直接完成 OPTIONS 预检，因此 PNA 响应头必须先写入。
	r.Response.CORSDefault()
	r.Middleware.Next()
}
